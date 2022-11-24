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

// GET /v1/authorized/movie/:id/
func (s *Server) HandleMovieGet(c echo.Context) error {
	// bind & validate params
	var params request.IDPathParam
	err := (&echo.DefaultBinder{}).BindPathParams(c, &params)
	if err == nil {
		err = params.Validate()
	}
	if err != nil {
		s.logger.Info(
			"server.HandleMovieGet: parameter binding/validation failed",
			zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusBadRequest,
			response.Error(response.StatusInvalidURLParameter),
		)
	}

	// fetch movie
	movie, err := s.app.MovieGet(c.Request().Context(), params.ID)
	if err != nil {
		if err == app.ErrNotFound {
			s.logger.Info(
				"server.HandleMovieGet: movie not found",
				zap.Int("id", params.ID),
			)
			return echo.NewHTTPError(
				http.StatusNotFound,
				response.Error(response.StatusNotFound),
			)
		}

		s.logger.Error(
			"server.HandleMovieGet: internal server error", zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			response.Error(response.StatusInternalServerError),
		)
	}

	return c.JSON(http.StatusOK, response.OK(movie))
}

// GET /v1/authorized/movie/?page=1&per_page=100
func (s *Server) HandleMoviesGetAll(c echo.Context) error {
	// parse pagination queries
	page, perPage, offset := request.ParsePaginationQueries(c.Request())

	// fetch movies
	movies, total, err := s.app.MoviesGetAll(
		c.Request().Context(),
		offset,
		perPage,
	)
	if err != nil {
		s.logger.Error(
			"server.HandleMoviesGetAll: internal server error",
			zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			response.Error(response.StatusInternalServerError),
		)
	}

	return c.JSON(
		http.StatusOK,
		response.Paginated(page, perPage, movies, total),
	)
}

// POST /v1/authorized/movie/
func (s *Server) HandleMovieCreate(c echo.Context) error {
	// bind & validate request
	var req dto.MovieCreateRequest
	err := (&echo.DefaultBinder{}).BindBody(c, &req)
	if err == nil {
		err = req.Validate()
	}
	if err != nil {
		s.logger.Info(
			"server.HandleMovieCreate: request binding/validation failed",
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
			"server.HandleMovieCreate: payload key not set on router context",
			zap.String("payload key", PayloadKey),
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			response.Error(response.StatusInternalServerError),
		)
	}

	// create movie
	movieID, err := s.app.MovieCreate(
		c.Request().Context(),
		payload.UserID,
		&req,
	)
	if err != nil {
		s.logger.Error(
			"server.HandleMovieCreate: internal server error",
			zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			response.Error(response.StatusInternalServerError),
		)
	}

	return c.JSON(http.StatusOK, response.OK(movieID))
}

// UPDATE /v1/authorized/movie/:id/
func (s *Server) HandleMovieUpdate(c echo.Context) error {
	// bind & validate params
	var params request.IDPathParam
	err := (&echo.DefaultBinder{}).BindPathParams(c, &params)
	if err == nil {
		err = params.Validate()
	}
	if err != nil {
		s.logger.Info(
			"server.HandleMovieUpdate: parameter binding/validation failed",
			zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusBadRequest,
			response.Error(response.StatusInvalidURLParameter),
		)
	}

	// bind & validate request
	var req dto.MovieUpdateRequest
	err = (&echo.DefaultBinder{}).BindBody(c, &req)
	if err == nil {
		err = req.Validate()
	}
	if err != nil {
		s.logger.Info(
			"server.HandleMovieUpdate: request binding/validation failed",
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
			"server.HandleMovieUpdate: payload key not set on router context",
			zap.String("payload key", PayloadKey),
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			response.Error(response.StatusInternalServerError),
		)
	}

	// update movie
	err = s.app.MovieUpdate(
		c.Request().Context(),
		params.ID,
		payload.UserID,
		&req,
	)
	if err != nil {
		if err == app.ErrNotFound {
			s.logger.Info(
				"server.HandleMovieUpdate: movie not found",
				zap.Int("id", params.ID),
			)
			return echo.NewHTTPError(
				http.StatusNotFound,
				response.Error(response.StatusNotFound),
			)
		}

		s.logger.Error(
			"server.HandleMovieUpdate: internal server error",
			zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			response.Error(response.StatusInternalServerError),
		)
	}

	return c.JSON(http.StatusOK, response.OK(nil))
}

// DELETE /v1/authorized/movie/:id/
func (s *Server) HandleMovieInvalidate(c echo.Context) error {
	// bind & validate params
	var params request.IDPathParam
	err := (&echo.DefaultBinder{}).BindPathParams(c, &params)
	if err == nil {
		err = params.Validate()
	}
	if err != nil {
		s.logger.Info(
			"server.HandleMovieInvalidate: parameter binding/validation failed",
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
			"server.HandleMovieInvalidate: request binding/validation failed",
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
			"server.HandleMovieInvalidate: payload key not set on router context",
			zap.String("payload key", PayloadKey),
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			response.Error(response.StatusInternalServerError),
		)
	}

	// invalidate movie
	err = s.app.MovieInvalidate(
		c.Request().Context(),
		params.ID,
		payload.UserID,
		&req,
	)
	if err != nil {
		if err == app.ErrNotFound {
			s.logger.Info(
				"server.HandleMovieInvalidate: movie not found",
				zap.Int("id", params.ID),
			)
			return echo.NewHTTPError(
				http.StatusNotFound,
				response.Error(response.StatusNotFound),
			)
		}

		s.logger.Error(
			"server.HandleMovieInvalidate: internal server error",
			zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			response.Error(response.StatusInternalServerError),
		)
	}

	return c.JSON(http.StatusOK, response.OK(nil))
}

// GET /v1/authorized/movie/:id/?page=1&per_page=100
func (s *Server) HandleMovieAuditsGetAll(c echo.Context) error {
	// bind & validate params
	var params request.IDPathParam
	err := (&echo.DefaultBinder{}).BindPathParams(c, &params)
	if err == nil {
		err = params.Validate()
	}
	if err != nil {
		s.logger.Info(
			"server.HandleMovieAuditsGetAll: parameter binding/validation failed",
			zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusBadRequest,
			response.Error(response.StatusInvalidURLParameter),
		)
	}

	page, perPage, offset := request.ParsePaginationQueries(c.Request())

	// fetch audits
	audits, total, err := s.app.MovieAuditsGetAll(
		c.Request().Context(),
		params.ID,
		offset,
		perPage,
	)
	if err != nil {
		if err == app.ErrNotFound {
			s.logger.Info(
				"server.HandleMovieAuditsGetAll: movie not found",
				zap.Int("id", params.ID),
			)
			return echo.NewHTTPError(
				http.StatusNotFound,
				response.Error(response.StatusNotFound),
			)
		}

		s.logger.Error(
			"server.HandleMovieAuditsGetAll: internal server error",
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

// GET /v1/authorized/movie/search/?query=query&page=1&per_page=60
func (s *Server) HandleMoviesSearch(c echo.Context) error {
	// bind & validate params
	var query request.SearchQuery
	err := (&echo.DefaultBinder{}).BindQueryParams(c, &query)
	if err == nil {
		err = query.Validate()
	}
	if err != nil {
		s.logger.Info(
			"server.HandleMoviesSearch: parameter binding/validation failed",
			zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusBadRequest,
			response.Error(response.StatusInvalidURLParameter, err.Error()),
		)
	}

	page, perPage, offset := request.ParsePaginationQueries(c.Request())

	// fetch movies
	movies, total, err := s.app.MoviesSearch(
		c.Request().Context(),
		query.Query,
		offset,
		perPage,
	)
	if err != nil {
		s.logger.Error(
			"server.HandleMoviesSearch: internal server error",
			zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			response.Error(response.StatusInternalServerError),
		)
	}

	return c.JSON(
		http.StatusOK,
		response.Paginated(page, perPage, movies, total),
	)
}
