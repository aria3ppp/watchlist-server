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

// GET /v1/authorized/series/:id/
func (s *Server) HandleSeriesGet(c echo.Context) error {
	// bind & validate id param
	var param request.IDPathParam
	if httpError := s.bindPath(c, &param); httpError != nil {
		return httpError
	}

	// fetch series
	series, err := s.app.SeriesGet(c.Request().Context(), param.ID)
	if err != nil {
		if err == app.ErrNotFound {
			s.logger.Info(
				"server.HandleSeriesGet: series not found",
				zap.Int("id", param.ID),
			)
			return echo.NewHTTPError(http.StatusNotFound)
		}

		s.logger.Error(
			"server.HandleSeriesGet: internal server error", zap.Error(err),
		)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, series)
}

// GET /v1/authorized/series/?page=1&page_size=60&sort_field=id&sort_order=desc
func (s *Server) HandleSeriesesGetAll(c echo.Context) error {
	// bind & validate query
	var pagQuery request.PaginationSortingQuery
	if httpError := s.bindQuery(c, pagQuery.SetValidationModel(models.TableNames.Serieses)); httpError != nil {
		return httpError
	}

	queryOptions := pagQuery.SetQueryIfNotSet(request.PaginationSortingQuery{
		SortingQuery: request.SortingQuery{
			SortField: models.SeriesColumns.ID,
			SortOrderQuery: request.SortOrderQuery{
				SortOrder: request.SortOrderAsc,
			},
		},
		PaginationQuery: request.PaginationQuery{
			Page:     config.Config.Validation.Pagination.Page.MinValue,
			PageSize: config.Config.Validation.Pagination.PageSize.DefaultValue,
		},
	}).ToQueryOptions()

	// fetch serieses
	serieses, total, err := s.app.SeriesesGetAll(
		c.Request().Context(),
		queryOptions,
	)
	if err != nil {
		s.logger.Error(
			"server.HandleSeriesesGetAll: internal server error",
			zap.Error(err),
		)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(
		http.StatusOK,
		response.Paginated(
			pagQuery.Page,
			pagQuery.PageSize,
			serieses,
			total,
		),
	)
}

// POST /v1/authorized/series/
func (s *Server) HandleSeriesCreate(c echo.Context) error {
	// bind & validate request
	var req dto.SeriesCreateRequest
	if httpError := s.bindBody(c, &req); httpError != nil {
		return httpError
	}

	payload, httpError := s.getUserPayload(c)
	if httpError != nil {
		return httpError
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
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, response.ID(seriesID))
}

// PATCH /v1/authorized/series/:id/
func (s *Server) HandleSeriesUpdate(c echo.Context) error {
	// bind & validate id param
	var param request.IDPathParam
	if httpError := s.bindPath(c, &param); httpError != nil {
		return httpError
	}

	// bind & validate request
	var req dto.SeriesUpdateRequest
	if httpError := s.bindBody(c, &req); httpError != nil {
		return httpError
	}

	payload, httpError := s.getUserPayload(c)
	if httpError != nil {
		return httpError
	}

	// update series
	err := s.app.SeriesUpdate(
		c.Request().Context(),
		param.ID,
		payload.UserID,
		&req,
	)
	if err != nil {
		if err == app.ErrNotFound {
			s.logger.Info(
				"server.HandleSeriesUpdate: series not found",
				zap.Int("id", param.ID),
			)
			return echo.NewHTTPError(http.StatusNotFound)
		}

		s.logger.Error(
			"server.HandleSeriesUpdate: internal server error",
			zap.Error(err),
		)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

// POST /v1/authorized/series/:id/invalidate
func (s *Server) HandleSeriesInvalidate(c echo.Context) error {
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

	// invalidate series
	err := s.app.SeriesInvalidate(
		c.Request().Context(),
		param.ID,
		payload.UserID,
		&req,
	)
	if err != nil {
		if err == app.ErrNotFound {
			s.logger.Info(
				"server.HandleSeriesInvalidate: series not found",
				zap.Int("id", param.ID),
			)
			return echo.NewHTTPError(http.StatusNotFound)
		}

		s.logger.Error(
			"server.HandleSeriesInvalidate: internal server error",
			zap.Error(err),
		)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

// GET /v1/authorized/series/:id/audits/?page=1&page_size=60&sort_order=desc
func (s *Server) HandleSeriesAuditsGetAll(c echo.Context) error {
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
	}).ToQueryOptions()

	// fetch audits
	audits, total, err := s.app.SeriesAuditsGetAll(
		c.Request().Context(),
		param.ID,
		queryOptions,
	)
	if err != nil {
		if err == app.ErrNotFound {
			s.logger.Info(
				"server.HandleSeriesAuditsGetAll: series not found",
				zap.Int("id", param.ID),
			)
			return echo.NewHTTPError(http.StatusNotFound)
		}

		s.logger.Error(
			"server.HandleSeriesAuditsGetAll: internal server error",
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

// GET /v1/authorized/series/search/?query=query&page=1&page_size=60
func (s *Server) HandleSeriesesSearch(c echo.Context) error {
	// bind & validate query
	var searchPagQuery request.SearchPaginationQuery
	if httpError := s.bindQuery(c, &searchPagQuery); httpError != nil {
		return httpError
	}

	queryOptions := searchPagQuery.SetQueryIfNotSet(request.PaginationQuery{
		Page:     config.Config.Validation.Pagination.Page.MinValue,
		PageSize: config.Config.Validation.Pagination.PageSize.DefaultValue,
	}).ToQueryOptions()

	// fetch serieses
	serieses, total, err := s.app.SeriesesSearch(
		c.Request().Context(),
		queryOptions,
	)
	if err != nil {
		s.logger.Error(
			"server.HandleSeriesesSearch: internal server error",
			zap.Error(err),
		)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(
		http.StatusOK,
		response.Paginated(
			searchPagQuery.Page,
			searchPagQuery.PageSize,
			serieses,
			total,
		),
	)
}

// PUT /v1/authorized/series/:id/poster/
func (s *Server) HandleSeriesPutPoster(c echo.Context) error {
	var (
		filename = config.Config.MinIO.Filename.Series
		bucket   = config.Config.MinIO.Bucket.Image.Name
		category = config.Config.MinIO.Category.Series
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
	uri, err := s.app.SeriesPutPoster(
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
			s.logger.Error(
				"server.HandleSeriesPutPoster: series not found",
				zap.Int("id", param.ID),
			)
			return echo.NewHTTPError(http.StatusNotFound)
		}

		s.logger.Error(
			"server.HandleSeriesPutPoster: failed putting poster",
			zap.String("bucket", bucket),
			zap.Error(err),
		)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, response.URI(uri))
}
