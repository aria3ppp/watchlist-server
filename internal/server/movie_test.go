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

func TestHandleMovieGet(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	server, appInstance, defaults, teardown := setup(OptEnableDefaultUser)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/authorized/movie/{id}"
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

	// movie not found
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

	// add a new movie
	movieCreateReq := &dto.MovieCreateRequest{
		Title:        "movie",
		DateReleased: testutils.Date(2000, 1, 1),
	}
	movieID, err := appInstance.MovieCreate(
		ctx,
		defaults.user.id,
		movieCreateReq,
	)
	require.NoError(err)

	gotMovie, err := appInstance.MovieGet(ctx, movieID)
	require.NoError(err)

	require.GreaterOrEqual(gotMovie.ContributedAt, createTime)

	// get movie
	e.Request(method, path).
		WithPath("id", movieID).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Equal(&models.Film{
			ID:            movieID,
			Title:         movieCreateReq.Title,
			Descriptions:  movieCreateReq.Descriptions,
			DateReleased:  movieCreateReq.DateReleased,
			Duration:      movieCreateReq.Duration,
			SeriesID:      null.Int{},
			SeasonNumber:  null.Int{},
			EpisodeNumber: null.Int{},
			Invalidation:  null.String{},
			ContributedBy: defaults.user.id,
			ContributedAt: gotMovie.ContributedAt,
		})
}

func TestHandleMoviesGetAll(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	server, appInstance, defaults, teardown := setup(OptEnableDefaultUser)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/authorized/movie"
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

	// no movie
	e.Request(method, path).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Equal(response.Paginated(
			config.Config.Validation.Pagination.Page.MinValue,
			config.Config.Validation.Pagination.PageSize.DefaultValue,
			([]*models.Film)(nil),
			0,
		))

	// add movies
	movieCreateReqs := []*dto.MovieCreateRequest{
		{
			Title:        "m1",
			DateReleased: testutils.Date(2000, 1, 1),
		},
		{
			Title:        "m2",
			DateReleased: testutils.Date(2001, 1, 1),
		},
		{
			Title:        "m3",
			DateReleased: testutils.Date(2002, 1, 1),
		},
		{
			Title:        "m4",
			DateReleased: testutils.Date(2003, 1, 1),
		},
		{
			Title:        "m5",
			DateReleased: testutils.Date(2004, 1, 1),
		},
	}
	movieIDs := make([]int, len(movieCreateReqs))

	createTime := time.Now()

	for i, req := range movieCreateReqs {
		var err error
		movieIDs[i], err = appInstance.MovieCreate(
			ctx,
			defaults.user.id,
			req,
		)
		require.NoError(err)
	}

	gotMovies, total, err := appInstance.MoviesGetAll(ctx, query.Options{
		Offset:    0,
		Limit:     config.Config.Validation.Pagination.PageSize.MaxValue,
		SortField: models.FilmColumns.ID,
		SortOrder: request.SortOrderAsc,
	})
	require.NoError(err)

	items := make([]*models.Film, len(gotMovies))

	for i := range gotMovies {
		require.GreaterOrEqual(gotMovies[i].ContributedAt, createTime)

		items[i] = &models.Film{
			ID:            movieIDs[i],
			Title:         movieCreateReqs[i].Title,
			Descriptions:  movieCreateReqs[i].Descriptions,
			DateReleased:  movieCreateReqs[i].DateReleased,
			Duration:      movieCreateReqs[i].Duration,
			SeriesID:      null.Int{},
			SeasonNumber:  null.Int{},
			EpisodeNumber: null.Int{},
			Invalidation:  null.String{},
			ContributedBy: defaults.user.id,
			ContributedAt: gotMovies[i].ContributedAt,
		}
	}

	// get all movies
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

func TestHandleMovieCreate(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	server, appInstance, defaults, teardown := setup(OptEnableDefaultUser)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/authorized/movie"
	method := http.MethodPost

	// invalid request
	e.Request(method, path).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(dto.MovieCreateRequest{}).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			validation.Errors{
				"title":         validation.ErrRequired,
				"date_released": validation.ErrRequired,
			}.Error(),
		))

	// create movie
	movieCreateReq := &dto.MovieCreateRequest{
		Title:        "movie",
		DateReleased: testutils.Date(2000, 1, 1),
	}

	createTime := time.Now()

	rawMovieID := e.Request(method, path).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(movieCreateReq).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("id").Number().Ge(0).Raw()

	movieID := int(rawMovieID)

	// check movie created

	gotMovie, err := appInstance.MovieGet(ctx, movieID)
	require.NoError(err)

	testutils.SetTimeLocation(
		&movieCreateReq.DateReleased,
		gotMovie.DateReleased.Location(),
	)

	require.GreaterOrEqual(gotMovie.ContributedAt, createTime)

	require.Equal(
		&models.Film{
			ID:            movieID,
			Title:         movieCreateReq.Title,
			Descriptions:  movieCreateReq.Descriptions,
			DateReleased:  movieCreateReq.DateReleased,
			Duration:      movieCreateReq.Duration,
			SeriesID:      null.Int{},
			SeasonNumber:  null.Int{},
			EpisodeNumber: null.Int{},
			Invalidation:  null.String{},
			ContributedBy: defaults.user.id,
			ContributedAt: gotMovie.ContributedAt,
		},
		gotMovie,
	)
}

func TestHandleMovieUpdate(t *testing.T) {
	ctx := context.Background()

	server, appInstance, defaults, teardown := setup(OptEnableDefaultUser)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/authorized/movie/{id}"
	method := http.MethodPatch

	// invalid id
	e.Request(method, path).
		WithPath("id", -1).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(&dto.MovieUpdateRequest{}).
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
		WithJSON(&dto.MovieUpdateRequest{Title: null.StringFrom("t")}).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			validation.Errors{
				"title": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.Film.Title.MinLength,
						"max": config.Config.Validation.Film.Title.MaxLength,
					},
				),
			}.Error(),
		))

	// movie not found
	e.Request(method, path).
		WithPath("id", 9999999).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(&dto.MovieUpdateRequest{}).
		Expect().
		Status(http.StatusNotFound).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			http.StatusText(http.StatusNotFound),
		))

	updates := []struct {
		name string
		req  *dto.MovieUpdateRequest
	}{
		{
			name: "u1",
			req:  &dto.MovieUpdateRequest{},
		},
		{
			name: "u2",
			req: &dto.MovieUpdateRequest{
				Title: null.StringFrom("updated_title"),
			},
		},
		{
			name: "u3",
			req: &dto.MovieUpdateRequest{
				Descriptions: null.StringFrom("updated_description"),
			},
		},
		{
			name: "u4",
			req: &dto.MovieUpdateRequest{
				DateReleased: null.TimeFrom(testutils.Date(2000, 1, 1)),
			},
		},
		{
			name: "u5",
			req: &dto.MovieUpdateRequest{
				Duration: null.IntFrom(10 * 60),
			},
		},
		{
			name: "u6",
			req: &dto.MovieUpdateRequest{
				Title:        null.StringFrom("updated_title"),
				Descriptions: null.StringFrom("updated_description"),
				DateReleased: null.TimeFrom(testutils.Date(2000, 1, 1)),
				Duration:     null.IntFrom(10 * 60),
			},
		},
	}

	for i, u := range updates {
		u := u
		i := i
		t.Run(u.name, func(t *testing.T) {
			require := prequire.New(t)

			// insert movie
			movieID, err := appInstance.MovieCreate(
				ctx,
				defaults.user.id,
				&dto.MovieCreateRequest{
					Title:        "movie" + strconv.Itoa(i),
					DateReleased: testutils.Date(1900, 3, 14),
				},
			)
			require.NoError(err)

			gotMovieBeforeUpdate, err := appInstance.MovieGet(ctx, movieID)
			require.NoError(err)

			updateTime := time.Now()

			// update
			e.Request(method, path).
				WithPath("id", movieID).
				WithHeader(echo.HeaderAuthorization, defaults.user.auth).
				WithJSON(u.req).
				Expect().
				Status(http.StatusOK).
				NoContent()

			// check updated fields
			gotMovieAfterUpdate, err := appInstance.MovieGet(ctx, movieID)
			require.NoError(err)

			require.GreaterOrEqual(
				gotMovieAfterUpdate.ContributedAt,
				updateTime,
			)

			updatedMovie := &models.Film{}
			updatedMovie.ID = movieID
			if u.req.Title.Valid {
				updatedMovie.Title = u.req.Title.String
			} else {
				updatedMovie.Title = gotMovieBeforeUpdate.Title
			}
			if u.req.Descriptions.Valid {
				updatedMovie.Descriptions = u.req.Descriptions
			} else {
				updatedMovie.Descriptions = gotMovieBeforeUpdate.Descriptions
			}
			if u.req.DateReleased.Valid {
				updatedMovie.DateReleased = u.req.DateReleased.Time
			} else {
				updatedMovie.DateReleased = gotMovieBeforeUpdate.DateReleased
			}
			if u.req.Duration.Valid {
				updatedMovie.Duration = u.req.Duration
			} else {
				updatedMovie.Duration = gotMovieBeforeUpdate.Duration
			}
			updatedMovie.Invalidation = null.String{}
			updatedMovie.ContributedBy = defaults.user.id
			updatedMovie.ContributedAt = gotMovieAfterUpdate.ContributedAt

			testutils.SetTimeLocation(
				&updatedMovie.DateReleased,
				gotMovieAfterUpdate.DateReleased.Location(),
			)

			require.Equal(updatedMovie, gotMovieAfterUpdate)
		})
	}
}

func TestHandleMovieInvalidate(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	server, appInstance, defaults, teardown := setup(OptEnableDefaultUser)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/authorized/movie/{id}/invalidate"
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

	// movie not found
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

	// invalidate movie
	movieCreateReq := &dto.MovieCreateRequest{
		Title:        "movie",
		DateReleased: testutils.Date(1900, 3, 14),
	}
	movieID, err := appInstance.MovieCreate(
		ctx,
		defaults.user.id,
		movieCreateReq,
	)
	require.NoError(err)

	e.Request(method, path).
		WithPath("id", movieID).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(invalidationRequest).
		Expect().
		Status(http.StatusOK).
		NoContent()

	// check movie invalidated
	gotInvalidatedMovie, err := appInstance.MovieGet(ctx, movieID)
	require.NoError(err)

	testutils.SetTimeLocation(
		&movieCreateReq.DateReleased,
		gotInvalidatedMovie.DateReleased.Location(),
	)

	require.Equal(
		&models.Film{
			ID:            movieID,
			Title:         movieCreateReq.Title,
			Descriptions:  movieCreateReq.Descriptions,
			DateReleased:  movieCreateReq.DateReleased,
			Duration:      movieCreateReq.Duration,
			SeriesID:      null.Int{},
			SeasonNumber:  null.Int{},
			EpisodeNumber: null.Int{},
			Invalidation:  null.StringFrom(invalidationRequest.Invalidation),
			ContributedBy: defaults.user.id,
			ContributedAt: gotInvalidatedMovie.ContributedAt,
		},
		gotInvalidatedMovie,
	)
}

func TestHandleMovieAuditsGetAll(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	server, appInstance, defaults, teardown := setup(OptEnableDefaultUser)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/authorized/movie/{id}/audits"
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

	// movie not found
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

	// add movie
	movieCreateReq := &dto.MovieCreateRequest{
		Title:        "movie",
		DateReleased: testutils.Date(2000, 1, 1),
	}
	movieID, err := appInstance.MovieCreate(
		ctx,
		defaults.user.id,
		movieCreateReq,
	)
	require.NoError(err)

	gotMovie, err := appInstance.MovieGet(ctx, movieID)
	require.NoError(err)

	require.GreaterOrEqual(gotMovie.ContributedAt, createTime)

	expMovieUpdateAudit := &models.FilmsAudit{
		ID:            movieID,
		Title:         movieCreateReq.Title,
		Descriptions:  movieCreateReq.Descriptions,
		DateReleased:  movieCreateReq.DateReleased,
		Duration:      movieCreateReq.Duration,
		SeriesID:      null.Int{},
		SeasonNumber:  null.Int{},
		EpisodeNumber: null.Int{},
		Invalidation:  null.String{},
		ContributedBy: defaults.user.id,
		ContributedAt: gotMovie.ContributedAt,
	}

	// no audits
	e.Request(method, path).
		WithPath("id", movieID).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Equal(response.Paginated(
			config.Config.Validation.Pagination.Page.MinValue,
			config.Config.Validation.Pagination.PageSize.DefaultValue,
			([]*models.FilmsAudit)(nil),
			0,
		))

	updateTime := time.Now()

	// update the movie
	movieUpdateReq := &dto.MovieUpdateRequest{
		Title:        null.StringFrom("updated title"),
		Descriptions: null.StringFrom("updated descriptions"),
		DateReleased: null.TimeFrom(
			testutils.Date(2005, 11, 14),
		),
		Duration: null.IntFrom(10 * 60),
	}
	err = appInstance.MovieUpdate(
		ctx,
		movieID,
		defaults.user.id,
		movieUpdateReq,
	)
	require.NoError(err)

	gotMovie, err = appInstance.MovieGet(ctx, movieID)
	require.NoError(err)

	require.GreaterOrEqual(gotMovie.ContributedAt, updateTime)

	expMovieInvalidationAudit := &models.FilmsAudit{
		ID:            movieID,
		Title:         movieUpdateReq.Title.String,
		Descriptions:  movieUpdateReq.Descriptions,
		DateReleased:  movieUpdateReq.DateReleased.Time,
		Duration:      movieUpdateReq.Duration,
		SeriesID:      null.Int{},
		SeasonNumber:  null.Int{},
		EpisodeNumber: null.Int{},
		Invalidation:  null.String{},
		ContributedBy: defaults.user.id,
		ContributedAt: gotMovie.ContributedAt,
	}

	// get update audits
	e.Request(method, path).
		WithPath("id", movieID).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Equal(response.Paginated(
			config.Config.Validation.Pagination.Page.MinValue,
			config.Config.Validation.Pagination.PageSize.DefaultValue,
			[]*models.FilmsAudit{expMovieUpdateAudit},
			1,
		))

	// invalidate the movie
	err = appInstance.MovieInvalidate(
		ctx,
		movieID,
		defaults.user.id,
		&dto.InvalidationRequest{Invalidation: "invalidation"},
	)
	require.NoError(err)

	// get invalidation audits
	e.Request(method, path).
		WithPath("id", movieID).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Equal(response.Paginated(
			config.Config.Validation.Pagination.Page.MinValue,
			config.Config.Validation.Pagination.PageSize.DefaultValue,
			[]*models.FilmsAudit{
				expMovieInvalidationAudit,
				expMovieUpdateAudit,
			},
			2,
		))
}

func TestHandleMoviesSearch(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	server, appInstance, defaults, teardown := setup(OptEnableDefaultUser)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/authorized/movie/search"
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

	// no movies
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
			[]*models.Film{},
			0,
		))

	// add movies

	queryMovies := []*dto.MovieCreateRequest{
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
	// interleave queryMovies
	movieCreateReqs := make([]*dto.MovieCreateRequest, len(queryMovies)*2)
	for i := range movieCreateReqs {
		var movie *dto.MovieCreateRequest
		if i%2 == 0 {
			qi := 0
			if i != 0 {
				qi = i / 2
			}
			movie = queryMovies[qi]
		} else {
			movie = &dto.MovieCreateRequest{
				Title:        "title_ignore_me",
				Descriptions: null.StringFrom("descriptions"),
			}
		}
		movieCreateReqs[i] = movie
	}

	movieIDs := make([]int, len(movieCreateReqs))

	createTime := time.Now()

	var thePosterUri string

	for i, req := range movieCreateReqs {
		// set DateReleased field to skip request validation error
		req.DateReleased = testutils.Date(2000, 1, 1)

		var err error
		movieIDs[i], err = appInstance.MovieCreate(
			ctx,
			defaults.user.id,
			req,
		)
		require.NoError(err)

		// set poster just for the first one
		if i == 0 {
			thePosterUri = e.Request(http.MethodPut, "/v1/authorized/movie/{id}/poster").
				WithPath("id", movieIDs[i]).
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

	// wait until movie got index
	err := testutils.WaitUntil(
		func() (bool, error) {
			c, err := searchtestutils.CountIndex(
				esClient,
				config.Config.Elasticsearch.Index.Movies,
			)
			if err != nil {
				return false, err
			}
			return c == len(movieCreateReqs), nil
		},
		10*time.Second,
		time.Second,
	)
	require.NoError(err)

	gotMovies, _, err := appInstance.MoviesGetAll(ctx, query.Options{
		Offset:    0,
		Limit:     config.Config.Validation.Pagination.PageSize.MaxValue,
		SortField: models.FilmColumns.ID,
		SortOrder: request.SortOrderAsc,
	})
	require.NoError(err)

	items := make([]*models.Film, len(queryMovies))

	itemsIndex := 0
	for i := range gotMovies {
		if gotMovies[i].Title != "title_ignore_me" {

			var poster null.String
			if i == 0 {
				poster = null.StringFrom(thePosterUri)
			}

			require.GreaterOrEqual(gotMovies[i].ContributedAt, createTime)

			items[itemsIndex] = &models.Film{
				ID:            movieIDs[i],
				Title:         movieCreateReqs[i].Title,
				Descriptions:  movieCreateReqs[i].Descriptions,
				DateReleased:  movieCreateReqs[i].DateReleased,
				Duration:      movieCreateReqs[i].Duration,
				Poster:        poster,
				Invalidation:  null.String{},
				ContributedBy: defaults.user.id,
				ContributedAt: gotMovies[i].ContributedAt,
			}

			itemsIndex++

		}
	}

	// search movie
	responseObject := e.Request(method, path).
		WithQuery("query", "query").
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object()

	reponseItems := responseObject.
		ValueEqual("page", config.Config.Validation.Pagination.Page.MinValue).
		ValueEqual("page_size", config.Config.Validation.Pagination.PageSize.DefaultValue).
		ValueEqual("total_pages", (len(items)+config.Config.Validation.Pagination.PageSize.DefaultValue-1)/config.Config.Validation.Pagination.PageSize.DefaultValue).
		ValueEqual("total_items", len(items)).
		Value("items").Array()

	toAnySlice := func(items []*models.Film) []any {
		anyItems := make([]any, len(items))
		for i, it := range items {
			anyItems[i] = it
		}
		return anyItems
	}

	reponseItems.Length().Equal(len(items))
	reponseItems.Contains(toAnySlice(items)...)
}

func TestHandleMoviePutPoster(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	server, appInstance, defaults, teardown := setup(OptEnableDefaultUser)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/authorized/movie/{id}/poster"
	method := http.MethodPut

	filename := config.Config.MinIO.Filename.Movie

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

	// movie not found
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

	movieCreateReq := &dto.MovieCreateRequest{
		Title:        "movie",
		DateReleased: testutils.Date(1900, 3, 14),
	}
	movieID, err := appInstance.MovieCreate(
		ctx,
		defaults.user.id,
		movieCreateReq,
	)
	require.NoError(err)

	// movie have no poster

	movieGetReq := struct {
		method string
		path   string
	}{
		method: http.MethodGet,
		path:   "/v1/authorized/movie/{id}",
	}

	e.Request(movieGetReq.method, movieGetReq.path).
		WithPath("id", movieID).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("poster").Null()

	// missing file
	e.Request(method, path).
		WithPath("id", movieID).
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
		WithPath("id", movieID).
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
		WithPath("id", movieID).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithMultipart().WithFile(filename, file.path).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("uri").String().NotEmpty().Raw()

	// now movie have poster
	e.Request(movieGetReq.method, movieGetReq.path).
		WithPath("id", movieID).
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
		WithPath("id", movieID).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithMultipart().WithFile(filename, overwrittenFile.path).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("uri").String().NotEmpty().Raw()

	// check overwritten movie poster
	e.Request(movieGetReq.method, movieGetReq.path).
		WithPath("id", movieID).
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
