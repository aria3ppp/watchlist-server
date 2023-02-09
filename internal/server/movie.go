package server

import (
	"net/http"

	"github.com/aria3ppp/watchlist-server/internal/app"
	"github.com/aria3ppp/watchlist-server/internal/config"
	"github.com/aria3ppp/watchlist-server/internal/dto"
	"github.com/aria3ppp/watchlist-server/internal/models"
	"github.com/aria3ppp/watchlist-server/internal/server/request"
	"github.com/aria3ppp/watchlist-server/internal/server/response"
	"github.com/aria3ppp/watchlist-server/internal/storage"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// GET /v1/authorized/movie/:id/
func (s *Server) HandleMovieGet(c echo.Context) error {
	// bind & validate id param
	var param request.IDPathParam
	if httpError := s.bindPath(c, &param); httpError != nil {
		return httpError
	}

	// fetch movie
	movie, err := s.app.MovieGet(c.Request().Context(), param.ID)
	if err != nil {
		if err == app.ErrNotFound {
			s.logger.Info(
				"server.HandleMovieGet: movie not found",
				zap.Int("id", param.ID),
			)
			return echo.NewHTTPError(http.StatusNotFound)
		}

		s.logger.Error(
			"server.HandleMovieGet: internal server error", zap.Error(err),
		)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, movie)
}

// GET /v1/authorized/movie/?page=1&page_size=60&sort_field=id&sort_order=asc
func (s *Server) HandleMoviesGetAll(c echo.Context) error {
	// bind & validate query
	var pagQuery request.PaginationSortingQuery
	if httpError := s.bindQuery(c, pagQuery.SetValidationModel(models.TableNames.Films)); httpError != nil {
		return httpError
	}

	queryOptions := pagQuery.SetQueryIfNotSet(request.PaginationSortingQuery{
		SortingQuery: request.SortingQuery{
			SortField: models.FilmColumns.ID,
			SortOrderQuery: request.SortOrderQuery{
				SortOrder: request.SortOrderAsc,
			},
		},
		PaginationQuery: request.PaginationQuery{
			Page:     config.Config.Validation.Pagination.Page.MinValue,
			PageSize: config.Config.Validation.Pagination.PageSize.DefaultValue,
		},
	}).ToQueryOptions()

	// fetch movies
	movies, total, err := s.app.MoviesGetAll(
		c.Request().Context(),
		queryOptions,
	)
	if err != nil {
		s.logger.Error(
			"server.HandleMoviesGetAll: internal server error",
			zap.Error(err),
		)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(
		http.StatusOK,
		response.Paginated(
			pagQuery.Page,
			pagQuery.PageSize,
			movies,
			total,
		),
	)
}

// POST /v1/authorized/movie/
func (s *Server) HandleMovieCreate(c echo.Context) error {
	// bind & validate request
	var req dto.MovieCreateRequest
	if httpError := s.bindBody(c, &req); httpError != nil {
		return httpError
	}

	payload, httpError := s.getUserPayload(c)
	if httpError != nil {
		return httpError
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
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, response.ID(movieID))
}

// UPDATE /v1/authorized/movie/:id/
func (s *Server) HandleMovieUpdate(c echo.Context) error {
	// bind & validate id param
	var param request.IDPathParam
	if httpError := s.bindPath(c, &param); httpError != nil {
		return httpError
	}

	// bind & validate request
	var req dto.MovieUpdateRequest
	if httpError := s.bindBody(c, &req); httpError != nil {
		return httpError
	}

	payload, httpError := s.getUserPayload(c)
	if httpError != nil {
		return httpError
	}

	// update movie
	err := s.app.MovieUpdate(
		c.Request().Context(),
		param.ID,
		payload.UserID,
		&req,
	)
	if err != nil {
		if err == app.ErrNotFound {
			s.logger.Info(
				"server.HandleMovieUpdate: movie not found",
				zap.Int("id", param.ID),
			)
			return echo.NewHTTPError(http.StatusNotFound)
		}

		s.logger.Error(
			"server.HandleMovieUpdate: internal server error",
			zap.Error(err),
		)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

// POST /v1/authorized/movie/:id/invalidate
func (s *Server) HandleMovieInvalidate(c echo.Context) error {
	// bind & validate id param
	var param request.IDPathParam
	if httpError := s.bindPath(c, &param); httpError != nil {
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

	// invalidate movie
	err := s.app.MovieInvalidate(
		c.Request().Context(),
		param.ID,
		payload.UserID,
		&req,
	)
	if err != nil {
		if err == app.ErrNotFound {
			s.logger.Info(
				"server.HandleMovieInvalidate: movie not found",
				zap.Int("id", param.ID),
			)
			return echo.NewHTTPError(http.StatusNotFound)
		}

		s.logger.Error(
			"server.HandleMovieInvalidate: internal server error",
			zap.Error(err),
		)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

// GET /v1/authorized/movie/:id/audits?page=1&page_size=100&sort_order=desc
func (s *Server) HandleMovieAuditsGetAll(c echo.Context) error {
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
			SortOrder: request.SortOrderDesc,
		},
	}).
		ToQueryOptions()

	// fetch audits
	audits, total, err := s.app.MovieAuditsGetAll(
		c.Request().Context(),
		param.ID,
		queryOptions,
	)
	if err != nil {
		if err == app.ErrNotFound {
			s.logger.Info(
				"server.HandleMovieAuditsGetAll: movie not found",
				zap.Int("id", param.ID),
			)
			return echo.NewHTTPError(http.StatusNotFound)
		}

		s.logger.Error(
			"server.HandleMovieAuditsGetAll: internal server error",
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

// GET /v1/authorized/movie/search/?query=query&page=1&page_size=60
func (s *Server) HandleMoviesSearch(c echo.Context) error {
	// bind & validate query
	var searchPagQuery request.SearchPaginationQuery
	if httpError := s.bindQuery(c, &searchPagQuery); httpError != nil {
		return httpError
	}

	queryOptions := searchPagQuery.SetQueryIfNotSet(request.PaginationQuery{
		Page:     config.Config.Validation.Pagination.Page.MinValue,
		PageSize: config.Config.Validation.Pagination.PageSize.DefaultValue,
	}).ToQueryOptions()

	// fetch movies
	movies, total, err := s.app.MoviesSearch(
		c.Request().Context(),
		queryOptions,
	)
	if err != nil {
		s.logger.Error(
			"server.HandleMoviesSearch: internal server error",
			zap.Error(err),
		)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(
		http.StatusOK,
		response.Paginated(
			searchPagQuery.Page,
			searchPagQuery.PageSize,
			movies,
			total,
		),
	)
}

// PUT /v1/authorized/movie/:id/poster/
func (s *Server) HandleMoviePutPoster(c echo.Context) error {
	var (
		filename = config.Config.MinIO.Filename.Movie
		bucket   = config.Config.MinIO.Bucket.Image.Name
		category = config.Config.MinIO.Category.Movie
	)

	// bind & validate id param
	var param request.IDPathParam
	if httpError := s.bindPath(c, &param); httpError != nil {
		return httpError
	}

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

	// put poster
	uri, err := s.app.MoviePutPoster(
		c.Request().Context(),
		param.ID,
		payload.UserID,
		file,
		&storage.PutOptions{
			Bucket:      bucket,
			Category:    category,
			CategoryID:  param.ID,
			Filename:    filename,
			ContentType: contentType,
			Size:        fileHeader.Size,
		},
	)
	if err != nil {
		if err == app.ErrNotFound {
			s.logger.Info(
				"server.HandleMoviePutPoster: movie not found",
				zap.Int("id", param.ID),
			)
			return echo.NewHTTPError(http.StatusNotFound)
		}

		s.logger.Error(
			"server.HandleMoviePutPoster: failed putting poster",
			zap.String("bucket", bucket),
			zap.Error(err),
		)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, response.URI(uri))
}
