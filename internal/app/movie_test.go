package app_test

import (
	"context"
	"errors"
	"testing"
	_ "unsafe"

	"github.com/aria3ppp/watchlist-server/internal/app"
	"github.com/aria3ppp/watchlist-server/internal/dto"
	"github.com/aria3ppp/watchlist-server/internal/models"
	"github.com/aria3ppp/watchlist-server/internal/repo"
	"github.com/aria3ppp/watchlist-server/internal/repo/mock_repo"
	"github.com/aria3ppp/watchlist-server/internal/search/mock_search"
	"github.com/aria3ppp/watchlist-server/internal/testutils"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/null/v8"
)

func TestMovieGet(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var (
		id       = 1
		expError = errors.New("error")
		expMovie = &models.Film{Title: "movie"}
	)

	type GetExp struct {
		movie *models.Film
		err   error
	}
	type Get struct {
		exp GetExp
	}
	type Exp struct {
		movie *models.Film
		err   error
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
					movie: nil,
					err:   expError,
				},
			},
			exp: Exp{
				movie: nil,
				err:   expError,
			},
		},
		{
			name: "not found",
			get: Get{
				exp: GetExp{
					movie: nil,
					err:   repo.ErrNoRecord,
				},
			},
			exp: Exp{
				movie: nil,
				err:   app.ErrNotFound,
			},
		},
		{
			name: "ok",
			get: Get{
				exp: GetExp{
					movie: expMovie,
					err:   nil,
				},
			},
			exp: Exp{
				movie: expMovie,
				err:   nil,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require := require.New(t)

			controller := gomock.NewController(t)
			mockRepo := mock_repo.NewMockRepositoryTx(controller)

			mockRepo.EXPECT().
				MovieGet(ctx, id).
				Return(tc.get.exp.movie, tc.get.exp.err)

			app := app.NewApplication(mockRepo, nil, nil, nil)

			movie, err := app.MovieGet(ctx, id)
			require.Equal(tc.exp.err, err)
			require.Equal(tc.exp.movie, movie)
		})
	}
}

func TestMoviesGetAll(t *testing.T) {
	t.Parallel()

	var (
		ctx = context.Background()

		offset = 0
		limit  = 50

		expMovies            = []*models.Film{{Title: "movie"}}
		expTotal             = 1000
		expMoviesGetAllError = errors.New("MoviesGetAll error")
		expMoviesCountError  = errors.New("MoviesCount error")
	)

	type GetAllExp struct {
		movies []*models.Film
		err    error
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
		movies []*models.Film
		total  int
		err    error
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
			name: "MoviesGetAll error",
			tx: Tx{
				exp: TxExp{
					err: expMoviesGetAllError,
				},
			},
			getAll: GetAll{
				exp: GetAllExp{
					movies: nil,
					err:    expMoviesGetAllError,
				},
			},
			exp: Exp{
				movies: nil,
				total:  0,
				err:    expMoviesGetAllError,
			},
		},

		{
			name: "MoviesCount error",
			tx: Tx{
				exp: TxExp{
					err: expMoviesCountError,
				},
			},
			getAll: GetAll{
				exp: GetAllExp{
					movies: expMovies,
					err:    nil,
				},
			},
			count: Count{
				exp: CountExp{
					total: 0,
					err:   expMoviesCountError,
				},
			},
			exp: Exp{
				movies: nil,
				total:  0,
				err:    expMoviesCountError,
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
					movies: expMovies,
					err:    nil,
				},
			},
			count: Count{
				exp: CountExp{
					total: expTotal,
					err:   nil,
				},
			},
			exp: Exp{
				movies: expMovies,
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
			mockRepo := mock_repo.NewMockRepositoryTx(controller)

			txCall := mockRepo.EXPECT().
				Transaction(ctx, gomock.Any()).
				Do(func(ctx context.Context, fn func(_ context.Context, _ repo.Service) error) {
					fn(ctx, mockRepo)
				}).
				Return(tc.tx.exp.err)

			getAllCall := mockRepo.EXPECT().
				MoviesGetAll(ctx, offset, limit).
				Return(tc.getAll.exp.movies, tc.getAll.exp.err).
				After(txCall)

			if tc.getAll.exp.err == nil {
				mockRepo.EXPECT().
					MoviesCount(ctx).
					Return(tc.count.exp.total, tc.count.exp.err).
					After(getAllCall)
			}

			app := app.NewApplication(mockRepo, nil, nil, nil)

			movies, total, err := app.MoviesGetAll(ctx, offset, limit)
			require.Equal(tc.exp.err, err)
			require.Equal(tc.exp.movies, movies)
			require.Equal(tc.exp.total, total)
		})
	}
}

func TestMovieCreate(t *testing.T) {
	t.Parallel()

	var (
		ctx = context.Background()

		movieID       = 1
		contributorID = 1
		req           = &dto.MovieCreateRequest{
			Title: "movie",
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
		movieID int
		err     error
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
				exp: CreateExp{err: expError},
			},
			exp: Exp{
				movieID: 0,
				err:     expError,
			},
		},

		{
			name: "ok",
			create: Create{
				exp: CreateExp{err: nil},
			},
			exp: Exp{
				movieID: movieID,
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
			mockRepo := mock_repo.NewMockRepositoryTx(controller)

			mockRepo.EXPECT().
				MovieCreate(
					ctx,
					contributorID,
					&models.Film{
						Title:        req.Title,
						Descriptions: req.Descriptions,
						DateReleased: req.DateReleased,
						Duration:     req.Duration,
					},
				).
				Do(func(_ context.Context, _ int, m *models.Film) {
					m.ID = movieID
				}).
				Return(tc.create.exp.err)

			app := app.NewApplication(mockRepo, nil, nil, nil)

			id, err := app.MovieCreate(ctx, contributorID, req)
			require.Equal(tc.exp.err, err)
			require.Equal(tc.exp.movieID, id)
		})
	}
}

//go:linkname movieUpdateRequestToValidMap github.com/aria3ppp/watchlist-server/internal/app.movieUpdateRequestToValidMap
func movieUpdateRequestToValidMap(
	req *dto.MovieUpdateRequest,
) map[string]any

func TestMovieUpdate(t *testing.T) {
	t.Parallel()

	var (
		ctx = context.Background()

		id            = 1
		contributorID = 1
		req           = &dto.MovieUpdateRequest{
			Title:        null.StringFrom("movie"),
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
			mockRepo := mock_repo.NewMockRepositoryTx(controller)

			mockRepo.EXPECT().
				MovieUpdate(ctx, id, contributorID, movieUpdateRequestToValidMap(req)).
				Return(tc.update.exp.err)

			app := app.NewApplication(mockRepo, nil, nil, nil)

			err := app.MovieUpdate(ctx, id, contributorID, req)
			require.Equal(tc.exp.err, err)
		})
	}
}

func TestMovieInvalidate(t *testing.T) {
	t.Parallel()

	var (
		ctx = context.Background()

		id            = 1
		contributorID = 1
		req           = &dto.InvalidationRequest{
			Invalidation: "invalidation",
		}

		expError = errors.New("error")
	)

	type MovieInvalidateExp struct {
		err error
	}
	type MovieInvalidate struct {
		exp MovieInvalidateExp
	}
	type Exp struct {
		err error
	}
	type TestCase struct {
		name            string
		movieInvalidate MovieInvalidate
		exp             Exp
	}

	testCases := []TestCase{
		{
			name: "error",
			movieInvalidate: MovieInvalidate{
				exp: MovieInvalidateExp{
					err: expError,
				},
			},
			exp: Exp{
				err: expError,
			},
		},

		{
			name: "not found",
			movieInvalidate: MovieInvalidate{
				exp: MovieInvalidateExp{
					err: repo.ErrNoRecord,
				},
			},
			exp: Exp{
				err: app.ErrNotFound,
			},
		},

		{
			name: "ok",
			movieInvalidate: MovieInvalidate{
				exp: MovieInvalidateExp{
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
			mockRepo := mock_repo.NewMockRepositoryTx(controller)

			mockRepo.EXPECT().
				MovieInvalidate(ctx, id, contributorID, req.Invalidation).
				Return(tc.movieInvalidate.exp.err)

			app := app.NewApplication(mockRepo, nil, nil, nil)

			err := app.MovieInvalidate(ctx, id, contributorID, req)
			require.Equal(tc.exp.err, err)
		})
	}
}

func TestMovieAuditsGetAll(t *testing.T) {
	t.Parallel()

	var (
		ctx = context.Background()

		id     = 1
		offset = 0
		limit  = 50

		expMovie = &models.Film{
			Title: "movie",
		}
		expMovieGetError          = errors.New("MovieGet error")
		expMovieAuditsGetAllError = errors.New("MovieAuditsGetAll error")
		expMovieAuditsCountError  = errors.New("MovieAuditsCount error")
		expAudits                 = []*models.FilmsAudit{{Title: "audit"}}
		expTotal                  = 100
	)

	type MovieGetExp struct {
		movie *models.Film
		err   error
	}
	type MovieAuditsCountExp struct {
		total int
		err   error
	}
	type MovieAuditsGetAllExp struct {
		audits []*models.FilmsAudit
		err    error
	}
	type TxExp struct {
		err error
	}
	type Tx struct {
		exp TxExp
	}
	type MovieGet struct {
		exp MovieGetExp
	}
	type MovieAuditsGetAll struct {
		exp MovieAuditsGetAllExp
	}
	type MovieAuditsCount struct {
		exp MovieAuditsCountExp
	}
	type Exp struct {
		audits []*models.FilmsAudit
		total  int
		err    error
	}
	type TestCase struct {
		name              string
		tx                Tx
		movieGet          MovieGet
		movieAuditsGetAll MovieAuditsGetAll
		movieAuditsCount  MovieAuditsCount
		exp               Exp
	}

	testCases := []TestCase{
		{
			name: "not found",
			tx: Tx{
				exp: TxExp{
					err: app.ErrNotFound,
				},
			},
			movieGet: MovieGet{
				exp: MovieGetExp{
					movie: nil,
					err:   repo.ErrNoRecord,
				},
			},
			exp: Exp{
				audits: nil,
				total:  0,
				err:    app.ErrNotFound,
			},
		},

		{
			name: "MovieGet error",
			tx: Tx{
				exp: TxExp{
					err: expMovieGetError,
				},
			},
			movieGet: MovieGet{
				exp: MovieGetExp{
					movie: nil,
					err:   expMovieGetError,
				},
			},
			exp: Exp{
				audits: nil,
				total:  0,
				err:    expMovieGetError,
			},
		},

		{
			name: "MovieAuditsGetAll error",
			tx: Tx{
				exp: TxExp{
					err: expMovieAuditsGetAllError,
				},
			},
			movieGet: MovieGet{
				exp: MovieGetExp{
					movie: expMovie,
					err:   nil,
				},
			},
			movieAuditsGetAll: MovieAuditsGetAll{
				exp: MovieAuditsGetAllExp{
					audits: nil,
					err:    expMovieAuditsGetAllError,
				},
			},
			exp: Exp{
				audits: nil,
				total:  0,
				err:    expMovieAuditsGetAllError,
			},
		},

		{
			name: "MovieAuditsCount error",
			tx: Tx{
				exp: TxExp{
					err: expMovieAuditsCountError,
				},
			},
			movieGet: MovieGet{
				exp: MovieGetExp{
					movie: expMovie,
					err:   nil,
				},
			},
			movieAuditsGetAll: MovieAuditsGetAll{
				exp: MovieAuditsGetAllExp{
					audits: expAudits,
					err:    nil,
				},
			},
			movieAuditsCount: MovieAuditsCount{
				exp: MovieAuditsCountExp{
					total: 0,
					err:   expMovieAuditsCountError,
				},
			},
			exp: Exp{
				audits: nil,
				total:  0,
				err:    expMovieAuditsCountError,
			},
		},

		{
			name: "ok",
			tx: Tx{
				exp: TxExp{
					err: nil,
				},
			},
			movieGet: MovieGet{
				exp: MovieGetExp{
					movie: expMovie,
					err:   nil,
				},
			},
			movieAuditsGetAll: MovieAuditsGetAll{
				exp: MovieAuditsGetAllExp{
					audits: expAudits,
					err:    nil,
				},
			},
			movieAuditsCount: MovieAuditsCount{
				exp: MovieAuditsCountExp{
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
			mockRepo := mock_repo.NewMockRepositoryTx(controller)

			txCall := mockRepo.EXPECT().
				Transaction(ctx, gomock.Any()).
				Do(func(ctx context.Context, fn func(_ context.Context, _ repo.Service) error) {
					fn(ctx, mockRepo)
				}).
				Return(tc.tx.exp.err)

			movieGetCall := mockRepo.EXPECT().
				MovieGet(ctx, id).
				Return(tc.movieGet.exp.movie, tc.movieGet.exp.err).
				After(txCall)

			if tc.movieGet.exp.err == nil {
				movieAuditsGetAllCall := mockRepo.EXPECT().
					MovieAuditsGetAll(ctx, id, offset, limit).
					Return(tc.movieAuditsGetAll.exp.audits, tc.movieAuditsGetAll.exp.err).
					After(movieGetCall)

				if tc.movieAuditsGetAll.exp.err == nil {
					mockRepo.EXPECT().
						MovieAuditsCount(ctx, id).
						Return(tc.movieAuditsCount.exp.total, tc.movieAuditsCount.exp.err).
						After(movieAuditsGetAllCall)
				}
			}

			app := app.NewApplication(mockRepo, nil, nil, nil)

			audits, total, err := app.MovieAuditsGetAll(ctx, id, offset, limit)
			require.Equal(tc.exp.err, err)
			require.Equal(tc.exp.audits, audits)
			require.Equal(tc.exp.total, total)
		})
	}
}

func TestMoviesSearch(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var (
		query     = "query"
		offset    = 0
		limit     = 10
		expError  = errors.New("error")
		expMovies = []*models.Film{{Title: "title"}}
		expTotal  = 100
	)

	type SearchExp struct {
		movies []*models.Film
		total  int
		err    error
	}
	type Search struct {
		exp SearchExp
	}
	type Exp struct {
		movies []*models.Film
		total  int
		err    error
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
					movies: nil,
					total:  0,
					err:    expError,
				},
			},
			exp: Exp{
				movies: nil,
				total:  0,
				err:    expError,
			},
		},
		{
			name: "ok",
			search: Search{
				exp: SearchExp{
					movies: expMovies,
					total:  expTotal,
					err:    nil,
				},
			},
			exp: Exp{
				movies: expMovies,
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
			mockSearch := mock_search.NewMockService(controller)

			mockSearch.EXPECT().
				SearchMovies(ctx, query, offset, limit).
				Return(tc.search.exp.movies, tc.exp.total, tc.search.exp.err)

			app := app.NewApplication(nil, nil, mockSearch, nil)

			series, total, err := app.MoviesSearch(ctx, query, offset, limit)
			require.Equal(tc.exp.err, err)
			require.Equal(tc.exp.movies, series)
			require.Equal(tc.exp.total, total)
		})
	}
}
