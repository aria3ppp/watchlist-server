package server

import (
	"net/http"

	"github.com/aria3ppp/watchlist-server/internal/app"
	"github.com/aria3ppp/watchlist-server/internal/dto"
	"github.com/aria3ppp/watchlist-server/internal/server/request"
	"github.com/aria3ppp/watchlist-server/internal/server/response"
	"github.com/aria3ppp/watchlist-server/internal/token"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// GET /v1/authorized/user/:id/
func (s *Server) HandleUserGet(c echo.Context) error {
	// bind & validate id param
	var param request.IDPathParam
	if httpError := s.bindPath(c, &param); httpError != nil {
		return httpError
	}

	// Read user
	user, err := s.app.UserGet(c.Request().Context(), param.ID)
	if err != nil {
		if err == app.ErrNotFound {
			s.logger.Info(
				"server.HandleUserGet: user not found",
				zap.Int("id", param.ID),
			)
			return echo.NewHTTPError(
				http.StatusNotFound,
				response.Error(response.StatusNotFound),
			)
		}

		s.logger.Error(
			"server.HandleUserGet: internal server error", zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			response.Error(response.StatusInternalServerError),
		)
	}

	return c.JSON(http.StatusOK, response.OK(user))
}

//------------------------------------------------------------------------------

// POST /v1/user/
func (s *Server) HandleUserCreate(c echo.Context) error {
	// bind & validate request
	var req dto.UserCreateRequest
	if httpError := s.bindBody(c, &req); httpError != nil {
		return httpError
	}

	// Create user
	userID, err := s.app.UserCreate(c.Request().Context(), &req)
	if err != nil {
		if err == app.ErrEmailAlreadyUsed {
			s.logger.Info(
				"server.HandleUserCreate: request email already used",
			)
			return echo.NewHTTPError(
				http.StatusBadRequest,
				response.Error(response.StatusEmailAlreadyUsed),
			)
		}

		s.logger.Error(
			"server.HandleUserCreate: internal server error", zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			response.Error(response.StatusInternalServerError),
		)

	}

	return c.JSON(http.StatusOK, response.OK(userID))
}

//------------------------------------------------------------------------------

// POST /v1/user/login/
func (s *Server) HandleUserLogin(c echo.Context) error {
	// bind & validate request
	var req dto.UserLoginRequest
	if httpError := s.bindBody(c, &req); httpError != nil {
		return httpError
	}

	// login
	accessToken, refreshToken, err := s.app.UserLogin(
		c.Request().Context(),
		&req,
	)
	if err != nil {
		if err == app.ErrNotFound {
			s.logger.Info(
				"server.HandleLoginUser: request email not found",
				zap.String("email", req.Email),
			)
			return echo.NewHTTPError(
				http.StatusBadRequest,
				response.Error(response.StatusEmailNotFound),
			)
		}

		if err == app.ErrIncorrectPassword {
			s.logger.Info(
				"server.HandleLoginUser: request password not matched",
			)
			return echo.NewHTTPError(
				http.StatusBadRequest,
				response.Error(response.StatusIncorrectPassword),
			)
		}

		s.logger.Error(
			"server.HandleLoginUser: internal server error",
			zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			response.Error(response.StatusInternalServerError),
		)

	}

	// return token
	return c.JSON(
		http.StatusOK,
		response.OK(response.TokenPair{
			Access:  accessToken,
			Refresh: refreshToken,
		}),
	)
}

//------------------------------------------------------------------------------

// POST /v1/user/refresh_token/
func (s *Server) HandleUserRefreshToken(c echo.Context) error {
	// parse authorization header
	auth := c.Request().Header.Get(echo.HeaderAuthorization)
	refreshToken := token.ExtractTokenFromAuth(auth)
	if refreshToken == "" {
		s.logger.Info("server.HandleRefreshToken: token malformed or missing")
		return echo.NewHTTPError(
			http.StatusBadRequest,
			response.Error(response.StatusTokenMissingOrMalformed),
		)
	}

	// refresh token
	newAccessToken, err := s.app.UserRefreshToken(
		c.Request().Context(),
		refreshToken,
	)
	if err != nil {
		if err == app.ErrTokenInvalid {
			s.logger.Info(
				"server.HandleRefreshToken: request refresh token not valid",
				zap.String("token", refreshToken),
			)
			return echo.NewHTTPError(
				http.StatusBadRequest,
				response.Error(response.StatusTokenInvalid),
			)
		}

		s.logger.Error(
			"server.HandleRefreshToken: internal server error", zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			response.Error(response.StatusInternalServerError),
		)

	}

	// return token
	return c.JSON(http.StatusOK, response.OK(newAccessToken))
}

//------------------------------------------------------------------------------

// PATCH /v1/authorized/user/
func (s *Server) HandleUserUpdate(c echo.Context) error {
	// bind & validate request
	var req dto.UserUpdateRequest
	if httpError := s.bindBody(c, &req); httpError != nil {
		return httpError
	}

	payload, httpError := s.getUserPayload(c)
	if httpError != nil {
		return httpError
	}

	// Update user
	err := s.app.UserUpdate(c.Request().Context(), payload.UserID, &req)
	if err != nil {
		if err == app.ErrNotFound {
			s.logger.Info(
				"server.HandleUserUpdate: user not found",
				zap.Int("id", payload.UserID),
			)
			return echo.NewHTTPError(
				http.StatusNotFound,
				response.Error(response.StatusNotFound),
			)
		}

		s.logger.Error(
			"server.HandleUserUpdate: internal server error", zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			response.Error(response.StatusInternalServerError),
		)

	}

	return c.JSON(http.StatusOK, response.OK(nil))
}

//------------------------------------------------------------------------------

// PUT /v1/authorized/user/email/
func (s *Server) HandleUserEmailUpdate(c echo.Context) error {
	// bind & validate request
	var req dto.UserEmailUpdateRequest
	if httpError := s.bindBody(c, &req); httpError != nil {
		return httpError
	}

	payload, httpError := s.getUserPayload(c)
	if httpError != nil {
		return httpError
	}

	// Change user email
	err := s.app.UserEmailUpdate(c.Request().Context(), payload.UserID, &req)
	if err != nil {
		if err == app.ErrNotFound {
			s.logger.Info(
				"server.HandleUserEmailUpdate: user not found",
				zap.Int("id", payload.UserID),
			)
			return echo.NewHTTPError(
				http.StatusNotFound,
				response.Error(response.StatusNotFound),
			)
		}

		s.logger.Error(
			"server.HandleUserEmailUpdate: internal server error",
			zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			response.Error(response.StatusInternalServerError),
		)

	}

	return c.JSON(http.StatusOK, response.OK(nil))
}

//------------------------------------------------------------------------------

// PUT /v1/authorized/user/password/
func (s *Server) HandleUserPasswordUpdate(c echo.Context) error {
	// bind & validate request
	var req dto.UserPasswordUpdateRequest
	if httpError := s.bindBody(c, &req); httpError != nil {
		return httpError
	}

	payload, httpError := s.getUserPayload(c)
	if httpError != nil {
		return httpError
	}

	// Change user password
	err := s.app.UserPasswordUpdate(
		c.Request().Context(),
		payload.UserID,
		&req,
	)
	if err != nil {
		if err == app.ErrSameNewPassword {
			s.logger.Info(
				"server.HandleUserPasswordUpdate: same new password",
			)
			return echo.NewHTTPError(
				http.StatusBadRequest,
				response.Error(response.StatusSameNewPassword),
			)
		}

		if err == app.ErrNotFound {
			s.logger.Info(
				"server.HandleUserPasswordUpdate: user not found",
				zap.Int("id", payload.UserID),
			)
			return echo.NewHTTPError(
				http.StatusNotFound,
				response.Error(response.StatusNotFound),
			)
		}

		if err == app.ErrIncorrectPassword {
			s.logger.Info(
				"server.HandleUserPasswordUpdate: request password not matched",
			)
			return echo.NewHTTPError(
				http.StatusBadRequest,
				response.Error(response.StatusIncorrectPassword),
			)
		}

		s.logger.Error(
			"server.HandleUserPasswordUpdate: internal server error",
			zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			response.Error(response.StatusInternalServerError),
		)

	}

	return c.JSON(http.StatusOK, response.OK(nil))
}

//------------------------------------------------------------------------------

// DELETE /v1/authorized/user/
func (s *Server) HandleUserDelete(c echo.Context) error {
	// bind & validate request
	var req dto.UserDeleteRequest
	if httpError := s.bindBody(c, &req); httpError != nil {
		return httpError
	}

	payload, httpError := s.getUserPayload(c)
	if httpError != nil {
		return httpError
	}

	// Delete user
	err := s.app.UserDelete(c.Request().Context(), payload.UserID, &req)
	if err != nil {
		if err == app.ErrNotFound {
			s.logger.Info(
				"server.HandleUserDelete: user not found",
				zap.Int("id", payload.UserID),
			)
			return echo.NewHTTPError(
				http.StatusNotFound,
				response.Error(response.StatusNotFound),
			)
		}

		if err == app.ErrIncorrectPassword {
			s.logger.Info(
				"server.HandleUserDelete: request password not matched",
			)
			return echo.NewHTTPError(
				http.StatusBadRequest,
				response.Error(response.StatusIncorrectPassword),
			)
		}

		s.logger.Error(
			"server.HandleUserDelete: internal server error", zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			response.Error(response.StatusInternalServerError),
		)

	}

	return c.JSON(http.StatusOK, response.OK(nil))
}
