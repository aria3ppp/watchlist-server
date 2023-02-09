package app_test

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"strings"
	"testing"
	_ "unsafe"

	"github.com/aria3ppp/watchlist-server/internal/app"
	"github.com/aria3ppp/watchlist-server/internal/dto"
	"github.com/aria3ppp/watchlist-server/internal/models"
	"github.com/aria3ppp/watchlist-server/internal/query"
	"github.com/aria3ppp/watchlist-server/internal/repo"
	"github.com/aria3ppp/watchlist-server/internal/repo/mock_repo"
	"github.com/aria3ppp/watchlist-server/internal/search/mock_search"
	"github.com/aria3ppp/watchlist-server/internal/storage"
	"github.com/aria3ppp/watchlist-server/internal/storage/mock_storage"
	"github.com/aria3ppp/watchlist-server/internal/testutils"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/null/v8"
)

func TestSeriesGet(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var (
		id        = 1
		expError  = errors.New("error")
		expSeries = &models.Series{Title: "series"}
	)

	type GetExp struct {
		series *models.Series
		err    error
	}
	type Get struct {
		exp GetExp
	}
	type Exp struct {
		series *models.Series
		err    error
	}
	type TestCase struct {
		name string
		get  Get
		exp  Exp
	}

	testCases := []TestCase{
		{
			name: "error",
			get: Get{
				exp: GetExp{
					series: nil,
					err:    expError,
				},
			},
			exp: Exp{
				series: nil,
				err:    expError,
			},
		},
		{
			name: "not found",
			get: Get{
				exp: GetExp{
					series: nil,
					err:    repo.ErrNoRecord,
				},
			},
			exp: Exp{
				series: nil,
				err:    app.ErrNotFound,
			},
		},
		{
			name: "ok",
			get: Get{
				exp: GetExp{
					series: expSeries,
					err:    nil,
				},
			},
			exp: Exp{
				series: expSeries,
				err:    nil,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require := require.New(t)

			controller := gomock.NewController(t)
			mockRepo := mock_repo.NewMockServiceTx(controller)

			mockRepo.EXPECT().
				SeriesGet(ctx, id).
				Return(tc.get.exp.series, tc.get.exp.err)

			app := app.NewApplication(mockRepo, nil, nil, nil, nil)

			series, err := app.SeriesGet(ctx, id)
			require.Equal(tc.exp.err, err)
			require.Equal(tc.exp.series, series)
		})
	}
}

func TestSeriesesGetAll(t *testing.T) {
	t.Parallel()

	var (
		ctx = context.Background()

		queryOptions = query.Options{
			Offset:    0,
			Limit:     math.MaxInt,
			SortField: models.SeriesColumns.ID,
			SortOrder: "asc",
		}

		expSerieses            = []*models.Series{{Title: "series"}}
		expTotal               = 1000
		expSeriesesGetAllError = errors.New("SeriesesGetAll error")
		expSeriesesCountError  = errors.New("SeriesesCount error")
	)

	type GetAllExp struct {
		serieses []*models.Series
		err      error
	}
	type CountExp struct {
		total int
		err   error
	}
	type TxExp struct {
		err error
	}
	type Tx struct {
		exp TxExp
	}
	type GetAll struct {
		exp GetAllExp
	}
	type Count struct {
		exp CountExp
	}
	type Exp struct {
		serieses []*models.Series
		total    int
		err      error
	}
	type TestCase struct {
		name   string
		tx     Tx
		getAll GetAll
		count  Count
		exp    Exp
	}

	testCases := []TestCase{
		{
			name: "SeriesesGetAll error",
			tx: Tx{
				exp: TxExp{
					err: expSeriesesGetAllError,
				},
			},
			getAll: GetAll{
				exp: GetAllExp{
					serieses: nil,
					err:      expSeriesesGetAllError,
				},
			},
			exp: Exp{
				serieses: nil,
				total:    0,
				err:      expSeriesesGetAllError,
			},
		},

		{
			name: "SeriesesCount error",
			tx: Tx{
				exp: TxExp{
					err: expSeriesesCountError,
				},
			},
			getAll: GetAll{
				exp: GetAllExp{
					serieses: expSerieses,
					err:      nil,
				},
			},
			count: Count{
				exp: CountExp{
					total: 0,
					err:   expSeriesesCountError,
				},
			},
			exp: Exp{
				serieses: nil,
				total:    0,
				err:      expSeriesesCountError,
			},
		},

		{
			name: "ok",
			tx: Tx{
				exp: TxExp{
					err: nil,
				},
			},
			getAll: GetAll{
				exp: GetAllExp{
					serieses: expSerieses,
					err:      nil,
				},
			},
			count: Count{
				exp: CountExp{
					total: expTotal,
					err:   nil,
				},
			},
			exp: Exp{
				serieses: expSerieses,
				total:    expTotal,
				err:      nil,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require := require.New(t)

			controller := gomock.NewController(t)
			mockRepo := mock_repo.NewMockServiceTx(controller)

			txCall := mockRepo.EXPECT().
				Tx(ctx, nil, gomock.Any()).
				Do(func(ctx context.Context, opts *sql.TxOptions, fn func(_ context.Context, _ repo.Service) error) {
					fn(ctx, mockRepo)
				}).
				Return(tc.tx.exp.err)

			getAllCall := mockRepo.EXPECT().
				SeriesesGetAll(ctx, queryOptions).
				Return(tc.getAll.exp.serieses, tc.getAll.exp.err).
				After(txCall)

			if tc.getAll.exp.err == nil {
				mockRepo.EXPECT().
					SeriesesCount(ctx).
					Return(tc.count.exp.total, tc.count.exp.err).
					After(getAllCall)
			}

			app := app.NewApplication(mockRepo, nil, nil, nil, nil)

			serieses, total, err := app.SeriesesGetAll(ctx, queryOptions)
			require.Equal(tc.exp.err, err)
			require.Equal(tc.exp.serieses, serieses)
			require.Equal(tc.exp.total, total)
		})
	}
}

func TestSeriesCreate(t *testing.T) {
	t.Parallel()

	var (
		ctx = context.Background()

		seriesID      = 1
		contributorID = 1
		req           = &dto.SeriesCreateRequest{
			Title: "series",
		}
		expError = errors.New("error")
	)

	type CreateExp struct {
		err error
	}
	type Create struct {
		exp CreateExp
	}
	type Exp struct {
		seriesID int
		err      error
	}
	type TestCase struct {
		name   string
		create Create
		exp    Exp
	}

	testCases := []TestCase{
		{
			name: "error",
			create: Create{
				exp: CreateExp{
					err: expError,
				},
			},
			exp: Exp{
				seriesID: 0,
				err:      expError,
			},
		},

		{
			name: "ok",
			create: Create{
				exp: CreateExp{
					err: nil,
				},
			},
			exp: Exp{
				seriesID: seriesID,
				err:      nil,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require := require.New(t)

			controller := gomock.NewController(t)
			mockRepo := mock_repo.NewMockServiceTx(controller)

			mockRepo.EXPECT().
				SeriesCreate(
					ctx,
					contributorID, &models.Series{
						Title:        req.Title,
						Descriptions: req.Descriptions,
						DateStarted:  req.DateStarted,
						DateEnded:    req.DateEnded,
					},
				).
				Do(func(_ context.Context, _ int, s *models.Series) {
					s.ID = seriesID
				}).
				Return(tc.create.exp.err)

			app := app.NewApplication(mockRepo, nil, nil, nil, nil)

			id, err := app.SeriesCreate(ctx, contributorID, req)
			require.Equal(tc.exp.err, err)
			require.Equal(tc.exp.seriesID, id)
		})
	}
}

//go:linkname seriesUpdateRequestToValidMap github.com/aria3ppp/watchlist-server/internal/app.seriesUpdateRequestToValidMap
func seriesUpdateRequestToValidMap(
	req *dto.SeriesUpdateRequest,
) map[string]any

func TestSeriesUpdate(t *testing.T) {
	t.Parallel()

	var (
		ctx = context.Background()

		seriesID      = 1
		contributorID = 1
		req           = &dto.SeriesUpdateRequest{
			Title:        null.StringFrom("series"),
			Descriptions: null.StringFrom("descriptions"),
			DateStarted: null.TimeFrom(
				testutils.Date(1994, 6, 1),
			),
			DateEnded: null.TimeFrom(
				testutils.Date(2003, 9, 13),
			),
		}
		expError = errors.New("error")
	)

	type UpdateExp struct {
		err error
	}
	type Update struct {
		exp UpdateExp
	}
	type Exp struct {
		err error
	}
	type TestCase struct {
		name   string
		update Update
		exp    Exp
	}

	testCases := []TestCase{
		{
			name: "error",
			update: Update{
				exp: UpdateExp{
					err: expError,
				},
			},
			exp: Exp{
				err: expError,
			},
		},

		{
			name: "not found",
			update: Update{
				exp: UpdateExp{
					err: repo.ErrNoRecord,
				},
			},
			exp: Exp{
				err: app.ErrNotFound,
			},
		},

		{
			name: "ok",
			update: Update{
				exp: UpdateExp{
					err: nil,
				},
			},
			exp: Exp{
				err: nil,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require := require.New(t)

			controller := gomock.NewController(t)
			mockRepo := mock_repo.NewMockServiceTx(controller)

			mockRepo.EXPECT().
				SeriesUpdate(ctx, seriesID, contributorID, seriesUpdateRequestToValidMap(req)).
				Return(tc.update.exp.err)

			app := app.NewApplication(mockRepo, nil, nil, nil, nil)

			err := app.SeriesUpdate(
				ctx,
				seriesID,
				contributorID,
				req,
			)
			require.Equal(tc.exp.err, err)
		})
	}
}

func TestSeriesInvalidate(t *testing.T) {
	t.Parallel()

	var (
		ctx = context.Background()

		seriesID      = 1
		contributorID = 1
		req           = &dto.InvalidationRequest{
			Invalidation: "invalidation",
		}

		expError = errors.New("SeriesDelete error")
	)

	type SeriesInvalidateExp struct {
		err error
	}
	type SeriesInvalidate struct {
		exp SeriesInvalidateExp
	}
	type Exp struct {
		err error
	}
	type TestCase struct {
		name             string
		seriesInvalidate SeriesInvalidate
		exp              Exp
	}

	testCases := []TestCase{
		{
			name: "error",
			seriesInvalidate: SeriesInvalidate{
				exp: SeriesInvalidateExp{
					err: expError,
				},
			},
			exp: Exp{
				err: expError,
			},
		},

		{
			name: "not found",
			seriesInvalidate: SeriesInvalidate{
				exp: SeriesInvalidateExp{
					err: repo.ErrNoRecord,
				},
			},
			exp: Exp{
				err: app.ErrNotFound,
			},
		},

		{
			name: "ok",
			seriesInvalidate: SeriesInvalidate{
				exp: SeriesInvalidateExp{
					err: nil,
				},
			},
			exp: Exp{
				err: nil,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require := require.New(t)

			controller := gomock.NewController(t)
			mockRepo := mock_repo.NewMockServiceTx(controller)

			mockRepo.EXPECT().
				SeriesUpdate(ctx, seriesID, contributorID, map[string]any{
					models.SeriesColumns.Invalidation: req.Invalidation,
				}).
				Return(tc.seriesInvalidate.exp.err)

			app := app.NewApplication(mockRepo, nil, nil, nil, nil)

			err := app.SeriesInvalidate(ctx, seriesID, contributorID, req)
			require.Equal(tc.exp.err, err)
		})
	}
}

func TestSeriesAuditsGetAll(t *testing.T) {
	t.Parallel()

	var (
		ctx = context.Background()

		seriesID     = 1
		queryOptions = query.SortOrderOptions{
			Offset:    0,
			Limit:     math.MaxInt,
			SortOrder: "desc",
		}

		expSeries = &models.Series{
			ID:    seriesID,
			Title: "series",
		}
		expSeriesGetError          = errors.New("SeriesGet error")
		expSeriesAuditsGetAllError = errors.New("SeriesAuditsGetAll error")
		expSeriesAuditsCountError  = errors.New("SeriesAuditsCount error")
		expAudits                  = []*models.SeriesesAudit{{Title: "audit"}}
		expTotal                   = 100
	)

	type SeriesGetExp struct {
		series *models.Series
		err    error
	}
	type SeriesAuditsCountExp struct {
		total int
		err   error
	}
	type SeriesAuditsGetAllExp struct {
		audits []*models.SeriesesAudit
		err    error
	}
	type TxExp struct {
		err error
	}
	type Tx struct {
		exp TxExp
	}
	type SeriesGet struct {
		exp SeriesGetExp
	}
	type SeriesAuditsGetAll struct {
		exp SeriesAuditsGetAllExp
	}
	type SeriesAuditsCount struct {
		exp SeriesAuditsCountExp
	}
	type Exp struct {
		audits []*models.SeriesesAudit
		total  int
		err    error
	}
	type TestCase struct {
		name               string
		tx                 Tx
		seriesGet          SeriesGet
		seriesAuditsGetAll SeriesAuditsGetAll
		seriesAuditsCount  SeriesAuditsCount
		exp                Exp
	}

	testCases := []TestCase{
		{
			name: "not found",
			tx: Tx{
				exp: TxExp{
					err: app.ErrNotFound,
				},
			},
			seriesGet: SeriesGet{
				exp: SeriesGetExp{
					series: nil,
					err:    repo.ErrNoRecord,
				},
			},
			exp: Exp{
				audits: nil,
				total:  0,
				err:    app.ErrNotFound,
			},
		},

		{
			name: "SeriesGet error",
			tx: Tx{
				exp: TxExp{
					err: expSeriesGetError,
				},
			},
			seriesGet: SeriesGet{
				exp: SeriesGetExp{
					series: nil,
					err:    expSeriesGetError,
				},
			},
			exp: Exp{
				audits: nil,
				total:  0,
				err:    expSeriesGetError,
			},
		},

		{
			name: "SeriesAuditsGetAll error",
			tx: Tx{
				exp: TxExp{
					err: expSeriesAuditsGetAllError,
				},
			},
			seriesGet: SeriesGet{
				exp: SeriesGetExp{
					series: expSeries,
					err:    nil,
				},
			},
			seriesAuditsGetAll: SeriesAuditsGetAll{
				exp: SeriesAuditsGetAllExp{
					audits: nil,
					err:    expSeriesAuditsGetAllError,
				},
			},
			exp: Exp{
				audits: nil,
				total:  0,
				err:    expSeriesAuditsGetAllError,
			},
		},

		{
			name: "SeriesAuditsCount error",
			tx: Tx{
				exp: TxExp{
					err: expSeriesAuditsCountError,
				},
			},
			seriesGet: SeriesGet{
				exp: SeriesGetExp{
					series: expSeries,
					err:    nil,
				},
			},
			seriesAuditsGetAll: SeriesAuditsGetAll{
				exp: SeriesAuditsGetAllExp{
					audits: expAudits,
					err:    nil,
				},
			},
			seriesAuditsCount: SeriesAuditsCount{
				exp: SeriesAuditsCountExp{
					total: 0,
					err:   expSeriesAuditsCountError,
				},
			},
			exp: Exp{
				audits: nil,
				total:  0,
				err:    expSeriesAuditsCountError,
			},
		},

		{
			name: "ok",
			tx: Tx{
				exp: TxExp{
					err: nil,
				},
			},
			seriesGet: SeriesGet{
				exp: SeriesGetExp{
					series: expSeries,
					err:    nil,
				},
			},
			seriesAuditsGetAll: SeriesAuditsGetAll{
				exp: SeriesAuditsGetAllExp{
					audits: expAudits,
					err:    nil,
				},
			},
			seriesAuditsCount: SeriesAuditsCount{
				exp: SeriesAuditsCountExp{
					total: expTotal,
					err:   nil,
				},
			},
			exp: Exp{
				audits: expAudits,
				total:  expTotal,
				err:    nil,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require := require.New(t)

			controller := gomock.NewController(t)
			mockRepo := mock_repo.NewMockServiceTx(controller)

			txCall := mockRepo.EXPECT().
				Tx(ctx, nil, gomock.Any()).
				Do(func(ctx context.Context, opts *sql.TxOptions, fn func(_ context.Context, _ repo.Service) error) {
					fn(ctx, mockRepo)
				}).
				Return(tc.tx.exp.err)

			seriesGetCall := mockRepo.EXPECT().
				SeriesGet(ctx, seriesID).
				Return(tc.seriesGet.exp.series, tc.seriesGet.exp.err).
				After(txCall)

			if tc.seriesGet.exp.err == nil {
				seriesAuditsGetAllCall := mockRepo.EXPECT().
					SeriesAuditsGetAll(ctx, seriesID, queryOptions).
					Return(tc.seriesAuditsGetAll.exp.audits, tc.seriesAuditsGetAll.exp.err).
					After(seriesGetCall)

				if tc.seriesAuditsGetAll.exp.err == nil {
					mockRepo.EXPECT().
						SeriesAuditsCount(ctx, seriesID).
						Return(tc.seriesAuditsCount.exp.total, tc.seriesAuditsCount.exp.err).
						After(seriesAuditsGetAllCall)
				}
			}

			app := app.NewApplication(mockRepo, nil, nil, nil, nil)

			audits, total, err := app.SeriesAuditsGetAll(
				ctx,
				seriesID,
				queryOptions,
			)
			require.Equal(tc.exp.err, err)
			require.Equal(tc.exp.audits, audits)
			require.Equal(tc.exp.total, total)
		})
	}
}

func TestSeriesesSearch(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var (
		queryOptions = query.SearchOptions{
			Query: "query",
			From:  0,
			Size:  10,
		}
		expError    = errors.New("error")
		expSerieses = []*models.Series{{Title: "title"}}
		expTotal    = 100
	)

	type SearchExp struct {
		serieses []*models.Series
		total    int
		err      error
	}
	type Search struct {
		exp SearchExp
	}
	type Exp struct {
		serieses []*models.Series
		total    int
		err      error
	}
	type TestCase struct {
		name   string
		search Search
		exp    Exp
	}

	testCases := []TestCase{
		{
			name: "error",
			search: Search{
				exp: SearchExp{
					serieses: nil,
					total:    0,
					err:      expError,
				},
			},
			exp: Exp{
				serieses: nil,
				total:    0,
				err:      expError,
			},
		},
		{
			name: "ok",
			search: Search{
				exp: SearchExp{
					serieses: expSerieses,
					total:    expTotal,
					err:      nil,
				},
			},
			exp: Exp{
				serieses: expSerieses,
				total:    expTotal,
				err:      nil,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require := require.New(t)

			controller := gomock.NewController(t)
			mockSearch := mock_search.NewMockService(controller)

			mockSearch.EXPECT().
				SearchSerieses(ctx, queryOptions).
				Return(tc.search.exp.serieses, tc.exp.total, tc.search.exp.err)

			app := app.NewApplication(nil, nil, mockSearch, nil, nil)

			series, total, err := app.SeriesesSearch(ctx, queryOptions)
			require.Equal(tc.exp.err, err)
			require.Equal(tc.exp.serieses, series)
			require.Equal(tc.exp.total, total)
		})
	}
}

func TestSeriesPutPoster(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var (
		seriesID      = 1
		contributorID = 11
		poster        = strings.NewReader("poster")
		options       = &storage.PutOptions{}

		expUri               = "expected uri :/"
		expPutFileError      = errors.New("PutFile error")
		expSeriesUpdateError = errors.New("SeriesUpdate error")
	)

	type PutFileExp struct {
		uri string
		err error
	}
	type PutFile struct {
		exp PutFileExp
	}
	type UpdateSeriesExp struct {
		err error
	}
	type UpdateSeries struct {
		exp UpdateSeriesExp
	}
	type Exp struct {
		uri string
		err error
	}
	testCases := []struct {
		name         string
		putFile      PutFile
		updateSeries UpdateSeries
		exp          Exp
	}{
		{
			name: "PutFile error",
			putFile: PutFile{
				exp: PutFileExp{
					uri: "",
					err: expPutFileError,
				},
			},
			exp: Exp{
				uri: "",
				err: expPutFileError,
			},
		},
		{
			name: "not found",
			putFile: PutFile{
				exp: PutFileExp{
					uri: expUri,
					err: nil,
				},
			},
			updateSeries: UpdateSeries{
				exp: UpdateSeriesExp{
					err: repo.ErrNoRecord,
				},
			},
			exp: Exp{
				uri: "",
				err: app.ErrNotFound,
			},
		},
		{
			name: "SeriesUpdate error",
			putFile: PutFile{
				exp: PutFileExp{
					uri: expUri,
					err: nil,
				},
			},
			updateSeries: UpdateSeries{
				exp: UpdateSeriesExp{
					err: expSeriesUpdateError,
				},
			},
			exp: Exp{
				uri: "",
				err: expSeriesUpdateError,
			},
		},
		{
			name: "ok",
			putFile: PutFile{
				exp: PutFileExp{
					uri: expUri,
					err: nil,
				},
			},
			updateSeries: UpdateSeries{
				exp: UpdateSeriesExp{
					err: nil,
				},
			},
			exp: Exp{
				uri: expUri,
				err: nil,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require := require.New(t)

			controller := gomock.NewController(t)
			mockStorage := mock_storage.NewMockService(controller)
			mockRepo := mock_repo.NewMockServiceTx(controller)

			putFileCall := mockStorage.EXPECT().
				PutFile(ctx, poster, options).
				Return(tc.putFile.exp.uri, tc.putFile.exp.err)

			if tc.putFile.exp.err == nil {
				mockRepo.EXPECT().
					SeriesUpdate(ctx, seriesID, contributorID, map[string]any{
						models.SeriesColumns.Poster: tc.putFile.exp.uri,
					}).
					Return(tc.updateSeries.exp.err).
					After(putFileCall)
			}

			app := app.NewApplication(mockRepo, nil, nil, nil, mockStorage)

			uri, err := app.SeriesPutPoster(
				ctx,
				seriesID,
				contributorID,
				poster,
				options,
			)
			require.Equal(tc.exp.err, err)
			require.Equal(tc.exp.uri, uri)
		})
	}
}
