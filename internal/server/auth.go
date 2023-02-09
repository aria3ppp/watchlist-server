package server

import (
	"net/http"

	"github.com/aria3ppp/watchlist-server/internal/auth"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

const contextKey string = "token_payload"

func (s *Server) getUserPayload(
	c echo.Context,
) (*auth.Payload, *echo.HTTPError) {
	payload, ok := c.Get(contextKey).(*auth.Payload)
	if !ok {
		s.logger.Error(
			"server.getUserPayload: context key not set",
			zap.String("key", contextKey),
		)
		return nil, echo.NewHTTPError(http.StatusInternalServerError)
	}
	return payload, nil
}
