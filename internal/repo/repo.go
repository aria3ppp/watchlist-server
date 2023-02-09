package repo

import (
	"context"
	"database/sql"

	"github.com/aria3ppp/watchlist-server/internal/models"
	"github.com/aria3ppp/watchlist-server/internal/query"
	"github.com/aria3ppp/watchlist-server/internal/watchlist"
)

//go:generate mockgen -destination mock_repo/mock_service.go . ServiceTx

type ServiceTx interface {
	Service
	Tx(
		context.Context,
		*sql.TxOptions,
		func(context.Context, Service) error,
	) error
}

type Service interface {
	// User
	UserGet(ctx context.Context, id int) (*models.User, error)
	UserGetByEmail(ctx context.Context, email string) (*models.User, error)
	UsersCount(ctx context.Context) (int, error)
	UserCreate(ctx context.Context, user *models.User) error
	UserUpdate(ctx context.Context, id int, columns map[string]any) error
	UserDelete(ctx context.Context, id int) error

	// Token
	TokenGet(
		ctx context.Context,
		userID int,
		refreshToken string,
	) (token *models.Token, err error)
	TokenCreate(
		ctx context.Context,
		token *models.Token,
	) error
	TokenUpdate(
		ctx context.Context,
		tokenID int,
		cols map[string]any,
	) error

	// Series
	SeriesGet(ctx context.Context, id int) (*models.Series, error)
	SeriesesGetAll(
		ctx context.Context,
		queryOptions query.Options,
	) ([]*models.Series, error)
	SeriesesCount(ctx context.Context) (int, error)
	SeriesCreate(
		ctx context.Context,
		contributorID int,
		series *models.Series,
	) error
	SeriesUpdate(
		ctx context.Context,
		seriesID int,
		contributorID int,
		cols map[string]any,
	) error
	SeriesAuditsGetAll(
		ctx context.Context,
		id int,
		queryOptions query.SortOrderOptions,
	) ([]*models.SeriesesAudit, error)
	SeriesAuditsCount(
		ctx context.Context,
		id int,
	) (int, error)

	// Episode
	// EpisodeGetByID(
	// 	ctx context.Context,
	// 	id int,
	// ) (*models.Film, error)
	EpisodeGet(
		ctx context.Context,
		seriesID, seasonNumber, episodeNumber int,
	) (*models.Film, error)
	EpisodesGetAllBySeries(
		ctx context.Context,
		seriesID int,
		queryOptions query.SortOrderOptions,
	) ([]*models.Film, error)
	EpisodesGetAllBySeason(
		ctx context.Context,
		seriesID int,
		seasonNumber int,
		queryOptions query.SortOrderOptions,
	) ([]*models.Film, error)
	EpisodesCountBySeries(
		ctx context.Context,
		seriesID int,
	) (int, error)
	EpisodesCountBySeason(
		ctx context.Context,
		seriesID int,
		seasonNumber int,
	) (int, error)
	EpisodePut(
		ctx context.Context,
		seriesID, seasonNumber, episodeNumber int,
		contributorID int,
		episode *models.Film,
	) error
	EpisodeUpdate(
		ctx context.Context,
		seriesID, seasonNumber, episodeNumber int,
		contributorID int,
		cols map[string]any,
	) error
	EpisodesInvalidateAllBySeason(
		ctx context.Context,
		seriesID int,
		seasonNumber int,
		contributorID int,
		invalidation string,
	) error
	// EpisodeAuditsGetAllByID(
	// 	ctx context.Context,
	// 	id int,
	// 	offset, limit int,
	// ) ([]*models.FilmsAudit, error)
	EpisodeAuditsGetAll(
		ctx context.Context,
		seriesID, seasonNumber, episodeNumber int,
		queryOptions query.SortOrderOptions,
	) ([]*models.FilmsAudit, error)
	EpisodeAuditsCount(
		ctx context.Context,
		seriesID, seasonNumber, episodeNumber int,
	) (int, error)

	// Movie
	MovieGet(
		ctx context.Context,
		id int,
	) (*models.Film, error)
	MoviesGetAll(
		ctx context.Context,
		queryOptions query.Options,
	) ([]*models.Film, error)
	MoviesCount(ctx context.Context) (int, error)
	MovieCreate(
		ctx context.Context,
		contributorID int,
		movie *models.Film,
	) error
	MovieUpdate(
		ctx context.Context,
		movieID int,
		contributorID int,
		cols map[string]any,
	) error
	MovieAuditsGetAll(
		ctx context.Context,
		id int,
		queryOptions query.SortOrderOptions,
	) ([]*models.FilmsAudit, error)
	MovieAuditsCount(
		ctx context.Context,
		id int,
	) (int, error)

	// Film
	FilmExists(ctx context.Context, filmID int) error

	// Watchlist
	WatchlistGet(
		ctx context.Context,
		userID int,
		queryOptions query.WatchlistOptions,
	) (watchlist []*watchlist.Item, err error)
	WatchlistCount(
		ctx context.Context,
		userID int,
		timeWatchedWhereClause string,
	) (count int, err error)
	WatchlistAdd(
		ctx context.Context,
		userID int,
		filmID int,
	) (watchID int, err error)
	WatchlistDelete(
		ctx context.Context,
		userID int,
		watchID int,
	) error
	WatchlistSetWatched(
		ctx context.Context,
		userID int,
		watchID int,
	) error
}

type Repository struct {
	exec Executor
}

var _ ServiceTx = (*Repository)(nil)

type Executor interface {
	Exec(string, ...any) (sql.Result, error)
	Query(string, ...any) (*sql.Rows, error)
	QueryRow(string, ...any) *sql.Row
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
	BeginTx(context.Context, *sql.TxOptions) (*sql.Tx, error)
}

func NewRepository(exec Executor) *Repository {
	return &Repository{exec}
}
