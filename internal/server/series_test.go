package server_test

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/aria3ppp/watchlist-server/internal/config"
	"github.com/aria3ppp/watchlist-server/internal/dto"
	"github.com/aria3ppp/watchlist-server/internal/models"
	"github.com/aria3ppp/watchlist-server/internal/query"
	"github.com/aria3ppp/watchlist-server/internal/search/searchtestutils"
	"github.com/aria3ppp/watchlist-server/internal/server/request"
	"github.com/aria3ppp/watchlist-server/internal/server/response"
	"github.com/aria3ppp/watchlist-server/internal/testutils"
	"github.com/gavv/httpexpect/v2"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	prequire "github.com/stretchr/testify/require"
	"github.com/volatiletech/null/v8"
)

func TestHandleSeriesGet(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	server, appInstance, defaults, teardown := setup(OptEnableDefaultUser)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/authorized/series/{id}"
	method := http.MethodGet

	// invalid id
	e.Request(method, path).
		WithPath("id", -1).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			validation.Errors{
				"id": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{"threshold": 1},
				),
			}.Error(),
		))

	// series not found
	e.Request(method, path).
		WithPath("id", 999).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusNotFound).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			http.StatusText(http.StatusNotFound),
		))

	createTime := time.Now()

	// add a new series
	seriesCreateReq := &dto.SeriesCreateRequest{
		Title:       "series",
		DateStarted: testutils.Date(2000, 1, 1),
	}
	seriesID, err := appInstance.SeriesCreate(
		ctx,
		defaults.user.id,
		seriesCreateReq,
	)
	require.NoError(err)

	gotSeries, err := appInstance.SeriesGet(ctx, seriesID)
	require.NoError(err)

	require.GreaterOrEqual(gotSeries.ContributedAt, createTime)

	// get series
	e.Request(method, path).
		WithPath("id", seriesID).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Equal(&models.Series{
			ID:            seriesID,
			Title:         seriesCreateReq.Title,
			Descriptions:  seriesCreateReq.Descriptions,
			DateStarted:   seriesCreateReq.DateStarted,
			DateEnded:     seriesCreateReq.DateEnded,
			Invalidation:  null.String{},
			ContributedBy: defaults.user.id,
			ContributedAt: gotSeries.ContributedAt,
		})
}

func TestHandleSeriesesGetAll(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	server, appInstance, defaults, teardown := setup(OptEnableDefaultUser)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/authorized/series"
	method := http.MethodGet

	// invalid query
	e.Request(method, path).
		WithQueryObject(request.PaginationQuery{Page: -1}).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			validation.Errors{
				"page": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{
						"threshold": config.Config.Validation.Pagination.Page.MinValue,
					},
				),
			}.Error(),
		))

	// no serieses
	e.Request(method, path).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Equal(response.Paginated(
			config.Config.Validation.Pagination.Page.MinValue,
			config.Config.Validation.Pagination.PageSize.DefaultValue,
			([]*models.Series)(nil),
			0,
		))

	// add serieses
	seriesCreateReqs := []*dto.SeriesCreateRequest{
		{
			Title:       "s1",
			DateStarted: testutils.Date(2000, 1, 1),
		},
		{
			Title:       "s2",
			DateStarted: testutils.Date(2001, 1, 1),
		},
		{
			Title:       "s3",
			DateStarted: testutils.Date(2002, 1, 1),
		},
		{
			Title:       "s4",
			DateStarted: testutils.Date(2003, 1, 1),
		},
		{
			Title:       "s5",
			DateStarted: testutils.Date(2004, 1, 1),
		},
	}
	seriesIDs := make([]int, len(seriesCreateReqs))

	createTime := time.Now()

	for i, req := range seriesCreateReqs {
		var err error
		seriesIDs[i], err = appInstance.SeriesCreate(
			ctx,
			defaults.user.id,
			req,
		)
		require.NoError(err)
	}

	gotSerieses, total, err := appInstance.SeriesesGetAll(ctx, query.Options{
		Offset:    0,
		Limit:     config.Config.Validation.Pagination.PageSize.MaxValue,
		SortField: models.SeriesColumns.ID,
		SortOrder: request.SortOrderAsc,
	})
	require.NoError(err)

	items := make([]*models.Series, len(gotSerieses))

	for i := range gotSerieses {
		require.GreaterOrEqual(gotSerieses[i].ContributedAt, createTime)

		items[i] = &models.Series{
			ID:            seriesIDs[i],
			Title:         seriesCreateReqs[i].Title,
			Descriptions:  seriesCreateReqs[i].Descriptions,
			DateStarted:   seriesCreateReqs[i].DateStarted,
			DateEnded:     seriesCreateReqs[i].DateEnded,
			Invalidation:  null.String{},
			ContributedBy: defaults.user.id,
			ContributedAt: gotSerieses[i].ContributedAt,
		}
	}

	// get all series
	e.Request(method, path).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Equal(response.Paginated(
			config.Config.Validation.Pagination.Page.MinValue,
			config.Config.Validation.Pagination.PageSize.DefaultValue,
			items,
			total,
		))
}

func TestHandleSeriesCreate(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	server, appInstance, defaults, teardown := setup(OptEnableDefaultUser)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/authorized/series"
	method := http.MethodPost

	// invalid request
	e.Request(method, path).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(dto.SeriesCreateRequest{}).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			validation.Errors{
				"title":        validation.ErrRequired,
				"date_started": validation.ErrRequired,
			}.Error(),
		))

	// create series
	seriesCreateReq := &dto.SeriesCreateRequest{
		Title:       "series",
		DateStarted: testutils.Date(2000, 1, 1),
	}

	createTime := time.Now()

	rawSeriesID := e.Request(method, path).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(seriesCreateReq).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("id").Number().Ge(0).Raw()

	seriesID := int(rawSeriesID)

	// check series created

	gotSeries, err := appInstance.SeriesGet(ctx, seriesID)
	require.NoError(err)

	testutils.SetTimeLocation(
		&seriesCreateReq.DateStarted,
		gotSeries.DateStarted.Location(),
	)
	testutils.SetTimeLocation(
		&seriesCreateReq.DateEnded.Time,
		gotSeries.DateEnded.Time.Location(),
	)

	require.GreaterOrEqual(gotSeries.ContributedAt, createTime)

	require.Equal(
		&models.Series{
			ID:            seriesID,
			Title:         seriesCreateReq.Title,
			Descriptions:  seriesCreateReq.Descriptions,
			DateStarted:   seriesCreateReq.DateStarted,
			DateEnded:     seriesCreateReq.DateEnded,
			Invalidation:  null.String{},
			ContributedBy: defaults.user.id,
			ContributedAt: gotSeries.ContributedAt,
		},
		gotSeries,
	)
}

func TestHandleSeriesUpdate(t *testing.T) {
	ctx := context.Background()

	server, appInstance, defaults, teardown := setup(OptEnableDefaultUser)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/authorized/series/{id}"
	method := http.MethodPatch

	// invalid id
	e.Request(method, path).
		WithPath("id", -1).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(&dto.SeriesUpdateRequest{}).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			validation.Errors{
				"id": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{"threshold": 1},
				),
			}.Error(),
		))

	// invalid request
	e.Request(method, path).
		WithPath("id", 1).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(&dto.SeriesUpdateRequest{Title: null.StringFrom("t")}).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			validation.Errors{
				"title": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.Series.Title.MinLength,
						"max": config.Config.Validation.Series.Title.MaxLength,
					},
				),
			}.Error(),
		))

	// series not found
	e.Request(method, path).
		WithPath("id", 9999999).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(&dto.SeriesUpdateRequest{}).
		Expect().
		Status(http.StatusNotFound).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			http.StatusText(http.StatusNotFound),
		))

	updates := []struct {
		name string
		req  *dto.SeriesUpdateRequest
	}{
		{
			name: "u1",
			req:  &dto.SeriesUpdateRequest{},
		},
		{
			name: "u2",
			req: &dto.SeriesUpdateRequest{
				Title: null.StringFrom("updated_title"),
			},
		},
		{
			name: "u3",
			req: &dto.SeriesUpdateRequest{
				Descriptions: null.StringFrom("updated_description"),
			},
		},
		{
			name: "u4",
			req: &dto.SeriesUpdateRequest{
				DateStarted: null.TimeFrom(testutils.Date(2000, 1, 1)),
			},
		},
		{
			name: "u5",
			req: &dto.SeriesUpdateRequest{
				DateEnded: null.TimeFrom(testutils.Date(2008, 8, 11)),
			},
		},
		{
			name: "u6",
			req: &dto.SeriesUpdateRequest{
				Title:        null.StringFrom("updated_title"),
				Descriptions: null.StringFrom("updated_description"),
				DateStarted:  null.TimeFrom(testutils.Date(2000, 1, 1)),
				DateEnded:    null.TimeFrom(testutils.Date(2008, 8, 11)),
			},
		},
	}

	for i, u := range updates {
		u := u
		i := i
		t.Run(u.name, func(t *testing.T) {
			require := prequire.New(t)

			// insert series
			seriesID, err := appInstance.SeriesCreate(
				ctx,
				defaults.user.id,
				&dto.SeriesCreateRequest{
					Title:       "series" + strconv.Itoa(i),
					DateStarted: testutils.Date(1900, 3, 14),
				},
			)
			require.NoError(err)

			gotSeriesBeforeUpdate, err := appInstance.SeriesGet(ctx, seriesID)
			require.NoError(err)

			updateTime := time.Now()

			// update
			e.Request(method, path).
				WithPath("id", seriesID).
				WithHeader(echo.HeaderAuthorization, defaults.user.auth).
				WithJSON(u.req).
				Expect().
				Status(http.StatusOK).
				NoContent()

			// check updated fields
			gotSeriesAfterUpdate, err := appInstance.SeriesGet(ctx, seriesID)
			require.NoError(err)

			require.GreaterOrEqual(
				gotSeriesAfterUpdate.ContributedAt,
				updateTime,
			)

			updatedSeries := &models.Series{}
			updatedSeries.ID = seriesID
			if u.req.Title.Valid {
				updatedSeries.Title = u.req.Title.String
			} else {
				updatedSeries.Title = gotSeriesBeforeUpdate.Title
			}
			if u.req.Descriptions.Valid {
				updatedSeries.Descriptions = u.req.Descriptions
			} else {
				updatedSeries.Descriptions = gotSeriesBeforeUpdate.Descriptions
			}
			if u.req.DateStarted.Valid {
				updatedSeries.DateStarted = u.req.DateStarted.Time
			} else {
				updatedSeries.DateStarted = gotSeriesBeforeUpdate.DateStarted
			}
			if u.req.DateEnded.Valid {
				updatedSeries.DateEnded = u.req.DateEnded
			} else {
				updatedSeries.DateEnded = gotSeriesBeforeUpdate.DateEnded
			}
			updatedSeries.Invalidation = null.String{}
			updatedSeries.ContributedBy = defaults.user.id
			updatedSeries.ContributedAt = gotSeriesAfterUpdate.ContributedAt

			testutils.SetTimeLocation(
				&updatedSeries.DateStarted,
				gotSeriesAfterUpdate.DateStarted.Location(),
			)
			testutils.SetTimeLocation(
				&updatedSeries.DateEnded.Time,
				gotSeriesAfterUpdate.DateEnded.Time.Location(),
			)

			require.Equal(updatedSeries, gotSeriesAfterUpdate)
		})
	}
}

func TestHandleSeriesInvalidate(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	server, appInstance, defaults, teardown := setup(OptEnableDefaultUser)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/authorized/series/{id}/invalidate"
	method := http.MethodPost

	invalidationRequest := &dto.InvalidationRequest{
		Invalidation: "invalidation",
	}

	// invalid id
	e.Request(method, path).
		WithPath("id", -1).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(invalidationRequest).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			validation.Errors{
				"id": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{"threshold": 1},
				),
			}.Error(),
		))

	// invalid request
	e.Request(method, path).
		WithPath("id", 1).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(dto.InvalidationRequest{}).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			validation.Errors{
				"invalidation": validation.ErrRequired,
			}.Error(),
		))

	// series not found
	e.Request(method, path).
		WithPath("id", 999).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(invalidationRequest).
		Expect().
		Status(http.StatusNotFound).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			http.StatusText(http.StatusNotFound),
		))

	// invalidate series
	seriesCreateReq := &dto.SeriesCreateRequest{
		Title:       "series",
		DateStarted: testutils.Date(1900, 3, 14),
	}
	seriesID, err := appInstance.SeriesCreate(
		ctx,
		defaults.user.id,
		seriesCreateReq,
	)
	require.NoError(err)

	e.Request(method, path).
		WithPath("id", seriesID).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(invalidationRequest).
		Expect().
		Status(http.StatusOK).
		NoContent()

	// check series invalidated
	gotInvalidatedSeries, err := appInstance.SeriesGet(ctx, seriesID)
	require.NoError(err)

	testutils.SetTimeLocation(
		&seriesCreateReq.DateStarted,
		gotInvalidatedSeries.DateStarted.Location(),
	)
	testutils.SetTimeLocation(
		&seriesCreateReq.DateEnded.Time,
		gotInvalidatedSeries.DateEnded.Time.Location(),
	)

	require.Equal(
		&models.Series{
			ID:            seriesID,
			Title:         seriesCreateReq.Title,
			Descriptions:  seriesCreateReq.Descriptions,
			DateStarted:   seriesCreateReq.DateStarted,
			DateEnded:     seriesCreateReq.DateEnded,
			Invalidation:  null.StringFrom(invalidationRequest.Invalidation),
			ContributedBy: defaults.user.id,
			ContributedAt: gotInvalidatedSeries.ContributedAt,
		},
		gotInvalidatedSeries,
	)
}

func TestHandleSeriesAuditsGetAll(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	server, appInstance, defaults, teardown := setup(OptEnableDefaultUser)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/authorized/series/{id}/audits"
	method := http.MethodGet

	// invalid id
	e.Request(method, path).
		WithPath("id", -1).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			validation.Errors{
				"id": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{"threshold": 1},
				),
			}.Error(),
		))

	// invalid query
	e.Request(method, path).
		WithPath("id", 1).
		WithQueryObject(request.PaginationQuery{Page: -1}).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			validation.Errors{
				"page": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{
						"threshold": config.Config.Validation.Pagination.Page.MinValue,
					},
				),
			}.Error(),
		))

	// series not found
	e.Request(method, path).
		WithPath("id", 999).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusNotFound).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			http.StatusText(http.StatusNotFound),
		))

	createTime := time.Now()

	// add series
	seriesCreateReq := &dto.SeriesCreateRequest{
		Title:       "series",
		DateStarted: testutils.Date(2000, 1, 1),
	}
	seriesID, err := appInstance.SeriesCreate(
		ctx,
		defaults.user.id,
		seriesCreateReq,
	)
	require.NoError(err)

	gotSeries, err := appInstance.SeriesGet(ctx, seriesID)
	require.NoError(err)

	require.GreaterOrEqual(gotSeries.ContributedAt, createTime)

	expSeriesUpdateAudit := &models.SeriesesAudit{
		ID:            seriesID,
		Title:         seriesCreateReq.Title,
		Descriptions:  seriesCreateReq.Descriptions,
		DateStarted:   seriesCreateReq.DateStarted,
		DateEnded:     seriesCreateReq.DateEnded,
		Invalidation:  null.String{},
		ContributedBy: defaults.user.id,
		ContributedAt: gotSeries.ContributedAt,
	}

	// no audits
	e.Request(method, path).
		WithPath("id", seriesID).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Equal(response.Paginated(
			config.Config.Validation.Pagination.Page.MinValue,
			config.Config.Validation.Pagination.PageSize.DefaultValue,
			([]*models.SeriesesAudit)(nil),
			0,
		))

	updateTime := time.Now()

	// update the series
	seriesUpdateReq := &dto.SeriesUpdateRequest{
		Title:        null.StringFrom("updated title"),
		Descriptions: null.StringFrom("updated descriptions"),
		DateStarted: null.TimeFrom(
			testutils.Date(2005, 11, 14),
		),
		DateEnded: null.TimeFrom(
			testutils.Date(2015, 3, 26),
		),
	}
	err = appInstance.SeriesUpdate(
		ctx,
		seriesID,
		defaults.user.id,
		seriesUpdateReq,
	)
	require.NoError(err)

	gotSeries, err = appInstance.SeriesGet(ctx, seriesID)
	require.NoError(err)

	require.GreaterOrEqual(gotSeries.ContributedAt, updateTime)

	expSeriesInvalidationAudit := &models.SeriesesAudit{
		ID:            seriesID,
		Title:         seriesUpdateReq.Title.String,
		Descriptions:  seriesUpdateReq.Descriptions,
		DateStarted:   seriesUpdateReq.DateStarted.Time,
		DateEnded:     seriesUpdateReq.DateEnded,
		Invalidation:  null.String{},
		ContributedBy: defaults.user.id,
		ContributedAt: gotSeries.ContributedAt,
	}

	// get update audits
	e.Request(method, path).
		WithPath("id", seriesID).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Equal(response.Paginated(
			config.Config.Validation.Pagination.Page.MinValue,
			config.Config.Validation.Pagination.PageSize.DefaultValue,
			[]*models.SeriesesAudit{expSeriesUpdateAudit},
			1,
		))

	// invalidate the series
	err = appInstance.SeriesInvalidate(
		ctx,
		seriesID,
		defaults.user.id,
		&dto.InvalidationRequest{Invalidation: "invalidation"},
	)
	require.NoError(err)

	// get invalidation audits
	e.Request(method, path).
		WithPath("id", seriesID).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Equal(response.Paginated(
			config.Config.Validation.Pagination.Page.MinValue,
			config.Config.Validation.Pagination.PageSize.DefaultValue,
			[]*models.SeriesesAudit{
				expSeriesInvalidationAudit,
				expSeriesUpdateAudit,
			},
			2,
		))
}

func TestHandleSeriesesSearch(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	server, appInstance, defaults, teardown := setup(OptEnableDefaultUser)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/authorized/series/search"
	method := http.MethodGet

	// invalid query
	e.Request(method, path).
		WithQueryObject(request.SearchPaginationQuery{}).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			validation.Errors{
				"query": validation.ErrRequired,
			}.Error(),
		))

	// no serieses
	e.Request(method, path).
		WithQuery("query", "query").
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Equal(response.Paginated(
			config.Config.Validation.Pagination.Page.MinValue,
			config.Config.Validation.Pagination.PageSize.DefaultValue,
			[]*models.Series{},
			0,
		))

	// add serieses

	querySerieses := []*dto.SeriesCreateRequest{
		{Title: "query"},
		{Title: "title before query"},
		{Title: "query before title"},
		{Title: "title then query then title"},

		{Title: "QUERY"},
		{Title: "title before QUERY"},
		{Title: "QUERY before title"},
		{Title: "title then QUERY then title"},

		{Title: "uery"},
		{Title: "title before uery"},
		{Title: "uery before title"},
		{Title: "title then uery then title"},

		{Title: "qury"},
		{Title: "title before qury"},
		{Title: "qury before title"},
		{Title: "title then qury then title"},

		{Title: "quer"},
		{Title: "title before quer"},
		{Title: "quer before title"},
		{Title: "title then quer then title"},

		{Title: "Xuery"},
		{Title: "title before Xuery"},
		{Title: "Xuery before title"},
		{Title: "title then Xuery then title"},

		{Title: "quXry"},
		{Title: "title before quXry"},
		{Title: "quXry before title"},
		{Title: "title then quXry then title"},

		{Title: "querX"},
		{Title: "title before querX"},
		{Title: "querX before title"},
		{Title: "title then Xuery then title"},

		{Descriptions: null.StringFrom("query")},
		{Descriptions: null.StringFrom("descriptions before query")},
		{Descriptions: null.StringFrom("query before descriptions")},
		{Descriptions: null.StringFrom(
			"descriptions then query then descriptions",
		)},

		{Descriptions: null.StringFrom("QUERY")},
		{Descriptions: null.StringFrom("descriptions before QUERY")},
		{Descriptions: null.StringFrom("QUERY before descriptions")},
		{Descriptions: null.StringFrom(
			"descriptions then QUERY then descriptions",
		)},

		{Descriptions: null.StringFrom("uery")},
		{Descriptions: null.StringFrom("descriptions before uery")},
		{Descriptions: null.StringFrom("uery before descriptions")},
		{Descriptions: null.StringFrom(
			"descriptions then uery then descriptions",
		)},

		{Descriptions: null.StringFrom("qury")},
		{Descriptions: null.StringFrom("descriptions before qury")},
		{Descriptions: null.StringFrom("qury before descriptions")},
		{Descriptions: null.StringFrom(
			"descriptions then qury then descriptions",
		)},

		{Descriptions: null.StringFrom("quer")},
		{Descriptions: null.StringFrom("descriptions before quer")},
		{Descriptions: null.StringFrom("quer before descriptions")},
		{Descriptions: null.StringFrom(
			"descriptions then quer then descriptions",
		)},

		{Descriptions: null.StringFrom("Xuery")},
		{Descriptions: null.StringFrom("descriptions before Xuery")},
		{Descriptions: null.StringFrom("Xuery before descriptions")},
		{Descriptions: null.StringFrom(
			"descriptions then Xuery then descriptions",
		)},

		{Descriptions: null.StringFrom("quXry")},
		{Descriptions: null.StringFrom("descriptions before quXry")},
		{Descriptions: null.StringFrom("quXry before descriptions")},
		{Descriptions: null.StringFrom(
			"descriptions then quXry then descriptions",
		)},

		{Descriptions: null.StringFrom("querX")},
		{Descriptions: null.StringFrom("descriptions before querX")},
		{Descriptions: null.StringFrom("querX before descriptions")},
		{Descriptions: null.StringFrom(
			"descriptions then Xuery then descriptions",
		)},
	}
	// interleave querySerieses
	seriesCreateReqs := make([]*dto.SeriesCreateRequest, len(querySerieses)*2)
	for i := range seriesCreateReqs {
		var series *dto.SeriesCreateRequest
		if i%2 == 0 {
			qi := 0
			if i != 0 {
				qi = i / 2
			}
			series = querySerieses[qi]
		} else {
			series = &dto.SeriesCreateRequest{
				Title:        "title_ignore_me",
				Descriptions: null.StringFrom("descriptions"),
			}
		}
		seriesCreateReqs[i] = series
	}

	seriesIDs := make([]int, len(seriesCreateReqs))

	createTime := time.Now()

	var thePosterUri string

	for i, req := range seriesCreateReqs {
		// set DateStarted field to skip request validation error
		req.DateStarted = testutils.Date(2000, 1, 1)

		var err error
		seriesIDs[i], err = appInstance.SeriesCreate(
			ctx,
			defaults.user.id,
			req,
		)
		require.NoError(err)

		// set poster just for the first one
		if i == 0 {
			thePosterUri = e.Request(http.MethodPut, "/v1/authorized/series/{id}/poster").
				WithPath("id", seriesIDs[i]).
				WithHeader(echo.HeaderAuthorization, defaults.user.auth).
				WithMultipart().
				WithFile("poster", filepath.Join("testdata", "gopher-1.webp")).
				Expect().
				Status(http.StatusOK).
				JSON().
				Object().
				Value("uri").
				String().
				NotEmpty().
				Raw()
		}
	}

	// wait until series got index
	err := testutils.WaitUntil(
		func() (bool, error) {
			c, err := searchtestutils.CountIndex(
				esClient,
				config.Config.Elasticsearch.Index.Serieses,
			)
			if err != nil {
				return false, err
			}
			return c == len(seriesCreateReqs), nil
		},
		10*time.Second,
		time.Second,
	)
	require.NoError(err)

	gotSerieses, _, err := appInstance.SeriesesGetAll(
		ctx, query.Options{
			Offset:    0,
			Limit:     config.Config.Validation.Pagination.PageSize.MaxValue,
			SortField: models.SeriesColumns.ID,
			SortOrder: request.SortOrderAsc,
		})

	require.NoError(err)

	items := make([]*models.Series, len(querySerieses))

	itemsIndex := 0
	for i := range gotSerieses {
		if gotSerieses[i].Title != "title_ignore_me" {

			var poster null.String
			if i == 0 {
				poster = null.StringFrom(thePosterUri)
			}

			require.GreaterOrEqual(gotSerieses[i].ContributedAt, createTime)

			items[itemsIndex] = &models.Series{
				ID:            seriesIDs[i],
				Title:         seriesCreateReqs[i].Title,
				Descriptions:  seriesCreateReqs[i].Descriptions,
				DateStarted:   seriesCreateReqs[i].DateStarted,
				DateEnded:     seriesCreateReqs[i].DateEnded,
				Poster:        poster,
				Invalidation:  null.String{},
				ContributedBy: defaults.user.id,
				ContributedAt: gotSerieses[i].ContributedAt,
			}

			itemsIndex++

		}
	}

	// search series
	responseObject := e.Request(method, path).
		WithQuery("query", "query").
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object()

	responseItems := responseObject.
		ValueEqual("page", config.Config.Validation.Pagination.Page.MinValue).
		ValueEqual("page_size", config.Config.Validation.Pagination.PageSize.DefaultValue).
		ValueEqual("total_pages", (len(items)+config.Config.Validation.Pagination.PageSize.DefaultValue-1)/config.Config.Validation.Pagination.PageSize.DefaultValue).
		ValueEqual("total_items", len(items)).
		Value("items").Array()

	toAnySlice := func(items []*models.Series) []any {
		anyItems := make([]any, len(items))
		for i, it := range items {
			anyItems[i] = it
		}
		return anyItems
	}

	responseItems.Length().Equal(len(items))
	responseItems.Contains(toAnySlice(items)...)
}

func TestHandleSeriesPutPoster(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	server, appInstance, defaults, teardown := setup(OptEnableDefaultUser)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/authorized/series/{id}/poster"
	method := http.MethodPut

	filename := config.Config.MinIO.Filename.Series

	type FileInfo struct {
		path        string
		file        *os.File
		size        int64
		contentType string
	}
	openFile := func(path string) FileInfo {
		file, err := os.Open(path)
		require.NoError(err)
		t.Cleanup(func() { file.Close() })

		stat, err := file.Stat()
		require.NoError(err)

		buf := make([]byte, 512)
		_, err = file.Read(buf)
		require.NoError(err)
		_, err = file.Seek(0, 0)
		require.NoError(err)

		return FileInfo{
			path:        path,
			file:        file,
			size:        stat.Size(),
			contentType: http.DetectContentType(buf),
		}
	}

	file := openFile(filepath.Join("testdata", "gopher-1.webp"))
	overwrittenFile := openFile(filepath.Join("testdata", "gopher-2.png"))
	unsupportedFile := openFile(filepath.Join("testdata", "gopher-3.ico"))

	// invalid id
	e.Request(method, path).
		WithPath("id", -1).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithMultipart().WithFile(filename, file.path).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			validation.Errors{
				"id": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{"threshold": 1},
				),
			}.Error(),
		))

	// series not found
	e.Request(method, path).
		WithPath("id", 999).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithMultipart().WithFile(filename, file.path).
		Expect().
		Status(http.StatusNotFound).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			http.StatusText(http.StatusNotFound),
		))

	seriesCreateReq := &dto.SeriesCreateRequest{
		Title:       "series",
		DateStarted: testutils.Date(1900, 3, 14),
	}
	seriesID, err := appInstance.SeriesCreate(
		ctx,
		defaults.user.id,
		seriesCreateReq,
	)
	require.NoError(err)

	// series have no poster

	seriesGetReq := struct {
		method string
		path   string
	}{
		method: http.MethodGet,
		path:   "/v1/authorized/series/{id}",
	}

	e.Request(seriesGetReq.method, seriesGetReq.path).
		WithPath("id", seriesID).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("poster").Null()

	// missing file
	e.Request(method, path).
		WithPath("id", seriesID).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithMultipart().
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			"missing file",
		))

	// unsupported file
	e.Request(method, path).
		WithPath("id", seriesID).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithMultipart().WithFile(filename, unsupportedFile.path).
		Expect().
		Status(http.StatusUnsupportedMediaType).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			"unsupported file type",
		))

	// put poster
	uri := e.Request(method, path).
		WithPath("id", seriesID).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithMultipart().WithFile(filename, file.path).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("uri").String().NotEmpty().Raw()

	// now series have poster
	e.Request(seriesGetReq.method, seriesGetReq.path).
		WithPath("id", seriesID).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		ValueEqual("poster", uri)

	// check file on storage

	storageExpect := httpexpect.New(t, "http://127.0.0.1:9000")

	uriSplit := strings.Split(uri, "?")
	require.Equal(2, len(uriSplit))
	versionIdQuery := strings.Split(uriSplit[1], "=")
	require.Equal(2, len(versionIdQuery))

	resp := storageExpect.Request(http.MethodGet, uriSplit[0]).
		WithQuery(versionIdQuery[0], versionIdQuery[1]).
		Expect().
		Status(http.StatusOK)

	resp.Header("Content-Length").Equal(strconv.FormatInt(file.size, 10))
	resp.Header("Content-Type").Equal(file.contentType)

	// overwrite poster
	uri = e.Request(method, path).
		WithPath("id", seriesID).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithMultipart().WithFile(filename, overwrittenFile.path).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("uri").String().NotEmpty().Raw()

	// check overwritten series poster
	e.Request(seriesGetReq.method, seriesGetReq.path).
		WithPath("id", seriesID).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		ValueEqual("poster", uri)

	// check overwritten file on storage
	uriSplit = strings.Split(uri, "?")
	require.Equal(2, len(uriSplit))
	versionIdQuery = strings.Split(uriSplit[1], "=")
	require.Equal(2, len(versionIdQuery))

	resp = storageExpect.Request(http.MethodGet, uriSplit[0]).
		WithQuery(versionIdQuery[0], versionIdQuery[1]).
		Expect().
		Status(http.StatusOK)

	resp.Header("Content-Length").
		Equal(strconv.FormatInt(overwrittenFile.size, 10))
	resp.Header("Content-Type").Equal(overwrittenFile.contentType)
}
