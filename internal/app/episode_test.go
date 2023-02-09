package app_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	_ "unsafe"

	"github.com/aria3ppp/watchlist-server/internal/app"
	"github.com/aria3ppp/watchlist-server/internal/dto"
	"github.com/aria3ppp/watchlist-server/internal/models"
	"github.com/aria3ppp/watchlist-server/internal/query"
	"github.com/aria3ppp/watchlist-server/internal/repo"
	"github.com/aria3ppp/watchlist-server/internal/repo/mock_repo"
	"github.com/aria3ppp/watchlist-server/internal/testutils"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/null/v8"
)

func TestEpisodeGet(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var (
		seriesID      = 1
		seasonNumber  = 1
		episodeNumber = 1
		expError      = errors.New("error")
		expEpisode    = &models.Film{Title: "episode"}
	)

	type GetExp struct {
		episode *models.Film
		err     error
	}
	type Get struct {
		exp GetExp
	}
	type Exp struct {
		episode *models.Film
		err     error
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
					episode: nil,
					err:     expError,
				},
			},
			exp: Exp{
				episode: nil,
				err:     expError,
			},
		},
		{
			name: "not found",
			get: Get{
				exp: GetExp{
					episode: nil,
					err:     repo.ErrNoRecord,
				},
			},
			exp: Exp{
				episode: nil,
				err:     app.ErrNotFound,
			},
		},
		{
			name: "ok",
			get: Get{
				exp: GetExp{
					episode: expEpisode,
					err:     nil,
				},
			},
			exp: Exp{
				episode: expEpisode,
				err:     nil,
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
				EpisodeGet(
					ctx,
					seriesID,
					seasonNumber,
					episodeNumber,
				).
				Return(tc.get.exp.episode, tc.get.exp.err)

			app := app.NewApplication(mockRepo, nil, nil, nil, nil)

			episode, err := app.EpisodeGet(
				ctx,
				seriesID,
				seasonNumber,
				episodeNumber,
			)
			require.Equal(tc.exp.err, err)
			require.Equal(tc.exp.episode, episode)
		})
	}
}

func TestEpisodesGetAllBySeries(t *testing.T) {
	t.Parallel()

	var (
		ctx = context.Background()

		seriesID     = 1
		queryOptions = query.SortOrderOptions{
			SortOrder: "asc",
			Offset:    0,
			Limit:     50,
		}

		expEpisodes                    = []*models.Film{{Title: "episode"}}
		expTotal                       = 1000
		expSeriesGetError              = errors.New("SeriesGet error")
		expEpisodesGetAllBySeriesError = errors.New(
			"EpisodesGetAllBySeries error",
		)
		expEpisodesCountBySeriesError = errors.New(
			"EpisodesCountBySeries error",
		)
	)

	type GetAllBySeriesExp struct {
		episodes []*models.Film
		err      error
	}
	type CountBySeriesExp struct {
		total int
		err   error
	}
	type TxExp struct {
		err error
	}
	type Tx struct {
		exp TxExp
	}
	type GetSeriesExp struct {
		err error
	}
	type GetSeries struct {
		exp GetSeriesExp
	}
	type GetAllBySeries struct {
		exp GetAllBySeriesExp
	}
	type CountBySeries struct {
		exp CountBySeriesExp
	}
	type Exp struct {
		episodes []*models.Film
		total    int
		err      error
	}
	type TestCase struct {
		name           string
		tx             Tx
		getSeries      GetSeries
		getAllBySeries GetAllBySeries
		countBySeries  CountBySeries
		exp            Exp
	}

	testCases := []TestCase{
		{
			name: "series not found",
			tx: Tx{
				exp: TxExp{
					err: app.ErrNotFound,
				},
			},
			getSeries: GetSeries{
				exp: GetSeriesExp{
					err: repo.ErrNoRecord,
				},
			},
			exp: Exp{
				episodes: nil,
				total:    0,
				err:      app.ErrNotFound,
			},
		},

		{
			name: "SeriesGet error",
			tx: Tx{
				exp: TxExp{
					err: expSeriesGetError,
				},
			},
			getSeries: GetSeries{
				exp: GetSeriesExp{
					err: expSeriesGetError,
				},
			},
			exp: Exp{
				episodes: nil,
				total:    0,
				err:      expSeriesGetError,
			},
		},

		{
			name: "EpisodesGetAllBySeries error",
			tx: Tx{
				exp: TxExp{
					err: expEpisodesGetAllBySeriesError,
				},
			},
			getSeries: GetSeries{
				exp: GetSeriesExp{
					err: nil,
				},
			},
			getAllBySeries: GetAllBySeries{
				exp: GetAllBySeriesExp{
					episodes: nil,
					err:      expEpisodesGetAllBySeriesError,
				},
			},
			exp: Exp{
				episodes: nil,
				total:    0,
				err:      expEpisodesGetAllBySeriesError,
			},
		},

		{
			name: "EpisodesCountBySeries error",
			tx: Tx{
				exp: TxExp{
					err: expEpisodesCountBySeriesError,
				},
			},
			getSeries: GetSeries{
				exp: GetSeriesExp{
					err: nil,
				},
			},
			getAllBySeries: GetAllBySeries{
				exp: GetAllBySeriesExp{
					episodes: expEpisodes,
					err:      nil,
				},
			},
			countBySeries: CountBySeries{
				exp: CountBySeriesExp{
					total: 0,
					err:   expEpisodesCountBySeriesError,
				},
			},
			exp: Exp{
				episodes: nil,
				total:    0,
				err:      expEpisodesCountBySeriesError,
			},
		},

		{
			name: "ok",
			tx: Tx{
				exp: TxExp{
					err: nil,
				},
			},
			getSeries: GetSeries{
				exp: GetSeriesExp{
					err: nil,
				},
			},
			getAllBySeries: GetAllBySeries{
				exp: GetAllBySeriesExp{
					episodes: expEpisodes,
					err:      nil,
				},
			},
			countBySeries: CountBySeries{
				exp: CountBySeriesExp{
					total: expTotal,
					err:   nil,
				},
			},
			exp: Exp{
				episodes: expEpisodes,
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

			getSeriesCall := mockRepo.EXPECT().
				SeriesGet(ctx, seriesID).
				Return(&models.Series{}, tc.getSeries.exp.err).
				After(txCall)

			if tc.getSeries.exp.err == nil {
				getAllCall := mockRepo.EXPECT().
					EpisodesGetAllBySeries(ctx, seriesID, queryOptions).
					Return(tc.getAllBySeries.exp.episodes, tc.getAllBySeries.exp.err).
					After(getSeriesCall)

				if tc.getAllBySeries.exp.err == nil {
					mockRepo.EXPECT().
						EpisodesCountBySeries(ctx, seriesID).
						Return(tc.countBySeries.exp.total, tc.countBySeries.exp.err).
						After(getAllCall)
				}
			}

			app := app.NewApplication(mockRepo, nil, nil, nil, nil)

			episodes, total, err := app.EpisodesGetAllBySeries(
				ctx,
				seriesID,
				queryOptions,
			)
			require.Equal(tc.exp.err, err)
			require.Equal(tc.exp.episodes, episodes)
			require.Equal(tc.exp.total, total)
		})
	}
}

func TestEpisodesGetAllBySeason(t *testing.T) {
	t.Parallel()

	var (
		ctx = context.Background()

		seriesID     = 1
		seasonNumber = 1
		queryOptions = query.SortOrderOptions{
			SortOrder: "asc",
			Offset:    0,
			Limit:     50,
		}

		expEpisodes                    = []*models.Film{{Title: "episode"}}
		expTotal                       = 1000
		expSeriesGetError              = errors.New("SeriesGet error")
		expEpisodesGetAllBySeasonError = errors.New(
			"EpisodesGetAllBySeason error",
		)
		expEpisodesCountBySeasonError = errors.New(
			"EpisodesCountBySeason error",
		)
	)

	type GetAllBySeasonExp struct {
		episodes []*models.Film
		err      error
	}
	type CountBySeasonExp struct {
		total int
		err   error
	}
	type TxExp struct {
		err error
	}
	type Tx struct {
		exp TxExp
	}
	type GetSeriesExp struct {
		err error
	}
	type GetSeries struct {
		exp GetSeriesExp
	}
	type GetAllBySeason struct {
		exp GetAllBySeasonExp
	}
	type CountBySeason struct {
		exp CountBySeasonExp
	}
	type Exp struct {
		episodes []*models.Film
		total    int
		err      error
	}
	type TestCase struct {
		name           string
		tx             Tx
		getSeries      GetSeries
		getAllBySeason GetAllBySeason
		countBySeason  CountBySeason
		exp            Exp
	}

	testCases := []TestCase{
		{
			name: "series not found",
			tx: Tx{
				exp: TxExp{
					err: app.ErrNotFound,
				},
			},
			getSeries: GetSeries{
				exp: GetSeriesExp{
					err: repo.ErrNoRecord,
				},
			},
			exp: Exp{
				episodes: nil,
				total:    0,
				err:      app.ErrNotFound,
			},
		},

		{
			name: "SeriesGet error",
			tx: Tx{
				exp: TxExp{
					err: expSeriesGetError,
				},
			},
			getSeries: GetSeries{
				exp: GetSeriesExp{
					err: expSeriesGetError,
				},
			},
			exp: Exp{
				episodes: nil,
				total:    0,
				err:      expSeriesGetError,
			},
		},

		{
			name: "EpisodesGetAllBySeason error",
			tx: Tx{
				exp: TxExp{
					err: expEpisodesGetAllBySeasonError,
				},
			},
			getSeries: GetSeries{
				exp: GetSeriesExp{
					err: nil,
				},
			},
			getAllBySeason: GetAllBySeason{
				exp: GetAllBySeasonExp{
					episodes: nil,
					err:      expEpisodesGetAllBySeasonError,
				},
			},
			exp: Exp{
				episodes: nil,
				total:    0,
				err:      expEpisodesGetAllBySeasonError,
			},
		},

		{
			name: "EpisodesCountBySeason error",
			tx: Tx{
				exp: TxExp{
					err: expEpisodesCountBySeasonError,
				},
			},
			getSeries: GetSeries{
				exp: GetSeriesExp{
					err: nil,
				},
			},
			getAllBySeason: GetAllBySeason{
				exp: GetAllBySeasonExp{
					episodes: expEpisodes,
					err:      nil,
				},
			},
			countBySeason: CountBySeason{
				exp: CountBySeasonExp{
					total: 0,
					err:   expEpisodesCountBySeasonError,
				},
			},
			exp: Exp{
				episodes: nil,
				total:    0,
				err:      expEpisodesCountBySeasonError,
			},
		},

		{
			name: "ok",
			tx: Tx{
				exp: TxExp{
					err: nil,
				},
			},
			getSeries: GetSeries{
				exp: GetSeriesExp{
					err: nil,
				},
			},
			getAllBySeason: GetAllBySeason{
				exp: GetAllBySeasonExp{
					episodes: expEpisodes,
					err:      nil,
				},
			},
			countBySeason: CountBySeason{
				exp: CountBySeasonExp{
					total: expTotal,
					err:   nil,
				},
			},
			exp: Exp{
				episodes: expEpisodes,
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

			getSeriesCall := mockRepo.EXPECT().
				SeriesGet(ctx, seriesID).
				Return(&models.Series{}, tc.getSeries.exp.err).
				After(txCall)

			if tc.getSeries.exp.err == nil {
				getAllCall := mockRepo.EXPECT().
					EpisodesGetAllBySeason(ctx, seriesID, seasonNumber, queryOptions).
					Return(tc.getAllBySeason.exp.episodes, tc.getAllBySeason.exp.err).
					After(getSeriesCall)

				if tc.getAllBySeason.exp.err == nil {
					mockRepo.EXPECT().
						EpisodesCountBySeason(ctx, seriesID, seasonNumber).
						Return(tc.countBySeason.exp.total, tc.countBySeason.exp.err).
						After(getAllCall)
				}
			}

			app := app.NewApplication(mockRepo, nil, nil, nil, nil)

			episodes, total, err := app.EpisodesGetAllBySeason(
				ctx,
				seriesID,
				seasonNumber,
				queryOptions,
			)
			require.Equal(tc.exp.err, err)
			require.Equal(tc.exp.episodes, episodes)
			require.Equal(tc.exp.total, total)
		})
	}
}

func TestEpisodePut(t *testing.T) {
	t.Parallel()

	var (
		ctx = context.Background()

		seriesID      = 1
		seasonNumber  = 1
		episodeNumber = 1
		contributorID = 1
		expSeries     = &models.Series{
			ID:    seriesID,
			Title: "series",
		}
		req = &dto.EpisodePutRequest{
			Title: "episode",
		}
		expSeriesGetError  = errors.New("SeriesGet error")
		expEpisodePutError = errors.New("EpisodePut error")
	)

	type TxExp struct {
		err error
	}
	type Tx struct {
		exp TxExp
	}
	type GetSeriesExp struct {
		series *models.Series
		err    error
	}
	type GetSeries struct {
		exp GetSeriesExp
	}
	type PutExp struct {
		err error
	}
	type Put struct {
		exp PutExp
	}
	type Exp struct {
		err error
	}
	type TestCase struct {
		name      string
		tx        Tx
		getSeries GetSeries
		put       Put
		exp       Exp
	}

	testCases := []TestCase{
		{
			name: "not found",
			tx: Tx{
				exp: TxExp{
					err: app.ErrNotFound,
				},
			},
			getSeries: GetSeries{
				exp: GetSeriesExp{
					series: nil,
					err:    repo.ErrNoRecord,
				},
			},
			exp: Exp{
				err: app.ErrNotFound,
			},
		},
		{
			name: "SeriesGet error",
			tx: Tx{
				exp: TxExp{
					err: expSeriesGetError,
				},
			},
			getSeries: GetSeries{
				exp: GetSeriesExp{
					series: nil,
					err:    expSeriesGetError,
				},
			},
			exp: Exp{
				err: expSeriesGetError,
			},
		},
		{
			name: "EpisodePut error",
			tx: Tx{
				exp: TxExp{
					err: expEpisodePutError,
				},
			},
			getSeries: GetSeries{
				exp: GetSeriesExp{
					series: expSeries,
					err:    nil,
				},
			},
			put: Put{
				exp: PutExp{err: expEpisodePutError},
			},
			exp: Exp{
				err: expEpisodePutError,
			},
		},

		{
			name: "ok",
			tx: Tx{
				exp: TxExp{
					err: nil,
				},
			},
			getSeries: GetSeries{
				exp: GetSeriesExp{
					series: expSeries,
					err:    nil,
				},
			},
			put: Put{
				exp: PutExp{err: nil},
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

			txCall := mockRepo.EXPECT().
				Tx(ctx, nil, gomock.Any()).
				Do(func(ctx context.Context, opts *sql.TxOptions, fn func(_ context.Context, _ repo.Service) error) {
					fn(ctx, mockRepo)
				}).
				Return(tc.tx.exp.err)

			seriesGetCall := mockRepo.EXPECT().
				SeriesGet(ctx, seriesID).
				Return(tc.getSeries.exp.series, tc.getSeries.exp.err).
				After(txCall)

			if tc.getSeries.exp.err == nil {
				mockRepo.EXPECT().
					EpisodePut(
						ctx,
						seriesID,
						seasonNumber,
						episodeNumber,
						contributorID,
						&models.Film{
							Title:        req.Title,
							Descriptions: req.Descriptions,
							DateReleased: req.DateReleased,
							Duration:     req.Duration,
						},
					).
					Return(tc.put.exp.err).
					After(seriesGetCall)
			}

			app := app.NewApplication(mockRepo, nil, nil, nil, nil)

			err := app.EpisodePut(
				ctx,
				seriesID,
				seasonNumber,
				episodeNumber,
				contributorID,
				req,
			)
			require.Equal(tc.exp.err, err)
		})
	}
}

func TestEpisodesPutAllBySeason(t *testing.T) {
	t.Parallel()

	var (
		ctx = context.Background()

		seriesID      = 1
		seasonNumber  = 1
		contributorID = 1
		expSeries     = &models.Series{
			ID:    seriesID,
			Title: "series",
		}
		req = &dto.EpisodesPutAllBySeasonRequest{
			Episodes: []*dto.EpisodePutRequest{
				{
					Title:        "episode",
					DateReleased: testutils.Date(2000, 1, 2),
				},
			},
		}
		expSeriesGetError = errors.New("SeriesGet error")
		expReplaceError   = errors.New("replace error")
	)

	type TxExp struct {
		err error
	}
	type Tx struct {
		exp TxExp
	}
	type SeriesGetExp struct {
		series *models.Series
		err    error
	}
	type SeriesGet struct {
		exp SeriesGetExp
	}
	type ReplaceEpisodesExp struct {
		err error
	}
	type ReplaceEpisodes struct {
		exp ReplaceEpisodesExp
	}
	type Exp struct {
		err error
	}
	type TestCase struct {
		name            string
		tx              Tx
		serieGet        SeriesGet
		replaceEpisodes ReplaceEpisodes
		exp             Exp
	}

	testCases := []TestCase{
		{
			name: "SeriesGet error",
			tx: Tx{
				exp: TxExp{
					err: expSeriesGetError,
				},
			},
			serieGet: SeriesGet{
				exp: SeriesGetExp{err: expSeriesGetError},
			},
			exp: Exp{
				err: expSeriesGetError,
			},
		},

		{
			name: "not found",
			tx: Tx{
				exp: TxExp{
					err: app.ErrNotFound,
				},
			},
			serieGet: SeriesGet{
				exp: SeriesGetExp{err: repo.ErrNoRecord},
			},
			exp: Exp{
				err: app.ErrNotFound,
			},
		},

		{
			name: "last element replace error",
			tx: Tx{
				exp: TxExp{
					err: expReplaceError,
				},
			},
			serieGet: SeriesGet{
				exp: SeriesGetExp{series: expSeries},
			},
			replaceEpisodes: ReplaceEpisodes{
				exp: ReplaceEpisodesExp{
					err: expReplaceError,
				},
			},
			exp: Exp{
				err: expReplaceError,
			},
		},

		{
			name: "ok",
			tx: Tx{
				exp: TxExp{
					err: nil,
				},
			},
			serieGet: SeriesGet{
				exp: SeriesGetExp{series: expSeries},
			},
			replaceEpisodes: ReplaceEpisodes{
				exp: ReplaceEpisodesExp{
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

			txCall := mockRepo.EXPECT().
				Tx(ctx, nil, gomock.Any()).
				Do(func(ctx context.Context, opts *sql.TxOptions, fn func(_ context.Context, _ repo.Service) error) {
					fn(ctx, mockRepo)
				}).
				Return(tc.tx.exp.err)

			prevCall := mockRepo.EXPECT().
				SeriesGet(ctx, seriesID).
				Return(tc.serieGet.exp.series, tc.serieGet.exp.err).
				After(txCall)

			if tc.serieGet.exp.err == nil && len(req.Episodes) > 0 {
				for i, expEpisode := range req.Episodes {
					episodeNumber := i + 1
					episodePutReq := &models.Film{
						Title:        expEpisode.Title,
						Descriptions: expEpisode.Descriptions,
						DateReleased: expEpisode.DateReleased,
						Duration:     expEpisode.Duration,
					}
					if i == len(req.Episodes)-1 {
						// set last element expected error (nil or non nil)
						mockRepo.EXPECT().
							EpisodePut(
								ctx,
								seriesID, seasonNumber, episodeNumber,
								contributorID,
								episodePutReq,
							).
							Return(tc.replaceEpisodes.exp.err).
							After(prevCall)
					} else {
						// let other element return nil error
						prevCall = mockRepo.EXPECT().
							EpisodePut(
								ctx,
								seriesID, seasonNumber, episodeNumber,
								contributorID,
								episodePutReq,
							).
							Return(nil).
							After(prevCall)
					}
				}
			}

			app := app.NewApplication(mockRepo, nil, nil, nil, nil)

			err := app.EpisodesPutAllBySeason(
				ctx,
				seriesID,
				seasonNumber,
				contributorID,
				req,
			)
			require.Equal(tc.exp.err, err)
		})
	}
}

//go:linkname episodeUpdateRequestToValidMap github.com/aria3ppp/watchlist-server/internal/app.episodeUpdateRequestToValidMap
func episodeUpdateRequestToValidMap(
	req *dto.EpisodeUpdateRequest,
) map[string]any

func TestEpisodeUpdate(t *testing.T) {
	t.Parallel()

	var (
		ctx = context.Background()

		seriesID      = 1
		seasonNumber  = 1
		episodeNumber = 1
		contributorID = 1
		req           = &dto.EpisodeUpdateRequest{
			Title:        null.StringFrom("episode"),
			Descriptions: null.StringFrom("descriptions"),
			DateReleased: null.TimeFrom(
				testutils.Date(1994, 6, 1),
			),
			Duration: null.IntFrom(3 * 60),
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
				EpisodeUpdate(ctx, seriesID, seasonNumber, episodeNumber, contributorID, episodeUpdateRequestToValidMap(req)).
				Return(tc.update.exp.err)

			app := app.NewApplication(mockRepo, nil, nil, nil, nil)

			err := app.EpisodeUpdate(
				ctx,
				seriesID,
				seasonNumber,
				episodeNumber,
				contributorID,
				req,
			)
			require.Equal(tc.exp.err, err)
		})
	}
}

func TestEpisodeInvalidate(t *testing.T) {
	t.Parallel()

	var (
		ctx = context.Background()

		seriesID      = 1
		seasonNumber  = 1
		episodeNumber = 1
		contributorID = 1
		req           = &dto.InvalidationRequest{
			Invalidation: "invalidation",
		}

		expError = errors.New("error")
	)

	type SeriesInvalidateExp struct {
		err error
	}
	type EpisodeInvalidate struct {
		exp SeriesInvalidateExp
	}
	type Exp struct {
		err error
	}
	type TestCase struct {
		name              string
		episodeInvalidate EpisodeInvalidate
		exp               Exp
	}

	testCases := []TestCase{
		{
			name: "error",
			episodeInvalidate: EpisodeInvalidate{
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
			episodeInvalidate: EpisodeInvalidate{
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
			episodeInvalidate: EpisodeInvalidate{
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
				EpisodeUpdate(ctx, seriesID, seasonNumber, episodeNumber, contributorID, map[string]any{
					models.FilmColumns.Invalidation: req.Invalidation,
				}).
				Return(tc.episodeInvalidate.exp.err)

			app := app.NewApplication(mockRepo, nil, nil, nil, nil)

			err := app.EpisodeInvalidate(
				ctx,
				seriesID,
				seasonNumber,
				episodeNumber,
				contributorID,
				req,
			)
			require.Equal(tc.exp.err, err)
		})
	}
}

func TestEpisodesInvalidateAllBySeason(t *testing.T) {
	t.Parallel()

	var (
		ctx = context.Background()

		seriesID      = 1
		seasonNumber  = 1
		contributorID = 1
		req           = &dto.InvalidationRequest{
			Invalidation: "invalidation",
		}

		expError = errors.New("error")
	)

	type SeriesInvalidateExp struct {
		err error
	}
	type EpisodesInvalidateAllBySeason struct {
		exp SeriesInvalidateExp
	}
	type Exp struct {
		err error
	}
	type TestCase struct {
		name                          string
		episodesInvalidateAllBySeason EpisodesInvalidateAllBySeason
		exp                           Exp
	}

	testCases := []TestCase{
		{
			name: "error",
			episodesInvalidateAllBySeason: EpisodesInvalidateAllBySeason{
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
			episodesInvalidateAllBySeason: EpisodesInvalidateAllBySeason{
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
			episodesInvalidateAllBySeason: EpisodesInvalidateAllBySeason{
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
				EpisodesInvalidateAllBySeason(ctx, seriesID, seasonNumber, contributorID, req.Invalidation).
				Return(tc.episodesInvalidateAllBySeason.exp.err)

			app := app.NewApplication(mockRepo, nil, nil, nil, nil)

			err := app.EpisodesInvalidateAllBySeason(
				ctx,
				seriesID,
				seasonNumber,
				contributorID,
				req,
			)
			require.Equal(tc.exp.err, err)
		})
	}
}

func TestEpisodeAuditsGetAll(t *testing.T) {
	t.Parallel()

	var (
		ctx = context.Background()

		seriesID      = 1
		seasonNumber  = 1
		episodeNumber = 1
		queryOptions  = query.SortOrderOptions{
			SortOrder: "desc",
			Offset:    0,
			Limit:     50,
		}

		expEpisode = &models.Film{
			Title: "episode",
		}
		expEpisodeGetError          = errors.New("EpisodeGet error")
		expEpisodeAuditsGetAllError = errors.New("EpisodeAuditsGetAll error")
		expEpisodeAuditsCountError  = errors.New("EpisodeAuditsCount error")
		expAudits                   = []*models.FilmsAudit{{Title: "audit"}}
		expTotal                    = 100
	)

	type EpisodeGetExp struct {
		episode *models.Film
		err     error
	}
	type EpisodeAuditsCountExp struct {
		total int
		err   error
	}
	type EpisodeAuditsGetAllExp struct {
		audits []*models.FilmsAudit
		err    error
	}
	type TxExp struct {
		err error
	}
	type Tx struct {
		exp TxExp
	}
	type EpisodeGet struct {
		exp EpisodeGetExp
	}
	type EpisodeAuditsGetAll struct {
		exp EpisodeAuditsGetAllExp
	}
	type EpisodeAuditsCount struct {
		exp EpisodeAuditsCountExp
	}
	type Exp struct {
		audits []*models.FilmsAudit
		total  int
		err    error
	}
	type TestCase struct {
		name                string
		tx                  Tx
		episodeGet          EpisodeGet
		episodeAuditsGetAll EpisodeAuditsGetAll
		episodeAuditsCount  EpisodeAuditsCount
		exp                 Exp
	}

	testCases := []TestCase{
		{
			name: "not found",
			tx: Tx{
				exp: TxExp{
					err: app.ErrNotFound,
				},
			},
			episodeGet: EpisodeGet{
				exp: EpisodeGetExp{
					episode: nil,
					err:     repo.ErrNoRecord,
				},
			},
			exp: Exp{
				audits: nil,
				total:  0,
				err:    app.ErrNotFound,
			},
		},

		{
			name: "EpisodeGet error",
			tx: Tx{
				exp: TxExp{
					err: expEpisodeGetError,
				},
			},
			episodeGet: EpisodeGet{
				exp: EpisodeGetExp{
					episode: nil,
					err:     expEpisodeGetError,
				},
			},
			exp: Exp{
				audits: nil,
				total:  0,
				err:    expEpisodeGetError,
			},
		},

		{
			name: "EpisodeAuditsGetAll error",
			tx: Tx{
				exp: TxExp{
					err: expEpisodeAuditsGetAllError,
				},
			},
			episodeGet: EpisodeGet{
				exp: EpisodeGetExp{
					episode: expEpisode,
					err:     nil,
				},
			},
			episodeAuditsGetAll: EpisodeAuditsGetAll{
				exp: EpisodeAuditsGetAllExp{
					audits: nil,
					err:    expEpisodeAuditsGetAllError,
				},
			},
			exp: Exp{
				audits: nil,
				total:  0,
				err:    expEpisodeAuditsGetAllError,
			},
		},

		{
			name: "EpisodeAuditsCount error",
			tx: Tx{
				exp: TxExp{
					err: expEpisodeAuditsCountError,
				},
			},
			episodeGet: EpisodeGet{
				exp: EpisodeGetExp{
					episode: expEpisode,
					err:     nil,
				},
			},
			episodeAuditsGetAll: EpisodeAuditsGetAll{
				exp: EpisodeAuditsGetAllExp{
					audits: expAudits,
					err:    nil,
				},
			},
			episodeAuditsCount: EpisodeAuditsCount{
				exp: EpisodeAuditsCountExp{
					total: 0,
					err:   expEpisodeAuditsCountError,
				},
			},
			exp: Exp{
				audits: nil,
				total:  0,
				err:    expEpisodeAuditsCountError,
			},
		},

		{
			name: "ok",
			tx: Tx{
				exp: TxExp{
					err: nil,
				},
			},
			episodeGet: EpisodeGet{
				exp: EpisodeGetExp{
					episode: expEpisode,
					err:     nil,
				},
			},
			episodeAuditsGetAll: EpisodeAuditsGetAll{
				exp: EpisodeAuditsGetAllExp{
					audits: expAudits,
					err:    nil,
				},
			},
			episodeAuditsCount: EpisodeAuditsCount{
				exp: EpisodeAuditsCountExp{
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
				EpisodeGet(ctx, seriesID, seasonNumber, episodeNumber).
				Return(tc.episodeGet.exp.episode, tc.episodeGet.exp.err).
				After(txCall)

			if tc.episodeGet.exp.err == nil {
				seriesAuditsGetAllCall := mockRepo.EXPECT().
					EpisodeAuditsGetAll(ctx, seriesID, seasonNumber, episodeNumber, queryOptions).
					Return(tc.episodeAuditsGetAll.exp.audits, tc.episodeAuditsGetAll.exp.err).
					After(seriesGetCall)

				if tc.episodeAuditsGetAll.exp.err == nil {
					mockRepo.EXPECT().
						EpisodeAuditsCount(ctx, seriesID, seasonNumber, episodeNumber).
						Return(tc.episodeAuditsCount.exp.total, tc.episodeAuditsCount.exp.err).
						After(seriesAuditsGetAllCall)
				}
			}

			app := app.NewApplication(mockRepo, nil, nil, nil, nil)

			audits, total, err := app.EpisodeAuditsGetAll(
				ctx,
				seriesID,
				seasonNumber,
				episodeNumber,
				queryOptions,
			)
			require.Equal(tc.exp.err, err)
			require.Equal(tc.exp.audits, audits)
			require.Equal(tc.exp.total, total)
		})
	}
}
