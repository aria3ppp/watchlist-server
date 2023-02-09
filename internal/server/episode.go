package server

import (
	"net/http"

	"github.com/aria3ppp/watchlist-server/internal/app"
	"github.com/aria3ppp/watchlist-server/internal/config"
	"github.com/aria3ppp/watchlist-server/internal/dto"
	"github.com/aria3ppp/watchlist-server/internal/server/request"
	"github.com/aria3ppp/watchlist-server/internal/server/response"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// GET /v1/authorized/series/:id/season/:season_number/episode/:episode_number/
func (s *Server) HandleEpisodeGet(c echo.Context) error {
	// bind & validate params
	var params request.SeriesSeasonEpisodeNumberPathParam
	if httpError := s.bindPath(c, &params); httpError != nil {
		return httpError
	}

	// fetch episode
	episode, err := s.app.EpisodeGet(
		c.Request().Context(),
		params.SeriesID,
		params.SeasonNumber,
		params.EpisodeNumber,
	)
	if err != nil {
		if err == app.ErrNotFound {
			s.logger.Info(
				"server.HandleEpisodeGet: episdoe not found",
				zap.Int("series id", params.SeriesID),
				zap.Int("season number", params.SeriesID),
				zap.Int("episode number", params.SeriesID),
			)
			return echo.NewHTTPError(http.StatusNotFound)
		}

		s.logger.Error(
			"server.HandleEpisodeGet: internal server error", zap.Error(err),
		)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, episode)
}

// GET /v1/authorized/series/:id/episode/?page=1&page_size=100&sort_order=asc
func (s *Server) HandleEpisodesGetAllBySeries(c echo.Context) error {
	// bind & validate id param
	var param request.IDPathParam
	if httpError := s.bindPath(c, &param); httpError != nil {
		return httpError
	}

	// bind & validate query
	var pagQuery request.PaginationSortOrderQuery
	if httpError := s.bindQuery(c, &pagQuery); httpError != nil {
		return httpError
	}

	queryOptions := pagQuery.SetQueryIfNotSet(request.PaginationSortOrderQuery{
		PaginationQuery: request.PaginationQuery{
			Page:     config.Config.Validation.Pagination.Page.MinValue,
			PageSize: config.Config.Validation.Pagination.PageSize.DefaultValue,
		},
		SortOrderQuery: request.SortOrderQuery{
			SortOrder: request.SortOrderAsc,
		},
	}).ToQueryOptions()

	// fetch episodes
	episodes, total, err := s.app.EpisodesGetAllBySeries(
		c.Request().Context(),
		param.ID,
		queryOptions,
	)
	if err != nil {
		if err == app.ErrNotFound {
			s.logger.Info(
				"server.HandleEpisodesGetAllBySeries: series not found",
				zap.Int("series id", param.ID),
			)
			return echo.NewHTTPError(http.StatusNotFound)
		}

		s.logger.Error(
			"server.HandleEpisodesGetAllBySeries: internal server error",
			zap.Error(err),
		)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(
		http.StatusOK,
		response.Paginated(
			pagQuery.Page,
			pagQuery.PageSize,
			episodes,
			total,
		),
	)
}

// GET /v1/authorized/series/:id/season/:season_number/episode/?page=1&page_size=100&sort_order=asc
func (s *Server) HandleEpisodesGetAllBySeason(c echo.Context) error {
	// bind & validate params
	var params request.SeriesSeasonNumberPathParam
	if httpError := s.bindPath(c, &params); httpError != nil {
		return httpError
	}

	// bind & validate query
	var pagQuery request.PaginationSortOrderQuery
	if httpError := s.bindQuery(c, &pagQuery); httpError != nil {
		return httpError
	}

	queryOptions := pagQuery.SetQueryIfNotSet(request.PaginationSortOrderQuery{
		PaginationQuery: request.PaginationQuery{
			Page:     config.Config.Validation.Pagination.Page.MinValue,
			PageSize: config.Config.Validation.Pagination.PageSize.DefaultValue,
		},
		SortOrderQuery: request.SortOrderQuery{
			SortOrder: request.SortOrderAsc,
		},
	}).ToQueryOptions()

	// fetch episodes
	episodes, total, err := s.app.EpisodesGetAllBySeason(
		c.Request().Context(),
		params.SeriesID,
		params.SeasonNumber,
		queryOptions,
	)
	if err != nil {
		if err == app.ErrNotFound {
			s.logger.Info(
				"server.HandleEpisodesGetAllBySeason: series not found",
				zap.Int("series id", params.SeriesID),
			)
			return echo.NewHTTPError(http.StatusNotFound)
		}

		s.logger.Error(
			"server.HandleEpisodesGetAllBySeason: internal server error",
			zap.Error(err),
		)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(
		http.StatusOK,
		response.Paginated(
			pagQuery.Page,
			pagQuery.PageSize,
			episodes,
			total,
		),
	)
}

// PUT /v1/authorized/series/:id/season/:season_number/episode/:episode_number/
func (s *Server) HandleEpisodePut(c echo.Context) error {
	// bind & validate params
	var params request.SeriesSeasonEpisodeNumberPathParam
	if httpError := s.bindPath(c, &params); httpError != nil {
		return httpError
	}

	// bind & validate request
	var req dto.EpisodePutRequest
	if httpError := s.bindBody(c, &req); httpError != nil {
		return httpError
	}

	payload, httpError := s.getUserPayload(c)
	if httpError != nil {
		return httpError
	}

	// put episode
	err := s.app.EpisodePut(
		c.Request().Context(),
		params.SeriesID,
		params.SeasonNumber,
		params.EpisodeNumber,
		payload.UserID,
		&req,
	)
	if err != nil {
		if err == app.ErrNotFound {
			s.logger.Info(
				"server.HandleEpisodePut: series not found",
				zap.Int("series id", params.SeriesID),
			)
			return echo.NewHTTPError(http.StatusNotFound)
		}

		s.logger.Error(
			"server.HandleEpisodePut: internal server error",
			zap.Error(err),
		)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

// PUT /v1/authorized/series/:id/season/:season_number/episode/
func (s *Server) HandleEpisodesPutAllBySeason(c echo.Context) error {
	// bind & validate params
	var params request.SeriesSeasonNumberPathParam
	if httpError := s.bindPath(c, &params); httpError != nil {
		return httpError
	}

	// bind & validate request
	var req dto.EpisodesPutAllBySeasonRequest
	if httpError := s.bindBody(c, &req); httpError != nil {
		return httpError
	}

	payload, httpError := s.getUserPayload(c)
	if httpError != nil {
		return httpError
	}

	// put all episodes by season
	err := s.app.EpisodesPutAllBySeason(
		c.Request().Context(),
		params.SeriesID,
		params.SeasonNumber,
		payload.UserID,
		&req,
	)
	if err != nil {
		if err == app.ErrNotFound {
			s.logger.Info(
				"server.HandleEpisodesPutAllBySeason: series not found",
				zap.Int("series id", params.SeriesID),
			)
			return echo.NewHTTPError(http.StatusNotFound)
		}

		s.logger.Error(
			"server.HandleEpisodesPutAllBySeason: internal server error",
			zap.Error(err),
		)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

// UPDATE /v1/authorized/series/:id/season/:season_number/episode/:episode_number/
func (s *Server) HandleEpisodeUpdate(c echo.Context) error {
	// bind & validate params
	var params request.SeriesSeasonEpisodeNumberPathParam
	if httpError := s.bindPath(c, &params); httpError != nil {
		return httpError
	}

	// bind & validate request
	var req dto.EpisodeUpdateRequest
	if httpError := s.bindBody(c, &req); httpError != nil {
		return httpError
	}

	payload, httpError := s.getUserPayload(c)
	if httpError != nil {
		return httpError
	}

	// update episode
	err := s.app.EpisodeUpdate(
		c.Request().Context(),
		params.SeriesID,
		params.SeasonNumber,
		params.EpisodeNumber,
		payload.UserID,
		&req,
	)
	if err != nil {
		if err == app.ErrNotFound {
			s.logger.Info(
				"server.HandleEpisodeUpdate: episode not found",
				zap.Int("series id", params.SeriesID),
				zap.Int("season number", params.SeasonNumber),
				zap.Int("episode number", params.EpisodeNumber),
			)
			return echo.NewHTTPError(http.StatusNotFound)
		}

		s.logger.Error(
			"server.HandleEpisodeUpdate: internal server error",
			zap.Error(err),
		)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

// POST /v1/authorized/series/:id/season/:season_number/episode/:episode_number/invalidate
func (s *Server) HandleEpisodeInvalidate(c echo.Context) error {
	// bind & validate params
	var params request.SeriesSeasonEpisodeNumberPathParam
	if httpError := s.bindPath(c, &params); httpError != nil {
		return httpError
	}

	// bind & validate request
	var req dto.InvalidationRequest
	if httpError := s.bindBody(c, &req); httpError != nil {
		return httpError
	}

	payload, httpError := s.getUserPayload(c)
	if httpError != nil {
		return httpError
	}

	// invalidate episode
	err := s.app.EpisodeInvalidate(
		c.Request().Context(),
		params.SeriesID,
		params.SeasonNumber,
		params.EpisodeNumber,
		payload.UserID,
		&req,
	)
	if err != nil {
		if err == app.ErrNotFound {
			s.logger.Info(
				"server.HandleEpisodeInvalidate: episode not found",
				zap.Int("series id", params.SeriesID),
				zap.Int("season number", params.SeasonNumber),
				zap.Int("episode number", params.EpisodeNumber),
			)
			return echo.NewHTTPError(http.StatusNotFound)
		}

		s.logger.Error(
			"server.HandleEpisodeInvalidate: internal server error",
			zap.Error(err),
		)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

// POST /v1/authorized/series/:id/season/:season_number/episode/invalidate
func (s *Server) HandleEpisodesInvalidateAllBySeason(c echo.Context) error {
	// bind & validate params
	var params request.SeriesSeasonNumberPathParam
	if httpError := s.bindPath(c, &params); httpError != nil {
		return httpError
	}

	// bind & validate request
	var req dto.InvalidationRequest
	if httpError := s.bindBody(c, &req); httpError != nil {
		return httpError
	}

	payload, httpError := s.getUserPayload(c)
	if httpError != nil {
		return httpError
	}

	// invalidate all episodes
	err := s.app.EpisodesInvalidateAllBySeason(
		c.Request().Context(),
		params.SeriesID,
		params.SeasonNumber,
		payload.UserID,
		&req,
	)
	if err != nil {
		if err == app.ErrNotFound {
			s.logger.Info(
				"server.HandleEpisodesInvalidateAllBySeason: season not found",
				zap.Int("series id", params.SeriesID),
				zap.Int("season number", params.SeasonNumber),
			)
			return echo.NewHTTPError(http.StatusNotFound)
		}

		s.logger.Error(
			"server.HandleEpisodesInvalidateAllBySeason: internal server error",
			zap.Error(err),
		)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

// GET /v1/authorized/series/:id/season/:season_number/episode/:episode_number/audits/?page=1&page_size=100&sort_order=desc
func (s *Server) HandleEpisodeAuditsGetAll(c echo.Context) error {
	// bind & validate params
	var params request.SeriesSeasonEpisodeNumberPathParam
	if httpError := s.bindPath(c, &params); httpError != nil {
		return httpError
	}

	// bind & validate query
	var pagQuery request.PaginationSortOrderQuery
	if httpError := s.bindQuery(c, &pagQuery); httpError != nil {
		return httpError
	}

	queryOptions := pagQuery.SetQueryIfNotSet(request.PaginationSortOrderQuery{
		PaginationQuery: request.PaginationQuery{
			Page:     config.Config.Validation.Pagination.Page.MinValue,
			PageSize: config.Config.Validation.Pagination.PageSize.DefaultValue,
		},
		SortOrderQuery: request.SortOrderQuery{
			SortOrder: request.SortOrderDesc,
		},
	}).ToQueryOptions()

	// fetch audits
	audits, total, err := s.app.EpisodeAuditsGetAll(
		c.Request().Context(),
		params.SeriesID,
		params.SeasonNumber,
		params.EpisodeNumber,
		queryOptions,
	)
	if err != nil {
		if err == app.ErrNotFound {
			s.logger.Info(
				"server.HandleEpisodeAuditsGetAll: episode not found",
				zap.Int("series id", params.SeriesID),
				zap.Int("season number", params.SeasonNumber),
				zap.Int("episode number", params.EpisodeNumber),
			)
			return echo.NewHTTPError(http.StatusNotFound)
		}

		s.logger.Error(
			"server.HandleEpisodeAuditsGetAll: internal server error",
			zap.Error(err),
		)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(
		http.StatusOK,
		response.Paginated(
			pagQuery.Page,
			pagQuery.PageSize,
			audits,
			total,
		),
	)
}
