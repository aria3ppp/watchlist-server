package server

import (
	"net/http"

	"github.com/aria3ppp/watchlist-server/internal/app"
	"github.com/aria3ppp/watchlist-server/internal/config"
	"github.com/aria3ppp/watchlist-server/internal/dto"
	"github.com/aria3ppp/watchlist-server/internal/server/request"
	"github.com/aria3ppp/watchlist-server/internal/server/response"
	"github.com/aria3ppp/watchlist-server/internal/storage"
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
			return echo.NewHTTPError(http.StatusNotFound)
		}

		s.logger.Error(
			"server.HandleUserGet: internal server error", zap.Error(err),
		)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, user)
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
		if err == app.ErrUsedEmail {
			s.logger.Info(
				"server.HandleUserCreate: request email already used",
			)
			return echo.NewHTTPError(http.StatusConflict)
		}

		s.logger.Error(
			"server.HandleUserCreate: internal server error", zap.Error(err),
		)
		return echo.NewHTTPError(http.StatusInternalServerError)

	}

	return c.JSON(http.StatusOK, response.ID(userID))
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
	resp, err := s.app.UserLogin(c.Request().Context(), &req)
	if err != nil {
		if err == app.ErrNotFound {
			s.logger.Info(
				"server.HandleLoginUser: request email not found",
				zap.String("email", req.Email),
			)
			return echo.NewHTTPError(http.StatusNotFound)
		}

		if err == app.ErrIncorrectPassword {
			s.logger.Info(
				"server.HandleLoginUser: request password not matched",
			)
			return echo.NewHTTPError(http.StatusUnauthorized)
		}

		s.logger.Error(
			"server.HandleLoginUser: internal server error",
			zap.Error(err),
		)
		return echo.NewHTTPError(http.StatusInternalServerError)

	}

	// return token
	return c.JSON(http.StatusOK, resp)
}

//------------------------------------------------------------------------------

// POST /v1/user/:id/logout/
func (s *Server) HandleUserLogout(c echo.Context) error {
	// bind & validate id param
	var param request.IDPathParam
	if httpError := s.bindPath(c, &param); httpError != nil {
		return httpError
	}

	// parse & validate token
	var req request.TokenBody
	if httpError := s.bindBody(c, &req); httpError != nil {
		return httpError
	}

	// refresh token
	err := s.app.UserLogout(c.Request().Context(), param.ID, req.Token)
	if err != nil {
		if err == app.ErrNotFound {
			s.logger.Info(
				"server.HandleUserLogout: user refresh token not found",
				zap.Int("user id", param.ID),
				zap.String("token", req.Token),
			)
			return echo.NewHTTPError(http.StatusNotFound)
		}

		s.logger.Error(
			"server.HandleUserLogout: internal server error", zap.Error(err),
		)
		return echo.NewHTTPError(http.StatusInternalServerError)

	}

	// return token
	return c.NoContent(http.StatusOK)
}

//------------------------------------------------------------------------------

// POST /v1/user/:id/refresh_token/
func (s *Server) HandleUserRefreshToken(c echo.Context) error {
	// bind & validate id param
	var param request.IDPathParam
	if httpError := s.bindPath(c, &param); httpError != nil {
		return httpError
	}

	// parse & validate token
	var req request.TokenBody
	if httpError := s.bindBody(c, &req); httpError != nil {
		return httpError
	}

	// refresh token
	resp, err := s.app.UserRefreshToken(
		c.Request().Context(),
		param.ID,
		req.Token,
	)
	if err != nil {
		if err == app.ErrNotFound {
			s.logger.Info(
				"server.HandleRefreshToken: user refresh token not found",
				zap.Int("user id", param.ID),
				zap.String("token", req.Token),
			)
			return echo.NewHTTPError(http.StatusNotFound)
		}

		s.logger.Error(
			"server.HandleRefreshToken: internal server error", zap.Error(err),
		)
		return echo.NewHTTPError(http.StatusInternalServerError)

	}

	// return token
	return c.JSON(http.StatusOK, resp)
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
			return echo.NewHTTPError(http.StatusNotFound)
		}

		s.logger.Error(
			"server.HandleUserUpdate: internal server error", zap.Error(err),
		)
		return echo.NewHTTPError(http.StatusInternalServerError)

	}

	return c.NoContent(http.StatusOK)
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
			return echo.NewHTTPError(http.StatusNotFound)
		}

		s.logger.Error(
			"server.HandleUserEmailUpdate: internal server error",
			zap.Error(err),
		)
		return echo.NewHTTPError(http.StatusInternalServerError)

	}

	return c.NoContent(http.StatusOK)
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
		if err == app.ErrSamePassword {
			s.logger.Info("server.HandleUserPasswordUpdate: same password")
			return echo.NewHTTPError(http.StatusBadRequest, "same password")
		}

		if err == app.ErrNotFound {
			s.logger.Info(
				"server.HandleUserPasswordUpdate: user not found",
				zap.Int("id", payload.UserID),
			)
			return echo.NewHTTPError(http.StatusNotFound)
		}

		if err == app.ErrIncorrectPassword {
			s.logger.Info("server.HandleUserPasswordUpdate: incorrect password")
			return echo.NewHTTPError(
				http.StatusUnauthorized,
				"incorrect password",
			)
		}

		s.logger.Error(
			"server.HandleUserPasswordUpdate: internal server error",
			zap.Error(err),
		)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
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
			return echo.NewHTTPError(http.StatusNotFound)
		}

		if err == app.ErrIncorrectPassword {
			s.logger.Info("server.HandleUserDelete: incorrect password")
			return echo.NewHTTPError(
				http.StatusUnauthorized,
				"incorrect password",
			)
		}

		s.logger.Error(
			"server.HandleUserDelete: internal server error", zap.Error(err),
		)
		return echo.NewHTTPError(http.StatusInternalServerError)

	}

	return c.NoContent(http.StatusOK)
}

//------------------------------------------------------------------------------

// PUT /v1/authorized/user/avatar/
func (s *Server) HandleUserPutAvatar(c echo.Context) error {
	var (
		filename = config.Config.MinIO.Filename.User
		bucket   = config.Config.MinIO.Bucket.Image.Name
		category = config.Config.MinIO.Category.User
	)

	// get form file
	file, fileHeader, httpError := s.getFormFile(c, filename)
	if httpError != nil {
		return httpError
	}
	defer file.Close()

	// detect content type
	contentType, httpError := s.ensureSupportedFileType(
		file,
		config.Config.MinIO.Bucket.Image.SupportedTypes,
	)
	if httpError != nil {
		return httpError
	}

	payload, httpError := s.getUserPayload(c)
	if httpError != nil {
		return httpError
	}

	// put avatar
	uri, err := s.app.UserPutAvatar(
		c.Request().Context(),
		payload.UserID,
		file,
		&storage.PutOptions{
			Bucket:      bucket,
			Category:    category,
			CategoryID:  payload.UserID,
			Filename:    filename,
			ContentType: contentType,
			Size:        fileHeader.Size,
		},
	)
	if err != nil {
		if err == app.ErrNotFound {
			s.logger.Error(
				"server.HandleUserPutAvatar: user not found",
				zap.Int("id", payload.UserID),
			)
			return echo.NewHTTPError(http.StatusNotFound)
		}

		s.logger.Error(
			"server.HandleUserPutAvatar: failed putting avatar",
			zap.String("bucket", bucket),
			zap.Error(err),
		)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, response.URI(uri))
}
