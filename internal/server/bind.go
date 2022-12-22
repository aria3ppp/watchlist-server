package server

import (
	"net/http"

	"github.com/aria3ppp/watchlist-server/internal/server/response"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func (s *Server) bindBody(
	c echo.Context,
	v validation.Validatable,
) *echo.HTTPError {
	return s.bind(
		defaultBinder.BindBody,
		response.StatusInvalidRequest,
		c,
		v,
	)
}

func (s *Server) bindQuery(
	c echo.Context,
	v validation.Validatable,
) *echo.HTTPError {
	return s.bind(
		defaultBinder.BindQueryParams,
		response.StatusInvalidURLParameter,
		c,
		v,
	)
}

func (s *Server) bindPath(
	c echo.Context,
	v validation.Validatable,
) *echo.HTTPError {
	return s.bind(
		defaultBinder.BindPathParams,
		response.StatusInvalidURLParameter,
		c,
		v,
	)
}

func (s *Server) bindHeaders(
	c echo.Context,
	v validation.Validatable,
) *echo.HTTPError {
	return s.bind(
		defaultBinder.BindHeaders,
		response.StatusInvalidRequest,
		c,
		v,
	)
}

func (s *Server) bind(
	binder func(echo.Context, any) error,
	failureStatus response.Status,
	c echo.Context,
	v validation.Validatable,
) *echo.HTTPError {
	if err := binder(c, v); err != nil {
		httpError := err.(*echo.HTTPError)
		s.logger.Info(
			"server.bind: binding failed",
			zap.Error(httpError.Internal),
		)
		return echo.NewHTTPError(
			httpError.Code,
			response.Error(failureStatus, httpError.Internal.Error()),
		)
	}
	if err := v.Validate(); err != nil {
		if validationInternalError, ok := err.(validation.InternalError); ok {
			s.logger.Error(
				"server.bind: validation internal error",
				zap.Error(validationInternalError),
			)
			return echo.NewHTTPError(
				http.StatusInternalServerError,
				response.Error(response.StatusInternalServerError),
			)
		}
		s.logger.Info("server.bind: validation failed", zap.Error(err))
		return echo.NewHTTPError(
			http.StatusBadRequest,
			response.Error(failureStatus, err.Error()),
		)
	}
	return nil
}

var defaultBinder = &echo.DefaultBinder{}
