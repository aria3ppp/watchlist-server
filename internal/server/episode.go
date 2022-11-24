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

// GET /v1/authorized/series/:id/season/:season_number/episode/:episode_number/
func (s *Server) HandleEpisodeGet(c echo.Context) error {
	// bind & validate params
	var params request.SeriesSeasonEpisodeNumberPathParam
	err := (&echo.DefaultBinder{}).BindPathParams(c, &params)
	if err == nil {
		err = params.Validate()
	}
	if err != nil {
		s.logger.Info(
			"server.HandleEpisodeGet: parameter binding/validation failed",
			zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusBadRequest,
			response.Error(response.StatusInvalidURLParameter),
		)
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
			return echo.NewHTTPError(
				http.StatusNotFound,
				response.Error(response.StatusNotFound),
			)
		}

		s.logger.Error(
			"server.HandleEpisodeGet: internal server error", zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			response.Error(response.StatusInternalServerError),
		)
	}

	return c.JSON(http.StatusOK, response.OK(episode))
}

// GET /v1/authorized/series/:id/episode/?page=1&per_page=100
func (s *Server) HandleEpisodesGetAllBySeries(c echo.Context) error {
	// bind & validate params
	var params request.IDPathParam
	err := (&echo.DefaultBinder{}).BindPathParams(c, &params)
	if err == nil {
		err = params.Validate()
	}
	if err != nil {
		s.logger.Info(
			"server.HandleEpisodesGetAllBySeries: parameter binding/validation failed",
			zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusBadRequest,
			response.Error(response.StatusInvalidURLParameter),
		)
	}

	// parse pagination params
	page, perPage, offset := request.ParsePaginationQueries(c.Request())

	// fetch episodes
	episodes, total, err := s.app.EpisodesGetAllBySeries(
		c.Request().Context(),
		params.ID,
		offset,
		perPage,
	)
	if err != nil {
		s.logger.Error(
			"server.HandleEpisodesGetAllBySeries: internal server error",
			zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			response.Error(response.StatusInternalServerError),
		)
	}

	return c.JSON(
		http.StatusOK,
		response.Paginated(page, perPage, episodes, total),
	)
}

// GET /v1/authorized/series/:id/season/:season_number/episode/?page=1&per_page=100
func (s *Server) HandleEpisodesGetAllBySeason(c echo.Context) error {
	// bind & validate params
	var params request.SeriesSeasonNumberPathParam
	err := (&echo.DefaultBinder{}).BindPathParams(c, &params)
	if err == nil {
		err = params.Validate()
	}
	if err != nil {
		s.logger.Info(
			"server.HandleEpisodesGetAllBySeason: parameter binding/validation failed",
			zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusBadRequest,
			response.Error(response.StatusInvalidURLParameter),
		)
	}

	// parse pagination params
	page, perPage, offset := request.ParsePaginationQueries(c.Request())

	// fetch episodes
	episodes, total, err := s.app.EpisodesGetAllBySeason(
		c.Request().Context(),
		params.SeriesID,
		params.SeasonNumber,
		offset,
		perPage,
	)
	if err != nil {
		s.logger.Error(
			"server.HandleEpisodesGetAllBySeason: internal server error",
			zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			response.Error(response.StatusInternalServerError),
		)
	}

	return c.JSON(
		http.StatusOK,
		response.Paginated(page, perPage, episodes, total),
	)
}

// PUT /v1/authorized/series/:id/season/:season_number/episode/:episode_number/
func (s *Server) HandleEpisodePut(c echo.Context) error {
	// bind & validate params
	var params request.SeriesSeasonEpisodeNumberPathParam
	err := (&echo.DefaultBinder{}).BindPathParams(c, &params)
	if err == nil {
		err = params.Validate()
	}
	if err != nil {
		s.logger.Info(
			"server.HandleEpisodePut: parameter binding/validation failed",
			zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusBadRequest,
			response.Error(response.StatusInvalidURLParameter),
		)
	}

	// bind & validate request
	var req dto.EpisodePutRequest
	err = (&echo.DefaultBinder{}).BindBody(c, &req)
	if err == nil {
		err = req.Validate()
	}
	if err != nil {
		s.logger.Info(
			"server.HandleEpisodePut: request binding/validation failed",
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
			"server.HandleEpisodePut: payload key not set on router context",
			zap.String("payload key", PayloadKey),
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			response.Error(response.StatusInternalServerError),
		)
	}

	// put episode
	err = s.app.EpisodePut(
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
			return echo.NewHTTPError(
				http.StatusNotFound,
				response.Error(response.StatusNotFound),
			)
		}

		s.logger.Error(
			"server.HandleEpisodePut: internal server error",
			zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			response.Error(response.StatusInternalServerError),
		)
	}

	return c.JSON(http.StatusOK, response.OK(nil))
}

// PUT /v1/authorized/series/:id/season/:season_number/episode/
func (s *Server) HandleEpisodesPutAllBySeason(c echo.Context) error {
	// bind & validate params
	var params request.SeriesSeasonNumberPathParam
	err := (&echo.DefaultBinder{}).BindPathParams(c, &params)
	if err == nil {
		err = params.Validate()
	}
	if err != nil {
		s.logger.Info(
			"server.HandleEpisodesPutAllBySeason: parameter binding/validation failed",
			zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusBadRequest,
			response.Error(response.StatusInvalidURLParameter),
		)
	}

	// bind & validate request
	var req dto.EpisodesPutAllBySeasonRequest
	err = (&echo.DefaultBinder{}).BindBody(c, &req)
	if err == nil {
		err = req.Validate()
	}
	if err != nil {
		s.logger.Info(
			"server.HandleEpisodesPutAllBySeason: request binding/validation failed",
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
			"server.HandleEpisodesPutAllBySeason: payload key not set on router context",
			zap.String("payload key", PayloadKey),
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			response.Error(response.StatusInternalServerError),
		)
	}

	// put all episodes by season
	err = s.app.EpisodesPutAllBySeason(
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
			return echo.NewHTTPError(
				http.StatusNotFound,
				response.Error(response.StatusNotFound),
			)
		}

		s.logger.Error(
			"server.HandleEpisodesPutAllBySeason: internal server error",
			zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			response.Error(response.StatusInternalServerError),
		)
	}

	return c.JSON(http.StatusOK, response.OK(nil))
}

// UPDATE /v1/authorized/series/:id/season/:season_number/episode/:episode_number/
func (s *Server) HandleEpisodeUpdate(c echo.Context) error {
	// bind & validate params
	var params request.SeriesSeasonEpisodeNumberPathParam
	err := (&echo.DefaultBinder{}).BindPathParams(c, &params)
	if err == nil {
		err = params.Validate()
	}
	if err != nil {
		s.logger.Info(
			"server.HandleEpisodeUpdate: parameter binding/validation failed",
			zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusBadRequest,
			response.Error(response.StatusInvalidURLParameter),
		)
	}

	// bind & validate request
	var req dto.EpisodeUpdateRequest
	err = (&echo.DefaultBinder{}).BindBody(c, &req)
	if err == nil {
		err = req.Validate()
	}
	if err != nil {
		s.logger.Info(
			"server.HandleEpisodeUpdate: request binding/validation failed",
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
			"server.HandleEpisodeUpdate: payload key not set on router context",
			zap.String("payload key", PayloadKey),
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			response.Error(response.StatusInternalServerError),
		)
	}

	// update episode
	err = s.app.EpisodeUpdate(
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
			return echo.NewHTTPError(
				http.StatusNotFound,
				response.Error(response.StatusNotFound),
			)
		}

		s.logger.Error(
			"server.HandleEpisodeUpdate: internal server error",
			zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			response.Error(response.StatusInternalServerError),
		)
	}

	return c.JSON(http.StatusOK, response.OK(nil))
}

// DELETE /v1/authorized/series/:id/season/:season_number/episode/:episode_number/
func (s *Server) HandleEpisodeInvalidate(c echo.Context) error {
	// bind & validate params
	var params request.SeriesSeasonEpisodeNumberPathParam
	err := (&echo.DefaultBinder{}).BindPathParams(c, &params)
	if err == nil {
		err = params.Validate()
	}
	if err != nil {
		s.logger.Info(
			"server.HandleEpisodeInvalidate: parameter binding/validation failed",
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
			"server.HandleEpisodeInvalidate: request binding/validation failed",
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
			"server.HandleEpisodeInvalidate: payload key not set on router context",
			zap.String("payload key", PayloadKey),
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			response.Error(response.StatusInternalServerError),
		)
	}

	// invalidate episode
	err = s.app.EpisodeInvalidate(
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
			return echo.NewHTTPError(
				http.StatusNotFound,
				response.Error(response.StatusNotFound),
			)
		}

		s.logger.Error(
			"server.HandleEpisodeInvalidate: internal server error",
			zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			response.Error(response.StatusInternalServerError),
		)
	}

	return c.JSON(http.StatusOK, response.OK(nil))
}

// DELETE /v1/authorized/series/:id/season/:season_number/episode/
func (s *Server) HandleEpisodesInvalidateAllBySeason(c echo.Context) error {
	// bind & validate params
	var params request.SeriesSeasonNumberPathParam
	err := (&echo.DefaultBinder{}).BindPathParams(c, &params)
	if err == nil {
		err = params.Validate()
	}
	if err != nil {
		s.logger.Info(
			"server.HandleEpisodesInvalidateAllBySeason: parameter binding/validation failed",
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
			"server.HandleEpisodesInvalidateAllBySeason: request binding/validation failed",
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
			"server.HandleEpisodesInvalidateAllBySeason: payload key not set on router context",
			zap.String("payload key", PayloadKey),
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			response.Error(response.StatusInternalServerError),
		)
	}

	// invalidate all episodes
	err = s.app.EpisodesInvalidateAllBySeason(
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
			return echo.NewHTTPError(
				http.StatusNotFound,
				response.Error(response.StatusNotFound),
			)
		}

		s.logger.Error(
			"server.HandleEpisodesInvalidateAllBySeason: internal server error",
			zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			response.Error(response.StatusInternalServerError),
		)
	}

	return c.JSON(http.StatusOK, response.OK(nil))
}

// GET /v1/authorized/series/:id/season/:season_number/episode/:episode_number/audits/?page=1&per_page=100
func (s *Server) HandleEpisodeAuditsGetAll(c echo.Context) error {
	// bind & validate params
	var params request.SeriesSeasonEpisodeNumberPathParam
	err := (&echo.DefaultBinder{}).BindPathParams(c, &params)
	if err == nil {
		err = params.Validate()
	}
	if err != nil {
		s.logger.Info(
			"server.HandleEpisodeAuditsGetAll: parameter binding/validation failed",
			zap.Error(err),
		)
		return echo.NewHTTPError(
			http.StatusBadRequest,
			response.Error(response.StatusInvalidURLParameter),
		)
	}

	page, perPage, offset := request.ParsePaginationQueries(c.Request())

	// fetch audits
	audits, total, err := s.app.EpisodeAuditsGetAll(
		c.Request().Context(),
		params.SeriesID,
		params.SeasonNumber,
		params.EpisodeNumber,
		offset,
		perPage,
	)
	if err != nil {
		if err == app.ErrNotFound {
			s.logger.Info(
				"server.HandleEpisodeAuditsGetAll: episode not found",
				zap.Int("series id", params.SeriesID),
				zap.Int("season number", params.SeasonNumber),
				zap.Int("episode number", params.EpisodeNumber),
			)
			return echo.NewHTTPError(
				http.StatusNotFound,
				response.Error(response.StatusNotFound),
			)
		}

		s.logger.Error(
			"server.HandleEpisodeAuditsGetAll: internal server error",
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
