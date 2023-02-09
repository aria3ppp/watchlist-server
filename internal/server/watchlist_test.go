package server_test

import (
	"context"
	"math"
	"net/http"
	"testing"
	"time"

	"github.com/aria3ppp/watchlist-server/internal/config"
	"github.com/aria3ppp/watchlist-server/internal/dto"
	"github.com/aria3ppp/watchlist-server/internal/models"
	"github.com/aria3ppp/watchlist-server/internal/query"
	"github.com/aria3ppp/watchlist-server/internal/repo"
	"github.com/aria3ppp/watchlist-server/internal/server/request"
	"github.com/aria3ppp/watchlist-server/internal/server/response"
	"github.com/aria3ppp/watchlist-server/internal/testutils"
	"github.com/aria3ppp/watchlist-server/internal/watchlist"
	"github.com/gavv/httpexpect/v2"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/null/v8"
)

func TestHandleWatchlistGet(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	server, appInstance, defaults, teardown := setup(OptEnableDefaultSeries)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/authorized/watchlist"
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

	// no watchlist
	e.Request(method, path).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Equal(response.Paginated(
			config.Config.Validation.Pagination.Page.MinValue,
			config.Config.Validation.Pagination.PageSize.DefaultValue,
			([]*watchlist.Item)(nil),
			0,
		))

	// add films

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
	episodePutReqs := []*dto.EpisodePutRequest{
		{
			Title:        "e1",
			DateReleased: testutils.Date(2005, 1, 1),
		},
		{
			Title:        "e2",
			DateReleased: testutils.Date(2006, 1, 1),
		},
		{
			Title:        "e3",
			DateReleased: testutils.Date(2007, 1, 1),
		},
		{
			Title:        "e4",
			DateReleased: testutils.Date(2008, 1, 1),
		},
		{
			Title:        "e5",
			DateReleased: testutils.Date(2009, 1, 1),
		},
	}

	watchlistItems := make(
		[]*watchlist.Item,
		len(movieCreateReqs)+len(episodePutReqs),
	)

	createTime := time.Now()

	for i, req := range movieCreateReqs {
		filmID, err := appInstance.MovieCreate(ctx, defaults.user.id, req)
		require.NoError(err)
		gotMovie, err := appInstance.MovieGet(ctx, filmID)
		require.NoError(err)
		require.LessOrEqual(createTime, gotMovie.ContributedAt)
		watchID, err := appInstance.WatchlistAdd(ctx, defaults.user.id, filmID)
		require.NoError(err)
		gotWatchlist, total, err := appInstance.WatchlistGet(
			ctx,
			defaults.user.id,
			query.WatchlistOptions{
				Offset: i, Limit: 1, SortOrder: request.SortOrderAsc, WhereTimeWatched: repo.RawSqlWhereTimeWatchedEmptyClause,
			},
		)
		require.NoError(err)
		require.Equal(i+1, total)
		require.Equal(1, len(gotWatchlist))
		require.LessOrEqual(createTime, gotWatchlist[0].TimeAdded)
		watchlistItems[i] = &watchlist.Item{
			Watchfilm: models.Watchfilm{
				ID:          watchID,
				UserID:      defaults.user.id,
				FilmID:      filmID,
				TimeAdded:   gotWatchlist[0].TimeAdded,
				TimeWatched: null.Time{},
			},
			Film: models.Film{
				ID:            filmID,
				Title:         req.Title,
				Descriptions:  req.Descriptions,
				DateReleased:  req.DateReleased,
				Duration:      req.Duration,
				SeriesID:      null.Int{},
				SeasonNumber:  null.Int{},
				EpisodeNumber: null.Int{},
				ContributedBy: defaults.user.id,
				ContributedAt: gotMovie.ContributedAt,
				Invalidation:  null.String{},
			},
		}
	}
	seasonNumber := 1
	for _i, req := range episodePutReqs {
		episodeNumber := _i + 1
		i := _i + len(movieCreateReqs)

		err := appInstance.EpisodePut(
			ctx,
			defaults.series.id,
			seasonNumber,
			episodeNumber,
			defaults.user.id,
			req,
		)
		require.NoError(err)
		gotEpisode, err := appInstance.EpisodeGet(
			ctx,
			defaults.series.id,
			seasonNumber,
			episodeNumber,
		)
		require.NoError(err)
		require.LessOrEqual(createTime, gotEpisode.ContributedAt)
		watchID, err := appInstance.WatchlistAdd(
			ctx,
			defaults.user.id,
			gotEpisode.ID,
		)
		require.NoError(err)
		gotWatchlist, total, err := appInstance.WatchlistGet(
			ctx,
			defaults.user.id,
			query.WatchlistOptions{
				Offset: i, Limit: 1, SortOrder: request.SortOrderAsc, WhereTimeWatched: repo.RawSqlWhereTimeWatchedEmptyClause,
			},
		)
		require.NoError(err)
		require.Equal(i+1, total)
		require.Equal(1, len(gotWatchlist))
		require.LessOrEqual(createTime, gotWatchlist[0].TimeAdded)
		watchlistItems[i] = &watchlist.Item{
			Watchfilm: models.Watchfilm{
				ID:          watchID,
				UserID:      defaults.user.id,
				FilmID:      gotEpisode.ID,
				TimeAdded:   gotWatchlist[0].TimeAdded,
				TimeWatched: null.Time{},
			},
			Film: models.Film{
				ID:            gotEpisode.ID,
				Title:         req.Title,
				Descriptions:  req.Descriptions,
				DateReleased:  req.DateReleased,
				Duration:      req.Duration,
				SeriesID:      null.IntFrom(defaults.series.id),
				SeasonNumber:  null.IntFrom(seasonNumber),
				EpisodeNumber: null.IntFrom(episodeNumber),
				ContributedBy: defaults.user.id,
				ContributedAt: gotEpisode.ContributedAt,
				Invalidation:  null.String{},
			},
		}
	}

	// get watchlist

	e.Request(method, path).
		WithQuery("sort_order", "asc").
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Equal(response.Paginated(
			config.Config.Validation.Pagination.Page.MinValue,
			config.Config.Validation.Pagination.PageSize.DefaultValue,
			watchlistItems,
			len(watchlistItems),
		))
}

func TestHandleWatchlistAdd(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	server, appInstance, defaults, teardown := setup(OptEnableDefaultUser)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/authorized/watchlist/add"
	method := http.MethodPost

	// invalid query
	e.Request(method, path).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			validation.Errors{"film_id": validation.ErrRequired}.Error(),
		))

	// not found
	e.Request(method, path).
		WithQueryObject(request.WatchlistAddQuery{FilmID: 999}).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusNotFound).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			http.StatusText(http.StatusNotFound),
		))

	// add film

	movieCreateReq := &dto.MovieCreateRequest{
		Title:        "film",
		DateReleased: testutils.Date(2000, 1, 2),
	}

	createTime := time.Now()

	filmID, err := appInstance.MovieCreate(
		ctx,
		defaults.user.id,
		movieCreateReq,
	)
	require.NoError(err)

	expWatchID := 1
	e.Request(method, path).
		WithQueryObject(request.WatchlistAddQuery{FilmID: filmID}).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Equal(response.ID(expWatchID))

	// check film added

	gotWatchlist, total, err := appInstance.WatchlistGet(
		ctx,
		defaults.user.id,
		query.WatchlistOptions{
			Offset: 0, Limit: 1, SortOrder: request.SortOrderAsc, WhereTimeWatched: repo.RawSqlWhereTimeWatchedEmptyClause,
		},
	)
	require.NoError(err)
	require.Equal(1, total)
	require.Equal(1, len(gotWatchlist))
	require.LessOrEqual(createTime, gotWatchlist[0].Watchfilm.TimeAdded)
	require.LessOrEqual(createTime, gotWatchlist[0].Film.ContributedAt)
	testutils.SetTimeLocation(
		&gotWatchlist[0].Film.DateReleased,
		movieCreateReq.DateReleased.Location(),
	)
	require.Equal(
		&watchlist.Item{
			Watchfilm: models.Watchfilm{
				ID:          expWatchID,
				UserID:      defaults.user.id,
				FilmID:      filmID,
				TimeAdded:   gotWatchlist[0].TimeAdded,
				TimeWatched: null.Time{},
			},
			Film: models.Film{
				ID:            filmID,
				Title:         movieCreateReq.Title,
				Descriptions:  movieCreateReq.Descriptions,
				DateReleased:  movieCreateReq.DateReleased,
				Duration:      movieCreateReq.Duration,
				ContributedBy: defaults.user.id,
				ContributedAt: gotWatchlist[0].Film.ContributedAt,
				Invalidation:  null.String{},
			},
		},
		gotWatchlist[0],
	)
}

func TestHandleWatchlistDelete(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	server, appInstance, defaults, teardown := setup(OptEnableDefaultUser)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/authorized/watchlist/{id}"
	method := http.MethodDelete

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

	// not found
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

	// delete

	movieCreateReq := &dto.MovieCreateRequest{
		Title:        "film",
		DateReleased: testutils.Date(2000, 1, 2),
	}
	filmID, err := appInstance.MovieCreate(
		ctx,
		defaults.user.id,
		movieCreateReq,
	)
	require.NoError(err)
	watchID, err := appInstance.WatchlistAdd(ctx, defaults.user.id, filmID)
	require.NoError(err)

	e.Request(method, path).
		WithPath("id", watchID).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusOK).
		NoContent()

	// check deleted

	watchlist, total, err := appInstance.WatchlistGet(
		ctx,
		defaults.user.id,
		query.WatchlistOptions{
			Offset: 0, Limit: math.MaxInt, SortOrder: request.SortOrderAsc, WhereTimeWatched: repo.RawSqlWhereTimeWatchedEmptyClause,
		},
	)
	require.NoError(err)
	require.Equal(0, total)
	require.Equal(0, len(watchlist))
}

func TestHandleWatchlistSetWatched(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	server, appInstance, defaults, teardown := setup(OptEnableDefaultUser)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/authorized/watchlist/{id}"
	method := http.MethodPatch

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

	// not found
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

	// set watched

	movieCreateReq := &dto.MovieCreateRequest{
		Title:        "film",
		DateReleased: testutils.Date(2000, 1, 2),
	}
	filmID, err := appInstance.MovieCreate(
		ctx,
		defaults.user.id,
		movieCreateReq,
	)
	require.NoError(err)
	watchID, err := appInstance.WatchlistAdd(ctx, defaults.user.id, filmID)
	require.NoError(err)

	watchTime := time.Now()

	e.Request(method, path).
		WithPath("id", watchID).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusOK).
		NoContent()

	// check watched

	watchlist, total, err := appInstance.WatchlistGet(
		ctx,
		defaults.user.id,
		query.WatchlistOptions{
			Offset: 0, Limit: math.MaxInt, SortOrder: request.SortOrderAsc, WhereTimeWatched: repo.RawSqlWhereTimeWatchedIsNotNull,
		},
	)
	require.NoError(err)
	require.Equal(1, total)
	require.Equal(1, len(watchlist))
	require.True(watchlist[0].TimeWatched.Valid)
	require.LessOrEqual(watchTime, watchlist[0].TimeWatched.Time)
}
