package app_test

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"testing"

	"github.com/aria3ppp/watchlist-server/internal/app"
	"github.com/aria3ppp/watchlist-server/internal/models"
	"github.com/aria3ppp/watchlist-server/internal/query"
	"github.com/aria3ppp/watchlist-server/internal/repo"
	"github.com/aria3ppp/watchlist-server/internal/repo/mock_repo"
	"github.com/aria3ppp/watchlist-server/internal/watchlist"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestWatchlistGet(t *testing.T) {
	t.Parallel()

	var (
		ctx = context.Background()

		userID       = 1
		queryOptions = query.WatchlistOptions{
			Offset:           0,
			Limit:            math.MaxInt,
			SortOrder:        "asc",
			WhereTimeWatched: repo.RawSqlWhereTimeWatchedEmptyClause,
		}

		expWatchlist = []*watchlist.Item{
			{
				Watchfilm: models.Watchfilm{UserID: userID, FilmID: 1},
				Film:      models.Film{ID: 1, Title: "film"},
			},
		}
		expTotal                = 1000
		expWatchlistGetAllError = errors.New("WatchlistGetAll error")
		expWatchlistCountError  = errors.New("WatchlistCount error")
	)

	type TxExp struct {
		err error
	}
	type Tx struct {
		exp TxExp
	}
	type GetAllExp struct {
		watchlist []*watchlist.Item
		err       error
	}
	type GetAll struct {
		exp GetAllExp
	}
	type CountExp struct {
		total int
		err   error
	}
	type Count struct {
		exp CountExp
	}
	type Exp struct {
		watchlist []*watchlist.Item
		total     int
		err       error
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
			name: "WatchlistGetAll error",
			tx: Tx{
				exp: TxExp{
					err: expWatchlistGetAllError,
				},
			},
			getAll: GetAll{
				exp: GetAllExp{
					watchlist: nil,
					err:       expWatchlistGetAllError,
				},
			},
			exp: Exp{
				watchlist: nil,
				total:     0,
				err:       expWatchlistGetAllError,
			},
		},

		{
			name: "WatchlistCount error",
			tx: Tx{
				exp: TxExp{
					err: expWatchlistCountError,
				},
			},
			getAll: GetAll{
				exp: GetAllExp{
					watchlist: expWatchlist,
					err:       nil,
				},
			},
			count: Count{
				exp: CountExp{
					total: 0,
					err:   expWatchlistCountError,
				},
			},
			exp: Exp{
				watchlist: nil,
				total:     0,
				err:       expWatchlistCountError,
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
					watchlist: expWatchlist,
					err:       nil,
				},
			},
			count: Count{
				exp: CountExp{
					total: expTotal,
					err:   nil,
				},
			},
			exp: Exp{
				watchlist: expWatchlist,
				total:     expTotal,
				err:       nil,
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
				WatchlistGet(ctx, userID, queryOptions).
				Return(tc.getAll.exp.watchlist, tc.getAll.exp.err).
				After(txCall)

			if tc.getAll.exp.err == nil {
				mockRepo.EXPECT().
					WatchlistCount(ctx, userID, queryOptions.WhereTimeWatched).
					Return(tc.count.exp.total, tc.count.exp.err).
					After(getAllCall)
			}

			app := app.NewApplication(mockRepo, nil, nil, nil, nil)

			watchlist, total, err := app.WatchlistGet(
				ctx,
				userID,
				queryOptions,
			)
			require.Equal(tc.exp.err, err)
			require.Equal(tc.exp.watchlist, watchlist)
			require.Equal(tc.exp.total, total)
		})
	}
}

func TestWatchlistAdd(t *testing.T) {
	t.Parallel()

	var (
		ctx = context.Background()

		userID = 1
		filmID = 1

		expWatchID           = 1
		expFilmExistsError   = errors.New("FilmExists error")
		expWatchlistAddError = errors.New("WatchlistAdd error")
	)

	type TxExp struct {
		err error
	}
	type Tx struct {
		exp TxExp
	}
	type FilmExistsExp struct {
		err error
	}
	type FilmExists struct {
		exp FilmExistsExp
	}
	type AddExp struct {
		watchID int
		err     error
	}
	type Add struct {
		exp AddExp
	}
	type Exp struct {
		watchID int
		err     error
	}
	type TestCase struct {
		name       string
		tx         Tx
		filmExists FilmExists
		add        Add
		exp        Exp
	}

	testCases := []TestCase{
		{
			name: "not found",
			tx: Tx{
				exp: TxExp{
					err: app.ErrNotFound,
				},
			},
			filmExists: FilmExists{
				exp: FilmExistsExp{
					err: repo.ErrNoRecord,
				},
			},
			exp: Exp{
				watchID: 0,
				err:     app.ErrNotFound,
			},
		},

		{
			name: "FilmExists error",
			tx: Tx{
				exp: TxExp{
					err: expFilmExistsError,
				},
			},
			filmExists: FilmExists{
				exp: FilmExistsExp{
					err: expFilmExistsError,
				},
			},
			exp: Exp{
				watchID: 0,
				err:     expFilmExistsError,
			},
		},

		{
			name: "WatchlistAdd error",
			tx: Tx{
				exp: TxExp{
					err: expWatchlistAddError,
				},
			},
			filmExists: FilmExists{
				exp: FilmExistsExp{
					err: nil,
				},
			},
			add: Add{
				exp: AddExp{
					watchID: 0,
					err:     expWatchlistAddError,
				},
			},
			exp: Exp{
				watchID: 0,
				err:     expWatchlistAddError,
			},
		},

		{
			name: "ok",
			tx: Tx{
				exp: TxExp{
					err: nil,
				},
			},
			filmExists: FilmExists{
				exp: FilmExistsExp{
					err: nil,
				},
			},
			add: Add{
				exp: AddExp{
					watchID: expWatchID,
					err:     nil,
				},
			},
			exp: Exp{
				watchID: expWatchID,
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

			txCall := mockRepo.EXPECT().
				Tx(ctx, nil, gomock.Any()).
				Do(func(ctx context.Context, opts *sql.TxOptions, fn func(_ context.Context, _ repo.Service) error) {
					fn(ctx, mockRepo)
				}).
				Return(tc.tx.exp.err)

			filmExistsCall := mockRepo.EXPECT().
				FilmExists(ctx, filmID).
				Return(tc.filmExists.exp.err).
				After(txCall)

			if tc.filmExists.exp.err == nil {
				mockRepo.EXPECT().
					WatchlistAdd(ctx, userID, filmID).
					Return(tc.add.exp.watchID, tc.add.exp.err).
					After(filmExistsCall)
			}

			app := app.NewApplication(mockRepo, nil, nil, nil, nil)

			watchID, err := app.WatchlistAdd(ctx, userID, filmID)
			require.Equal(tc.exp.err, err)
			require.Equal(tc.exp.watchID, watchID)
		})
	}
}

func TestWatchlistDelete(t *testing.T) {
	t.Parallel()

	var (
		ctx = context.Background()

		userID  = 1
		watchID = 1

		expWatchlistDeleteError = errors.New("WatchlistDelete error")
	)

	type DeleteExp struct {
		err error
	}
	type Delete struct {
		exp DeleteExp
	}
	type Exp struct {
		err error
	}
	type TestCase struct {
		name   string
		delete Delete
		exp    Exp
	}

	testCases := []TestCase{
		{
			name: "not found",
			delete: Delete{
				exp: DeleteExp{
					err: repo.ErrNoRecord,
				},
			},
			exp: Exp{
				err: app.ErrNotFound,
			},
		},

		{
			name: "WatchlistDelete error",
			delete: Delete{
				exp: DeleteExp{
					err: expWatchlistDeleteError,
				},
			},
			exp: Exp{
				err: expWatchlistDeleteError,
			},
		},

		{
			name: "ok",
			delete: Delete{
				exp: DeleteExp{
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
				WatchlistDelete(ctx, userID, watchID).
				Return(tc.delete.exp.err)

			app := app.NewApplication(mockRepo, nil, nil, nil, nil)

			err := app.WatchlistDelete(ctx, userID, watchID)
			require.Equal(tc.exp.err, err)
		})
	}
}

func TestWatchlistSetWatched(t *testing.T) {
	t.Parallel()

	var (
		ctx = context.Background()

		userID  = 1
		watchID = 1

		expWatchlistSetWatchedError = errors.New("WatchlistSetWatched error")
	)

	type SetWatchedExp struct {
		err error
	}
	type SetWatched struct {
		exp SetWatchedExp
	}
	type Exp struct {
		err error
	}
	type TestCase struct {
		name       string
		setWatched SetWatched
		exp        Exp
	}

	testCases := []TestCase{
		{
			name: "not found",
			setWatched: SetWatched{
				exp: SetWatchedExp{
					err: repo.ErrNoRecord,
				},
			},
			exp: Exp{
				err: app.ErrNotFound,
			},
		},

		{
			name: "WatchlistSetWatched error",
			setWatched: SetWatched{
				exp: SetWatchedExp{
					err: expWatchlistSetWatchedError,
				},
			},
			exp: Exp{
				err: expWatchlistSetWatchedError,
			},
		},

		{
			name: "ok",
			setWatched: SetWatched{
				exp: SetWatchedExp{
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
				WatchlistSetWatched(ctx, userID, watchID).
				Return(tc.setWatched.exp.err)

			app := app.NewApplication(mockRepo, nil, nil, nil, nil)

			err := app.WatchlistSetWatched(ctx, userID, watchID)
			require.Equal(tc.exp.err, err)
		})
	}
}
