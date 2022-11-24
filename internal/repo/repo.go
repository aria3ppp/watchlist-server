package repo

import (
	"context"
	"database/sql"

	"github.com/aria3ppp/watchlist-server/internal/models"
)

//go:generate mockgen -destination mock_repo/mock_service.go . ServiceTx

type ServiceTx interface {
	Transaction(
		context.Context,
		func(context.Context, Service) error,
	) error
	Service
}

type Service interface {
	// User
	UserGet(ctx context.Context, id int) (*models.User, error)
	UserGetByEmail(ctx context.Context, email string) (*models.User, error)
	UsersCount(ctx context.Context) (int, error)
	UserCreate(ctx context.Context, user *models.User) error
	UserUpdate(ctx context.Context, id int, columns map[string]any) error
	UserDelete(ctx context.Context, id int) error

	// Series
	SeriesGet(ctx context.Context, id int) (*models.Series, error)
	SeriesesGetAll(
		ctx context.Context,
		offset, limit int,
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
	SeriesInvalidate(
		ctx context.Context,
		seriesID int,
		contributorID int,
		invalidation string,
	) error
	SeriesAuditsGetAll(
		ctx context.Context,
		id int,
		offset, limit int,
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
		offset, limit int,
	) ([]*models.Film, error)
	EpisodesGetAllBySeason(
		ctx context.Context,
		seriesID int,
		seasonNumber int,
		offset, limit int,
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
	EpisodeInvalidate(
		ctx context.Context,
		seriesID, seasonNumber, episodeNumber int,
		contributorID int,
		invalidation string,
	) error
	EpisodesInvalidateAllBySeason(
		ctx context.Context,
		seriesID int,
		seasonNumber int,
		contributorID int,
		invalidation string,
	) error
	EpisodesInvalidateAllBySeries(
		ctx context.Context,
		seriesID int,
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
		offset, limit int,
	) ([]*models.FilmsAudit, error)
	EpisodeAuditsCount(
		ctx context.Context,
		seriesID, seasonNumber, episodeNumber int,
	) (int, error)
	EpisodesAuditsGetAllBySeason(
		ctx context.Context,
		seriesID int,
		seasonNumber int,
		offset, limit int,
	) ([]*models.FilmsAudit, error)
	EpisodesAuditsCountBySeason(
		ctx context.Context,
		seriesID int,
		seasonNumber int,
	) (int, error)
	EpisodesAuditsGetAllBySeries(
		ctx context.Context,
		seriesID int,
		offset, limit int,
	) ([]*models.FilmsAudit, error)
	EpisodesAuditsCountBySeries(
		ctx context.Context,
		seriesID int,
	) (int, error)

	// Movie
	MovieGet(
		ctx context.Context,
		id int,
	) (*models.Film, error)
	MoviesGetAll(
		ctx context.Context,
		offset, limit int,
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
	MovieInvalidate(
		ctx context.Context,
		movieID int,
		contributorID int,
		invalidation string,
	) error
	MovieAuditsGetAll(
		ctx context.Context,
		id int,
		offset, limit int,
	) ([]*models.FilmsAudit, error)
	MovieAuditsCount(
		ctx context.Context,
		id int,
	) (int, error)
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
