package server

import (
	"net/http"

	"github.com/aria3ppp/watchlist-server/internal/server/response"
	token_service "github.com/aria3ppp/watchlist-server/internal/token"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

const PayloadKey = "user_payload"

var (
	ErrTokenMissingOrMalformed error = echo.NewHTTPError(
		http.StatusUnauthorized,
		response.Error(response.StatusTokenMissingOrMalformed),
	)
	ErrTokenInvalid error = echo.NewHTTPError(
		http.StatusUnauthorized,
		response.Error(response.StatusTokenInvalid),
	)
)

func FetchUserPayload(c echo.Context) *token_service.Payload {
	payload, ok := c.Get(PayloadKey).(*token_service.Payload)
	if ok {
		return payload
	}
	return nil
}

func (s *Server) AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// extract token
		auth := c.Request().Header.Get(echo.HeaderAuthorization)
		token := token_service.ExtractTokenFromAuth(auth)
		if token == "" {
			s.logger.Info(
				"server.AuthMiddleware: token missing/malformed",
				zap.String(echo.HeaderAuthorization, auth),
			)
			return ErrTokenMissingOrMalformed
		}

		// validate token
		payload, err := s.tokenService.ValidateToken(token)
		if err != nil {
			if err == token_service.ErrInvalidToken {
				s.logger.Info(
					"server.AuthMiddleware: invalid token",
					zap.String("token", token),
				)
				return ErrTokenInvalid
			}

			s.logger.Error(
				"server.JWTMiddleware: internal server error", zap.Error(err),
			)
			return echo.NewHTTPError(
				http.StatusInternalServerError,
				response.Error(response.StatusInternalServerError),
			)
		}

		// set payload in context
		c.Set(PayloadKey, payload)
		return next(c)
	}
}
