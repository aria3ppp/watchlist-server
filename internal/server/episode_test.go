package server_test

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/aria3ppp/watchlist-server/internal/config"
	"github.com/aria3ppp/watchlist-server/internal/dto"
	"github.com/aria3ppp/watchlist-server/internal/models"
	"github.com/aria3ppp/watchlist-server/internal/query"
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

func TestHandleEpisodeGet(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	server, appInstance, defaults, teardown := setup(OptEnableDefaultSeries)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/authorized/series/{id}/season/{se}/episode/{ep}"
	method := http.MethodGet

	var (
		seasonNumber  = 1
		episodeNumber = 1
	)

	// invalid id/numbers
	e.Request(method, path).
		WithPath("id", -1).
		WithPath("se", -1).
		WithPath("ep", -1).
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
				"season_number": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{"threshold": 1},
				),
				"episode_number": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{"threshold": 1},
				),
			}.Error(),
		))

	// episode not found
	e.Request(method, path).
		WithPath("id", defaults.series.id).
		WithPath("se", seasonNumber).
		WithPath("ep", episodeNumber).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusNotFound).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			http.StatusText(http.StatusNotFound),
		))

	putTime := time.Now()

	// add a new episode
	episodePutReq := &dto.EpisodePutRequest{
		Title:        "episode",
		DateReleased: testutils.Date(2000, 1, 1),
	}
	err := appInstance.EpisodePut(
		ctx,
		defaults.series.id, seasonNumber, episodeNumber,
		defaults.user.id,
		episodePutReq,
	)
	require.NoError(err)

	gotEpisode, err := appInstance.EpisodeGet(
		ctx,
		defaults.series.id,
		seasonNumber,
		episodeNumber,
	)
	require.NoError(err)

	require.GreaterOrEqual(gotEpisode.ContributedAt, putTime)

	// get episode
	e.Request(method, path).
		WithPath("id", defaults.series.id).
		WithPath("se", seasonNumber).
		WithPath("ep", episodeNumber).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Equal(&models.Film{
			ID:            gotEpisode.ID,
			Title:         episodePutReq.Title,
			Descriptions:  episodePutReq.Descriptions,
			DateReleased:  episodePutReq.DateReleased,
			Duration:      episodePutReq.Duration,
			SeriesID:      null.IntFrom(defaults.series.id),
			SeasonNumber:  null.IntFrom(seasonNumber),
			EpisodeNumber: null.IntFrom(episodeNumber),
			Invalidation:  null.String{},
			ContributedBy: defaults.user.id,
			ContributedAt: gotEpisode.ContributedAt,
		})
}

func TestHandleEpisodesGetAllBySeries(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	server, appInstance, defaults, teardown := setup(OptEnableDefaultSeries)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/authorized/series/{id}/episode"
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

	// no episodes
	e.Request(method, path).
		WithPath("id", defaults.series.id).
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

	// add episodes
	episodePutReqs := [][]*dto.EpisodePutRequest{
		{
			{
				Title:        "s1e1",
				DateReleased: testutils.Date(2000, 1, 1),
			},
			{
				Title:        "s1e2",
				DateReleased: testutils.Date(2001, 1, 1),
			},
			{
				Title:        "s1e3",
				DateReleased: testutils.Date(2002, 1, 1),
			},
		},
		{
			{
				Title:        "s2e1",
				DateReleased: testutils.Date(2005, 1, 1),
			},
			{
				Title:        "s2e2",
				DateReleased: testutils.Date(2006, 1, 1),
			},
			{
				Title:        "s2e3",
				DateReleased: testutils.Date(2007, 1, 1),
			},
			{
				Title:        "s2e4",
				DateReleased: testutils.Date(2008, 1, 1),
			},
			{
				Title:        "s2e5",
				DateReleased: testutils.Date(2009, 1, 1),
			},
		},
		{
			{
				Title:        "s3e1",
				DateReleased: testutils.Date(2010, 1, 1),
			},
			{
				Title:        "s3e2",
				DateReleased: testutils.Date(2011, 1, 1),
			},
			{
				Title:        "s3e3",
				DateReleased: testutils.Date(2012, 1, 1),
			},
			{
				Title:        "s3e4",
				DateReleased: testutils.Date(2013, 1, 1),
			},
		},
	}
	seasonEpisodeNumbers := make([][]struct{ se, ep int }, len(episodePutReqs))

	putTime := time.Now()

	for s, sreq := range episodePutReqs {
		seasonEpisodeNumbers[s] = make([]struct{ se, ep int }, len(sreq))
		for e, ereq := range sreq {
			se := s + 1
			ep := e + 1
			err := appInstance.EpisodePut(
				ctx,
				defaults.series.id,
				se,
				ep,
				defaults.user.id,
				ereq,
			)
			require.NoError(err)
			seasonEpisodeNumbers[s][e].se = se
			seasonEpisodeNumbers[s][e].ep = ep
		}
	}

	gotEpisodes, total, err := appInstance.EpisodesGetAllBySeries(
		ctx,
		defaults.series.id,
		query.SortOrderOptions{
			SortOrder: "asc",
			Offset:    0,
			Limit:     config.Config.Validation.Pagination.PageSize.MaxValue,
		},
	)
	require.NoError(err)

	items := make([]*models.Film, len(gotEpisodes))

	i := 0
	for s, sreq := range seasonEpisodeNumbers {
		for e := range sreq {
			require.GreaterOrEqual(gotEpisodes[i].ContributedAt, putTime)

			items[i] = &models.Film{
				ID:            gotEpisodes[i].ID,
				Title:         episodePutReqs[s][e].Title,
				Descriptions:  episodePutReqs[s][e].Descriptions,
				DateReleased:  episodePutReqs[s][e].DateReleased,
				Duration:      episodePutReqs[s][e].Duration,
				SeriesID:      null.IntFrom(defaults.series.id),
				SeasonNumber:  null.IntFrom(seasonEpisodeNumbers[s][e].se),
				EpisodeNumber: null.IntFrom(seasonEpisodeNumbers[s][e].ep),
				Invalidation:  null.String{},
				ContributedBy: defaults.user.id,
				ContributedAt: gotEpisodes[i].ContributedAt,
			}

			i++
		}
	}

	// get all episodes
	rawResp := e.Request(method, path).
		WithPath("id", defaults.series.id).
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
		)).
		Raw()

	t.Log("[DEBUG] all episodes from all seasons:")
	json, err := json.MarshalIndent(rawResp, "", "\t")
	require.NoError(err)
	t.Log(string(json))
}

func TestHandleEpisodesGetAllBySeason(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	server, appInstance, defaults, teardown := setup(OptEnableDefaultSeries)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/authorized/series/{id}/season/{se}/episode"
	method := http.MethodGet

	seasonNumber := 1

	// invalid id/number
	e.Request(method, path).
		WithPath("id", -1).
		WithPath("se", -1).
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
				"season_number": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{"threshold": 1},
				),
			}.Error(),
		))

	// invalid query
	e.Request(method, path).
		WithPath("id", 1).
		WithPath("se", 1).
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
		WithPath("se", seasonNumber).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusNotFound).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			http.StatusText(http.StatusNotFound),
		))

	// no episodes
	e.Request(method, path).
		WithPath("id", defaults.series.id).
		WithPath("se", seasonNumber).
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

	// add episodes
	episodePutReqs := []*dto.EpisodePutRequest{
		{
			Title:        "e1",
			DateReleased: testutils.Date(2000, 1, 1),
		},
		{
			Title:        "e2",
			DateReleased: testutils.Date(2001, 1, 1),
		},
		{
			Title:        "e3",
			DateReleased: testutils.Date(2002, 1, 1),
		},
		{
			Title:        "e4",
			DateReleased: testutils.Date(2003, 1, 1),
		},
		{
			Title:        "e5",
			DateReleased: testutils.Date(2004, 1, 1),
		},
	}
	episodeNumbers := make([]int, len(episodePutReqs))

	putTime := time.Now()

	for i, req := range episodePutReqs {
		ep := i + 1
		err := appInstance.EpisodePut(
			ctx,
			defaults.series.id,
			seasonNumber,
			ep,
			defaults.user.id,
			req,
		)
		require.NoError(err)
		episodeNumbers[i] = ep
	}

	gotEpisodes, total, err := appInstance.EpisodesGetAllBySeason(
		ctx,
		defaults.series.id,
		seasonNumber,
		query.SortOrderOptions{
			SortOrder: "asc",
			Offset:    0,
			Limit:     config.Config.Validation.Pagination.PageSize.MaxValue,
		},
	)
	require.NoError(err)

	items := make([]*models.Film, len(gotEpisodes))

	for i := range gotEpisodes {
		require.GreaterOrEqual(gotEpisodes[i].ContributedAt, putTime)

		items[i] = &models.Film{
			ID:            gotEpisodes[i].ID,
			Title:         episodePutReqs[i].Title,
			Descriptions:  episodePutReqs[i].Descriptions,
			DateReleased:  episodePutReqs[i].DateReleased,
			Duration:      episodePutReqs[i].Duration,
			SeriesID:      null.IntFrom(defaults.series.id),
			SeasonNumber:  null.IntFrom(seasonNumber),
			EpisodeNumber: null.IntFrom(episodeNumbers[i]),
			Invalidation:  null.String{},
			ContributedBy: defaults.user.id,
			ContributedAt: gotEpisodes[i].ContributedAt,
		}
	}

	// get all episodes
	e.Request(method, path).
		WithPath("id", defaults.series.id).
		WithPath("se", seasonNumber).
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

func TestHandleEpisodePut(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	server, appInstance, defaults, teardown := setup(OptEnableDefaultSeries)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/authorized/series/{id}/season/{se}/episode/{ep}"
	method := http.MethodPut

	var (
		seasonNumber  = 1
		episodeNumber = 1
	)

	episodePutReq := &dto.EpisodePutRequest{
		Title:        "episode",
		DateReleased: testutils.Date(2000, 1, 1),
	}

	// invalid id/number
	e.Request(method, path).
		WithPath("id", -1).
		WithPath("se", -1).
		WithPath("ep", -1).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(episodePutReq).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			validation.Errors{
				"id": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{"threshold": 1},
				),
				"season_number": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{"threshold": 1},
				),
				"episode_number": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{"threshold": 1},
				),
			}.Error(),
		))

	// invalid request
	e.Request(method, path).
		WithPath("id", 1).
		WithPath("se", 1).
		WithPath("ep", 1).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(dto.EpisodePutRequest{}).
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

	// series not found
	e.Request(method, path).
		WithPath("id", 999).
		WithPath("se", 999).
		WithPath("ep", 999).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(episodePutReq).
		Expect().
		Status(http.StatusNotFound).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			http.StatusText(http.StatusNotFound),
		))

	// put episode
	putTime := time.Now()

	e.Request(method, path).
		WithPath("id", defaults.series.id).
		WithPath("se", seasonNumber).
		WithPath("ep", episodeNumber).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(episodePutReq).
		Expect().
		Status(http.StatusOK).
		NoContent()

	// check episode put in place
	gotEpisode, err := appInstance.EpisodeGet(
		ctx,
		defaults.series.id,
		seasonNumber,
		episodeNumber,
	)
	require.NoError(err)

	testutils.SetTimeLocation(
		&episodePutReq.DateReleased,
		gotEpisode.DateReleased.Location(),
	)

	require.GreaterOrEqual(gotEpisode.ContributedAt, putTime)

	require.Equal(
		&models.Film{
			ID:            gotEpisode.ID,
			Title:         episodePutReq.Title,
			Descriptions:  episodePutReq.Descriptions,
			DateReleased:  episodePutReq.DateReleased,
			Duration:      episodePutReq.Duration,
			SeriesID:      null.IntFrom(defaults.series.id),
			SeasonNumber:  null.IntFrom(seasonNumber),
			EpisodeNumber: null.IntFrom(episodeNumber),
			Invalidation:  null.String{},
			ContributedBy: defaults.user.id,
			ContributedAt: gotEpisode.ContributedAt,
		},
		gotEpisode,
	)
}

func TestHandleEpisodesPutAllBySeason(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	server, appInstance, defaults, teardown := setup(OptEnableDefaultSeries)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/authorized/series/{id}/season/{se}/episode"
	method := http.MethodPut

	seasonNumber := 1

	episodePutAllReq := &dto.EpisodesPutAllBySeasonRequest{
		Episodes: []*dto.EpisodePutRequest{
			{
				Title:        "episode1",
				DateReleased: testutils.Date(2000, 1, 1),
			},
			{
				Title:        "episode2",
				DateReleased: testutils.Date(2001, 1, 1),
			},
			{
				Title:        "episode3",
				DateReleased: testutils.Date(2002, 1, 1),
			},
			{
				Title:        "episode4",
				DateReleased: testutils.Date(2003, 1, 1),
			},
			{
				Title:        "episode5",
				DateReleased: testutils.Date(2004, 1, 1),
			},
		},
	}

	// invalid id/number
	e.Request(method, path).
		WithPath("id", -1).
		WithPath("se", -1).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(episodePutAllReq).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			validation.Errors{
				"id": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{"threshold": 1},
				),
				"season_number": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{"threshold": 1},
				),
			}.Error(),
		))

	// invalid request
	e.Request(method, path).
		WithPath("id", 1).
		WithPath("se", 1).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(dto.EpisodesPutAllBySeasonRequest{}).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			validation.Errors{
				"episodes": validation.ErrRequired,
			}.Error(),
		))

	// series not found
	e.Request(method, path).
		WithPath("id", 999).
		WithPath("se", 1).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(episodePutAllReq).
		Expect().
		Status(http.StatusNotFound).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			http.StatusText(http.StatusNotFound),
		))

	// put episodes
	putTime := time.Now()

	e.Request(method, path).
		WithPath("id", defaults.series.id).
		WithPath("se", seasonNumber).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(episodePutAllReq).
		Expect().
		Status(http.StatusOK).
		NoContent()

	// check episode put in place
	gotEpisodes, total, err := appInstance.EpisodesGetAllBySeason(
		ctx,
		defaults.series.id,
		seasonNumber,
		query.SortOrderOptions{
			SortOrder: "asc",
			Offset:    0,
			Limit:     math.MaxInt,
		},
	)
	require.NoError(err)
	require.Equal(len(episodePutAllReq.Episodes), total)

	for i, re := range episodePutAllReq.Episodes {
		testutils.SetTimeLocation(
			&re.DateReleased,
			gotEpisodes[i].DateReleased.Location(),
		)

		require.GreaterOrEqual(gotEpisodes[i].ContributedAt, putTime)

		episodeNumber := i + 1

		require.Equal(
			&models.Film{
				ID:            gotEpisodes[i].ID,
				Title:         re.Title,
				Descriptions:  re.Descriptions,
				DateReleased:  re.DateReleased,
				Duration:      re.Duration,
				SeriesID:      null.IntFrom(defaults.series.id),
				SeasonNumber:  null.IntFrom(seasonNumber),
				EpisodeNumber: null.IntFrom(episodeNumber),
				Invalidation:  null.String{},
				ContributedBy: defaults.user.id,
				ContributedAt: gotEpisodes[i].ContributedAt,
			},
			gotEpisodes[i],
		)
	}
}

func TestHandleEpisodeUpdate(t *testing.T) {
	ctx := context.Background()

	server, appInstance, defaults, teardown := setup(OptEnableDefaultSeries)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/authorized/series/{id}/season/{se}/episode/{ep}"
	method := http.MethodPatch

	seasonNumber := 1

	// invalid id
	e.Request(method, path).
		WithPath("id", -1).
		WithPath("se", -1).
		WithPath("ep", -1).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(&dto.EpisodeUpdateRequest{}).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			validation.Errors{
				"id": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{"threshold": 1},
				),
				"season_number": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{"threshold": 1},
				),
				"episode_number": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{"threshold": 1},
				),
			}.Error(),
		))

	// invalid request
	e.Request(method, path).
		WithPath("id", 1).
		WithPath("se", 1).
		WithPath("ep", 1).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(&dto.EpisodeUpdateRequest{Title: null.StringFrom("t")}).
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

	// episode not found
	e.Request(method, path).
		WithPath("id", 999).
		WithPath("se", 999).
		WithPath("ep", 999).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(&dto.EpisodeUpdateRequest{}).
		Expect().
		Status(http.StatusNotFound).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			http.StatusText(http.StatusNotFound),
		))

	updates := []struct {
		name string
		req  *dto.EpisodeUpdateRequest
	}{
		{
			name: "u1",
			req:  &dto.EpisodeUpdateRequest{},
		},
		{
			name: "u2",
			req: &dto.EpisodeUpdateRequest{
				Title: null.StringFrom("updated_title"),
			},
		},
		{
			name: "u3",
			req: &dto.EpisodeUpdateRequest{
				Descriptions: null.StringFrom("updated_description"),
			},
		},
		{
			name: "u4",
			req: &dto.EpisodeUpdateRequest{
				DateReleased: null.TimeFrom(
					testutils.Date(2000, 1, 1),
				),
			},
		},
		{
			name: "u5",
			req: &dto.EpisodeUpdateRequest{
				DateReleased: null.TimeFrom(
					testutils.Date(2008, 8, 11),
				),
			},
		},
		{
			name: "u6",
			req: &dto.EpisodeUpdateRequest{
				Title:        null.StringFrom("updated_title"),
				Descriptions: null.StringFrom("updated_description"),
				DateReleased: null.TimeFrom(
					testutils.Date(2000, 1, 1),
				),
				Duration: null.IntFrom(7 * 60),
			},
		},
	}

	for i, u := range updates {
		u := u
		i := i
		// change episode number as underlying db instance is shared between all parallel test cases
		episodeNumber := i + 1
		t.Run(u.name, func(t *testing.T) {
			require := prequire.New(t)

			// insert episode
			err := appInstance.EpisodePut(
				ctx,
				defaults.series.id, seasonNumber, episodeNumber,
				defaults.user.id,
				&dto.EpisodePutRequest{
					Title:        "episode" + strconv.Itoa(i),
					DateReleased: testutils.Date(1900, 3, 14),
				},
			)
			require.NoError(err)

			gotEpisodeBeforeUpdate, err := appInstance.EpisodeGet(
				ctx,
				defaults.series.id,
				seasonNumber,
				episodeNumber,
			)
			require.NoError(err)

			updateTime := time.Now()

			// update
			e.Request(method, path).
				WithPath("id", defaults.series.id).
				WithPath("se", seasonNumber).
				WithPath("ep", episodeNumber).
				WithHeader(echo.HeaderAuthorization, defaults.user.auth).
				WithJSON(u.req).
				Expect().
				Status(http.StatusOK).
				NoContent()

			// check updated fields
			gotEpisodeAfterUpdate, err := appInstance.EpisodeGet(
				ctx,
				defaults.series.id,
				seasonNumber,
				episodeNumber,
			)
			require.NoError(err)

			require.GreaterOrEqual(
				gotEpisodeAfterUpdate.ContributedAt,
				updateTime,
			)

			updatedEpisode := &models.Film{}
			updatedEpisode.ID = gotEpisodeAfterUpdate.ID
			updatedEpisode.SeriesID = null.IntFrom(defaults.series.id)
			updatedEpisode.SeasonNumber = null.IntFrom(seasonNumber)
			updatedEpisode.EpisodeNumber = null.IntFrom(episodeNumber)
			if u.req.Title.Valid {
				updatedEpisode.Title = u.req.Title.String
			} else {
				updatedEpisode.Title = gotEpisodeBeforeUpdate.Title
			}
			if u.req.Descriptions.Valid {
				updatedEpisode.Descriptions = u.req.Descriptions
			} else {
				updatedEpisode.Descriptions = gotEpisodeBeforeUpdate.Descriptions
			}
			if u.req.DateReleased.Valid {
				updatedEpisode.DateReleased = u.req.DateReleased.Time
			} else {
				updatedEpisode.DateReleased = gotEpisodeBeforeUpdate.DateReleased
			}
			if u.req.Duration.Valid {
				updatedEpisode.Duration = u.req.Duration
			} else {
				updatedEpisode.Duration = gotEpisodeBeforeUpdate.Duration
			}
			updatedEpisode.Invalidation = null.String{}
			updatedEpisode.ContributedBy = defaults.user.id
			updatedEpisode.ContributedAt = gotEpisodeAfterUpdate.ContributedAt

			testutils.SetTimeLocation(
				&updatedEpisode.DateReleased,
				gotEpisodeAfterUpdate.DateReleased.Location(),
			)

			require.Equal(updatedEpisode, gotEpisodeAfterUpdate)
		})
	}
}

func TestHandleEpisodeInvalidate(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	server, appInstance, defaults, teardown := setup(OptEnableDefaultSeries)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/authorized/series/{id}/season/{se}/episode/{ep}/invalidate"
	method := http.MethodPost

	invalidationRequest := &dto.InvalidationRequest{
		Invalidation: "invalidation",
	}

	var (
		seasonNumber  = 1
		episodeNumber = 1
	)

	// invalid id
	e.Request(method, path).
		WithPath("id", -1).
		WithPath("se", -1).
		WithPath("ep", -1).
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
				"season_number": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{"threshold": 1},
				),
				"episode_number": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{"threshold": 1},
				),
			}.Error(),
		))

	// invalid request
	e.Request(method, path).
		WithPath("id", 1).
		WithPath("se", 1).
		WithPath("ep", 1).
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

	// episode not found
	e.Request(method, path).
		WithPath("id", 999).
		WithPath("se", 999).
		WithPath("ep", 999).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(invalidationRequest).
		Expect().
		Status(http.StatusNotFound).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			http.StatusText(http.StatusNotFound),
		))

	// invalidate episode
	episodeCreateReq := &dto.EpisodePutRequest{
		Title:        "episode",
		DateReleased: testutils.Date(1900, 3, 14),
	}
	err := appInstance.EpisodePut(
		ctx,
		defaults.series.id, seasonNumber, episodeNumber,
		defaults.user.id,
		episodeCreateReq,
	)
	require.NoError(err)

	e.Request(method, path).
		WithPath("id", defaults.series.id).
		WithPath("se", seasonNumber).
		WithPath("ep", episodeNumber).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(invalidationRequest).
		Expect().
		Status(http.StatusOK).
		NoContent()

	// check episode invalidated
	gotInvalidatedEpisode, err := appInstance.EpisodeGet(
		ctx,
		defaults.series.id,
		seasonNumber,
		episodeNumber,
	)
	require.NoError(err)

	testutils.SetTimeLocation(
		&episodeCreateReq.DateReleased,
		gotInvalidatedEpisode.DateReleased.Location(),
	)

	require.Equal(
		&models.Film{
			ID:            gotInvalidatedEpisode.ID,
			Title:         episodeCreateReq.Title,
			Descriptions:  episodeCreateReq.Descriptions,
			DateReleased:  episodeCreateReq.DateReleased,
			Duration:      episodeCreateReq.Duration,
			SeriesID:      null.IntFrom(defaults.series.id),
			SeasonNumber:  null.IntFrom(seasonNumber),
			EpisodeNumber: null.IntFrom(episodeNumber),
			Invalidation:  null.StringFrom(invalidationRequest.Invalidation),
			ContributedBy: defaults.user.id,
			ContributedAt: gotInvalidatedEpisode.ContributedAt,
		},
		gotInvalidatedEpisode,
	)
}

func TestHandleEpisodesInvalidateAllBySeason(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	server, appInstance, defaults, teardown := setup(OptEnableDefaultSeries)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/authorized/series/{id}/season/{se}/episode/invalidate"
	method := http.MethodPost

	invalidationRequest := &dto.InvalidationRequest{
		Invalidation: "invalidation",
	}

	seasonNumber := 1

	// invalid id
	e.Request(method, path).
		WithPath("id", -1).
		WithPath("se", -1).
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
				"season_number": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{"threshold": 1},
				),
			}.Error(),
		))

	// invalid request
	e.Request(method, path).
		WithPath("id", 1).
		WithPath("se", 1).
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

	// not found
	e.Request(method, path).
		WithPath("id", 999).
		WithPath("se", 999).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(invalidationRequest).
		Expect().
		Status(http.StatusNotFound).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			http.StatusText(http.StatusNotFound),
		))

	// invalidate all episodes
	episodePutAllReq := &dto.EpisodesPutAllBySeasonRequest{
		Episodes: []*dto.EpisodePutRequest{
			{
				Title:        "episode1",
				DateReleased: testutils.Date(2000, 1, 2),
			},
			{
				Title:        "episode2",
				DateReleased: testutils.Date(2001, 1, 2),
			},
			{
				Title:        "episode3",
				DateReleased: testutils.Date(2002, 1, 2),
			},
			{
				Title:        "episode4",
				DateReleased: testutils.Date(2003, 1, 2),
			},
			{
				Title:        "episode5",
				DateReleased: testutils.Date(2004, 1, 2),
			},
		},
	}
	err := appInstance.EpisodesPutAllBySeason(
		ctx,
		defaults.series.id, seasonNumber,
		defaults.user.id,
		episodePutAllReq,
	)
	require.NoError(err)

	invalidationTime := time.Now()

	e.Request(method, path).
		WithPath("id", defaults.series.id).
		WithPath("se", seasonNumber).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(invalidationRequest).
		Expect().
		Status(http.StatusOK).
		NoContent()

	// check episodes invalidated
	gotInvalidatedEpisodes, total, err := appInstance.EpisodesGetAllBySeason(
		ctx,
		defaults.series.id,
		seasonNumber,
		query.SortOrderOptions{
			SortOrder: "asc",
			Offset:    0,
			Limit:     math.MaxInt,
		},
	)
	require.NoError(err)
	require.Equal(len(episodePutAllReq.Episodes), total)

	for i, ep := range episodePutAllReq.Episodes {
		testutils.SetTimeLocation(
			&ep.DateReleased,
			gotInvalidatedEpisodes[i].DateReleased.Location(),
		)

		require.GreaterOrEqual(
			gotInvalidatedEpisodes[i].ContributedAt,
			invalidationTime,
		)

		episodeNumber := i + 1

		require.Equal(
			&models.Film{
				ID:            gotInvalidatedEpisodes[i].ID,
				Title:         ep.Title,
				Descriptions:  ep.Descriptions,
				DateReleased:  ep.DateReleased,
				Duration:      ep.Duration,
				SeriesID:      null.IntFrom(defaults.series.id),
				SeasonNumber:  null.IntFrom(seasonNumber),
				EpisodeNumber: null.IntFrom(episodeNumber),
				Invalidation: null.StringFrom(
					invalidationRequest.Invalidation,
				),
				ContributedBy: defaults.user.id,
				ContributedAt: gotInvalidatedEpisodes[i].ContributedAt,
			},
			gotInvalidatedEpisodes[i],
		)
	}
}

func TestHandleEpisodeAuditsGetAll(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	server, appInstance, defaults, teardown := setup(OptEnableDefaultSeries)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/authorized/series/{id}/season/{se}/episode/{ep}/audits"
	method := http.MethodGet

	var (
		seasonNumber  = 1
		episodeNumber = 1
	)

	// invalid id
	e.Request(method, path).
		WithPath("id", -1).
		WithPath("se", -1).
		WithPath("ep", -1).
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
				"season_number": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{"threshold": 1},
				),
				"episode_number": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{"threshold": 1},
				),
			}.Error(),
		))

	// invalid query
	e.Request(method, path).
		WithPath("id", 1).
		WithPath("se", 1).
		WithPath("ep", 1).
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

	// episode not found
	e.Request(method, path).
		WithPath("id", 999).
		WithPath("se", 999).
		WithPath("ep", 999).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusNotFound).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			http.StatusText(http.StatusNotFound),
		))

	putTime := time.Now()

	// add episode
	episodePutReq := &dto.EpisodePutRequest{
		Title:        "episode",
		DateReleased: testutils.Date(2000, 1, 1),
	}
	err := appInstance.EpisodePut(
		ctx,
		defaults.series.id, seasonNumber, episodeNumber,
		defaults.user.id,
		episodePutReq,
	)
	require.NoError(err)

	gotEpisode, err := appInstance.EpisodeGet(
		ctx,
		defaults.series.id,
		seasonNumber,
		episodeNumber,
	)
	require.NoError(err)

	require.GreaterOrEqual(gotEpisode.ContributedAt, putTime)

	expEpisodeUpdateAudit := &models.FilmsAudit{
		ID:            gotEpisode.ID,
		Title:         episodePutReq.Title,
		Descriptions:  episodePutReq.Descriptions,
		DateReleased:  episodePutReq.DateReleased,
		Duration:      episodePutReq.Duration,
		SeriesID:      null.IntFrom(defaults.series.id),
		SeasonNumber:  null.IntFrom(seasonNumber),
		EpisodeNumber: null.IntFrom(episodeNumber),
		Invalidation:  null.String{},
		ContributedBy: defaults.user.id,
		ContributedAt: gotEpisode.ContributedAt,
	}

	// no audits
	e.Request(method, path).
		WithPath("id", defaults.series.id).
		WithPath("se", seasonNumber).
		WithPath("ep", episodeNumber).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Equal(response.Paginated(
			config.Config.Validation.Pagination.Page.MinValue,
			config.Config.Validation.Pagination.PageSize.DefaultValue,
			([]*models.FilmsAudit)(nil),
			0),
		)

	updateTime := time.Now()

	// update the episode
	episodeUpdateReq := &dto.EpisodeUpdateRequest{
		Title:        null.StringFrom("updated title"),
		Descriptions: null.StringFrom("updated descriptions"),
		DateReleased: null.TimeFrom(
			testutils.Date(2005, 11, 14),
		),
		Duration: null.IntFrom(10 * 60),
	}
	err = appInstance.EpisodeUpdate(
		ctx,
		defaults.series.id, seasonNumber, episodeNumber,
		defaults.user.id,
		episodeUpdateReq,
	)
	require.NoError(err)

	gotEpisode, err = appInstance.EpisodeGet(
		ctx,
		defaults.series.id,
		seasonNumber,
		episodeNumber,
	)
	require.NoError(err)

	require.GreaterOrEqual(gotEpisode.ContributedAt, updateTime)

	expEpisodeInvalidationAudit := &models.FilmsAudit{
		ID:            gotEpisode.ID,
		Title:         episodeUpdateReq.Title.String,
		Descriptions:  episodeUpdateReq.Descriptions,
		DateReleased:  episodeUpdateReq.DateReleased.Time,
		Duration:      episodeUpdateReq.Duration,
		SeriesID:      null.IntFrom(defaults.series.id),
		SeasonNumber:  null.IntFrom(seasonNumber),
		EpisodeNumber: null.IntFrom(episodeNumber),
		Invalidation:  null.String{},
		ContributedBy: defaults.user.id,
		ContributedAt: gotEpisode.ContributedAt,
	}

	// get update audits
	e.Request(method, path).
		WithPath("id", defaults.series.id).
		WithPath("se", seasonNumber).
		WithPath("ep", episodeNumber).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Equal(response.Paginated(
			config.Config.Validation.Pagination.Page.MinValue,
			config.Config.Validation.Pagination.PageSize.DefaultValue,
			[]*models.FilmsAudit{expEpisodeUpdateAudit},
			1,
		))

	// invalidate the episode
	err = appInstance.EpisodeInvalidate(
		ctx,
		defaults.series.id, seasonNumber, episodeNumber,
		defaults.user.id,
		&dto.InvalidationRequest{Invalidation: "invalidation"},
	)
	require.NoError(err)

	// get invalidation audits
	e.Request(method, path).
		WithPath("id", defaults.series.id).
		WithPath("se", seasonNumber).
		WithPath("ep", episodeNumber).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Equal(response.Paginated(
			config.Config.Validation.Pagination.Page.MinValue,
			config.Config.Validation.Pagination.PageSize.DefaultValue,
			[]*models.FilmsAudit{
				expEpisodeInvalidationAudit,
				expEpisodeUpdateAudit,
			},
			2,
		))
}
