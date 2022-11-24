package server

import (
	"net/http"

	"github.com/aria3ppp/watchlist-server/internal/app"
	"github.com/aria3ppp/watchlist-server/internal/dto"
	"github.com/aria3ppp/watchlist-server/internal/server/request"
	"github.com/aria3ppp/watchlist-server/internal/server/response"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// GET /v1/authorized/series/:id/
func (s *Server) HandleSeriesGet(c echo.Context) error {
	// bind & validate params
	var params request.IDPathParam
	err := (&echo.DefaultBinder{}).BindPathParams(c, &params)
	if err == nil {
		err = params.Validate()
	}
	if err != nil {
		s.logger.Info(
			"server.HandleSeriesGet: parameter binding/validation failed",
			zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusBadRequest,
			response.Error(response.StatusInvalidURLParameter),
		)
	}

	// fetch series
	series, err := s.app.SeriesGet(c.Request().Context(), params.ID)
	if err != nil {
		if err == app.ErrNotFound {
			s.logger.Info(
				"server.HandleSeriesGet: series not found",
				zap.Int("id", params.ID),
			)
			return echo.NewHTTPError(
				http.StatusNotFound,
				response.Error(response.StatusNotFound),
			)
		}

		s.logger.Error(
			"server.HandleSeriesGet: internal server error", zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			response.Error(response.StatusInternalServerError),
		)
	}

	return c.JSON(http.StatusOK, response.OK(series))
}

// GET /v1/authorized/series/?page=1&per_page=60
func (s *Server) HandleSeriesesGetAll(c echo.Context) error {
	// parse pagination params
	page, perPage, offset := request.ParsePaginationQueries(c.Request())

	// fetch serieses
	serieses, total, err := s.app.SeriesesGetAll(
		c.Request().Context(),
		offset,
		perPage,
	)
	if err != nil {
		s.logger.Error(
			"server.HandleSeriesesGetAll: internal server error",
			zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			response.Error(response.StatusInternalServerError),
		)
	}

	return c.JSON(
		http.StatusOK,
		response.Paginated(page, perPage, serieses, total),
	)
}

// POST /v1/authorized/series/
func (s *Server) HandleSeriesCreate(c echo.Context) error {
	// bind & validate request
	var req dto.SeriesCreateRequest
	err := (&echo.DefaultBinder{}).BindBody(c, &req)
	if err == nil {
		err = req.Validate()
	}
	if err != nil {
		s.logger.Info(
			"server.HandleSeriesCreate: request binding/validation failed",
			zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusBadRequest,
			response.Error(response.StatusInvalidRequest, err.Error()),
		)
	}

	payload := FetchUserPayload(c)
	if payload == nil {
		s.logger.Error(
			"server.HandleSeriesCreate: payload key not set on router context",
			zap.String("payload key", PayloadKey),
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			response.Error(response.StatusInternalServerError),
		)
	}

	// create series
	seriesID, err := s.app.SeriesCreate(
		c.Request().Context(),
		payload.UserID,
		&req,
	)
	if err != nil {
		s.logger.Error(
			"server.HandleSeriesCreate: internal server error",
			zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			response.Error(response.StatusInternalServerError),
		)
	}

	return c.JSON(http.StatusOK, response.OK(seriesID))
}

// PATCH /v1/authorized/series/:id/
func (s *Server) HandleSeriesUpdate(c echo.Context) error {
	// bind & validate params
	var params request.IDPathParam
	err := (&echo.DefaultBinder{}).BindPathParams(c, &params)
	if err == nil {
		err = params.Validate()
	}
	if err != nil {
		s.logger.Info(
			"server.HandleSeriesUpdate: parameter binding/validation failed",
			zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusBadRequest,
			response.Error(response.StatusInvalidURLParameter),
		)
	}

	// bind & validate request
	var req dto.SeriesUpdateRequest
	err = (&echo.DefaultBinder{}).BindBody(c, &req)
	if err == nil {
		err = req.Validate()
	}
	if err != nil {
		s.logger.Info(
			"server.HandleSeriesUpdate: request binding/validation failed",
			zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusBadRequest,
			response.Error(response.StatusInvalidRequest, err.Error()),
		)
	}

	payload := FetchUserPayload(c)
	if payload == nil {
		s.logger.Error(
			"server.HandleSeriesUpdate: payload key not set on router context",
			zap.String("payload key", PayloadKey),
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			response.Error(response.StatusInternalServerError),
		)
	}

	// update series
	err = s.app.SeriesUpdate(
		c.Request().Context(),
		params.ID,
		payload.UserID,
		&req,
	)
	if err != nil {
		if err == app.ErrNotFound {
			s.logger.Info(
				"server.HandleSeriesUpdate: series not found",
				zap.Int("id", params.ID),
			)
			return echo.NewHTTPError(
				http.StatusNotFound,
				response.Error(response.StatusNotFound),
			)
		}

		s.logger.Error(
			"server.HandleSeriesUpdate: internal server error",
			zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			response.Error(response.StatusInternalServerError),
		)
	}

	return c.JSON(http.StatusOK, response.OK(nil))
}

// DELETE /v1/authorized/series/:id/
func (s *Server) HandleSeriesInvalidate(c echo.Context) error {
	// bind & validate params
	var params request.IDPathParam
	err := (&echo.DefaultBinder{}).BindPathParams(c, &params)
	if err == nil {
		err = params.Validate()
	}
	if err != nil {
		s.logger.Info(
			"server.HandleSeriesInvalidate: parameter binding/validation failed",
			zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusBadRequest,
			response.Error(response.StatusInvalidURLParameter),
		)
	}

	// bind & validate request
	var req dto.InvalidationRequest
	err = (&echo.DefaultBinder{}).BindBody(c, &req)
	if err == nil {
		err = req.Validate()
	}
	if err != nil {
		s.logger.Info(
			"server.HandleSeriesInvalidate: request binding/validation failed",
			zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusBadRequest,
			response.Error(response.StatusInvalidRequest, err.Error()),
		)
	}

	payload := FetchUserPayload(c)
	if payload == nil {
		s.logger.Error(
			"server.HandleSeriesInvalidate: payload key not set on router context",
			zap.String("payload key", PayloadKey),
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			response.Error(response.StatusInternalServerError),
		)
	}

	// invalidate series
	err = s.app.SeriesInvalidate(
		c.Request().Context(),
		params.ID,
		payload.UserID,
		&req,
	)
	if err != nil {
		if err == app.ErrNotFound {
			s.logger.Info(
				"server.HandleSeriesInvalidate: series not found",
				zap.Int("id", params.ID),
			)
			return echo.NewHTTPError(
				http.StatusNotFound,
				response.Error(response.StatusNotFound),
			)
		}

		s.logger.Error(
			"server.HandleSeriesInvalidate: internal server error",
			zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			response.Error(response.StatusInternalServerError),
		)
	}

	return c.JSON(http.StatusOK, response.OK(nil))
}

// GET /v1/authorized/series/:id/audits/?page=1&per_page=60
func (s *Server) HandleSeriesAuditsGetAll(c echo.Context) error {
	// bind & validate params
	var params request.IDPathParam
	err := (&echo.DefaultBinder{}).BindPathParams(c, &params)
	if err == nil {
		err = params.Validate()
	}
	if err != nil {
		s.logger.Info(
			"server.HandleSeriesAuditsGetAll: parameter binding/validation failed",
			zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusBadRequest,
			response.Error(response.StatusInvalidURLParameter),
		)
	}

	page, perPage, offset := request.ParsePaginationQueries(c.Request())

	// fetch audits
	audits, total, err := s.app.SeriesAuditsGetAll(
		c.Request().Context(),
		params.ID,
		offset,
		perPage,
	)
	if err != nil {
		if err == app.ErrNotFound {
			s.logger.Info(
				"server.HandleSeriesAuditsGetAll: series not found",
				zap.Int("id", params.ID),
			)
			return echo.NewHTTPError(
				http.StatusNotFound,
				response.Error(response.StatusNotFound),
			)
		}

		s.logger.Error(
			"server.HandleSeriesAuditsGetAll: internal server error",
			zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			response.Error(response.StatusInternalServerError),
		)
	}

	return c.JSON(
		http.StatusOK,
		response.Paginated(page, perPage, audits, total),
	)
}

// GET /v1/authorized/series/search/?query=query&page=1&per_page=60
func (s *Server) HandleSeriesesSearch(c echo.Context) error {
	// bind & validate params
	var query request.SearchQuery
	err := (&echo.DefaultBinder{}).BindQueryParams(c, &query)
	if err == nil {
		err = query.Validate()
	}
	if err != nil {
		s.logger.Info(
			"server.HandleSeriesesSearch: parameter binding/validation failed",
			zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusBadRequest,
			response.Error(response.StatusInvalidURLParameter, err.Error()),
		)
	}

	page, perPage, offset := request.ParsePaginationQueries(c.Request())

	// fetch serieses
	serieses, total, err := s.app.SeriesesSearch(
		c.Request().Context(),
		query.Query,
		offset,
		perPage,
	)
	if err != nil {
		s.logger.Error(
			"server.HandleSeriesesSearch: internal server error",
			zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			response.Error(response.StatusInternalServerError),
		)
	}

	return c.JSON(
		http.StatusOK,
		response.Paginated(page, perPage, serieses, total),
	)
}
