package server

import (
	"net/http"

	"github.com/aria3ppp/watchlist-server/internal/app"
	"github.com/aria3ppp/watchlist-server/internal/config"
	"github.com/aria3ppp/watchlist-server/internal/server/request"
	"github.com/aria3ppp/watchlist-server/internal/server/response"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// GET /v1/authorized/watchlist/?page=12&page_size=10&sort_order=desc&filter=all
func (s *Server) HandleWatchlistGet(c echo.Context) error {
	// bind & validate query
	var query request.WatchlistGetQuery
	if httpError := s.bindQuery(c, &query); httpError != nil {
		return httpError
	}

	payload, httpError := s.getUserPayload(c)
	if httpError != nil {
		return httpError
	}

	queryOptions := query.SetQueryIfNotSet(request.WatchlistGetQuery{
		Filter: request.WatchlistFilterAll,
		SortOrderQuery: request.SortOrderQuery{
			SortOrder: request.SortOrderDesc,
		},
		PaginationQuery: request.PaginationQuery{
			Page:     config.Config.Validation.Pagination.Page.MinValue,
			PageSize: config.Config.Validation.Pagination.PageSize.DefaultValue,
		},
	}).ToQueryOptions()

	// fetch watchlist
	watchlist, total, err := s.app.WatchlistGet(
		c.Request().Context(),
		payload.UserID,
		queryOptions,
	)
	if err != nil {
		s.logger.Error(
			"server.HandleWatchlistGet: internal server error",
			zap.Error(err),
		)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(
		http.StatusOK,
		response.Paginated(
			query.Page,
			query.PageSize,
			watchlist,
			total,
		),
	)
}

// POST /v1/authorized/watchlist/add/?film_id=345
func (s *Server) HandleWatchlistAdd(c echo.Context) error {
	// bind & validate query
	var query request.WatchlistAddQuery
	if httpError := s.bindQuery(c, &query); httpError != nil {
		return httpError
	}

	payload, httpError := s.getUserPayload(c)
	if httpError != nil {
		return httpError
	}

	// add film to watchlist
	watchID, err := s.app.WatchlistAdd(
		c.Request().Context(),
		payload.UserID,
		query.FilmID,
	)
	if err != nil {
		if err == app.ErrNotFound {
			s.logger.Info(
				"server.HandleWatchlistAdd: film not found",
				zap.Int("film id", query.FilmID),
			)
			return echo.NewHTTPError(http.StatusNotFound)
		}

		s.logger.Error(
			"server.HandleWatchlistAdd: internal server error",
			zap.Error(err),
		)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, response.ID(watchID))
}

// DELETE /v1/authorized/watchlist/:id/
func (s *Server) HandleWatchlistDelete(c echo.Context) error {
	// bind & validate id param
	var param request.IDPathParam
	if httpError := s.bindPath(c, &param); httpError != nil {
		return httpError
	}

	payload, httpError := s.getUserPayload(c)
	if httpError != nil {
		return httpError
	}

	// delete film from watchlist
	err := s.app.WatchlistDelete(
		c.Request().Context(),
		payload.UserID,
		param.ID,
	)
	if err != nil {
		if err == app.ErrNotFound {
			s.logger.Info(
				"server.HandleWatchlistDelete: watchlist record not found",
				zap.Int("user id", payload.UserID),
				zap.Int("watch id", param.ID),
			)
			return echo.NewHTTPError(http.StatusNotFound)
		}

		s.logger.Error(
			"server.HandleWatchlistDelete: internal server error",
			zap.Error(err),
		)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

// PATCH /v1/authorized/watchlist/:id/
func (s *Server) HandleWatchlistSetWatched(c echo.Context) error {
	// bind & validate id param
	var param request.IDPathParam
	if httpError := s.bindPath(c, &param); httpError != nil {
		return httpError
	}

	payload, httpError := s.getUserPayload(c)
	if httpError != nil {
		return httpError
	}

	// set film watched
	err := s.app.WatchlistSetWatched(
		c.Request().Context(),
		payload.UserID,
		param.ID,
	)
	if err != nil {
		if err == app.ErrNotFound {
			s.logger.Info(
				"server.HandleWatchlistSetWatched: watchlist record not found",
				zap.Int("user id", payload.UserID),
				zap.Int("watch id", param.ID),
			)
			return echo.NewHTTPError(http.StatusNotFound)
		}

		s.logger.Error(
			"server.HandleWatchlistSetWatched: internal server error",
			zap.Error(err),
		)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}
