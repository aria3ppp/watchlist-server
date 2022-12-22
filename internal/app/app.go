package app

import (
	"context"

	"github.com/aria3ppp/watchlist-server/internal/dto"
	"github.com/aria3ppp/watchlist-server/internal/hasher"
	"github.com/aria3ppp/watchlist-server/internal/models"
	"github.com/aria3ppp/watchlist-server/internal/query"
	"github.com/aria3ppp/watchlist-server/internal/repo"
	"github.com/aria3ppp/watchlist-server/internal/search"
	"github.com/aria3ppp/watchlist-server/internal/token"
	"github.com/aria3ppp/watchlist-server/internal/watchlist"
)

// remove leading comment symbols to enable mocking
////go:generate mockgen -destination mock_app/mock_service.go . Service

type Service interface {
	// User
	UserGet(ctx context.Context, id int) (*models.User, error)
	UserCreate(
		ctx context.Context,
		req *dto.UserCreateRequest,
	) (int, error)
	UserUpdate(
		ctx context.Context,
		id int,
		req *dto.UserUpdateRequest,
	) error
	UserDelete(
		ctx context.Context,
		id int,
		req *dto.UserDeleteRequest,
	) error
	UserEmailUpdate(
		ctx context.Context,
		userID int,
		req *dto.UserEmailUpdateRequest,
	) error
	UserPasswordUpdate(
		ctx context.Context,
		id int,
		req *dto.UserPasswordUpdateRequest,
	) error
	UserLogin(
		ctx context.Context,
		req *dto.UserLoginRequest,
	) (accessToken string, refreshToken string, err error)
	UserRefreshToken(
		ctx context.Context,
		refreshToken string,
	) (accessToken string, err error)

	// Movie
	MovieGet(ctx context.Context, id int) (*models.Film, error)
	MoviesGetAll(
		ctx context.Context,
		queryOptions query.Options,
	) (movies []*models.Film, total int, err error)
	MovieCreate(
		ctx context.Context,
		contributorID int,
		req *dto.MovieCreateRequest,
	) (movieID int, err error)
	MovieUpdate(
		ctx context.Context,
		id int,
		contributorID int,
		req *dto.MovieUpdateRequest,
	) error
	MovieInvalidate(
		ctx context.Context,
		id int,
		contributorID int,
		req *dto.InvalidationRequest,
	) error
	MovieAuditsGetAll(
		ctx context.Context,
		id int,
		queryOptions query.SortOrderOptions,
	) (audits []*models.FilmsAudit, total int, err error)
	MoviesSearch(
		ctx context.Context,
		queryOptions query.SearchOptions,
	) (results []*models.Film, total int, err error)

	// Series
	SeriesGet(ctx context.Context, id int) (*models.Series, error)
	SeriesesGetAll(
		ctx context.Context,
		queryOptions query.Options,
	) (series []*models.Series, total int, err error)
	SeriesCreate(
		ctx context.Context,
		contributorID int,
		req *dto.SeriesCreateRequest,
	) (seriesID int, err error)
	SeriesUpdate(
		ctx context.Context,
		seriesID int,
		contributorID int,
		req *dto.SeriesUpdateRequest,
	) error
	SeriesInvalidate(
		ctx context.Context,
		seriesID int,
		contributorID int,
		req *dto.InvalidationRequest,
	) error
	SeriesAuditsGetAll(
		ctx context.Context,
		id int,
		queryOptions query.SortOrderOptions,
	) (audits []*models.SeriesesAudit, total int, err error)
	SeriesesSearch(
		ctx context.Context,
		queryOptions query.SearchOptions,
	) (results []*models.Series, total int, err error)

	// Episode
	EpisodeGet(
		ctx context.Context,
		seriesID, seasonNumber, episodeNumber int,
	) (*models.Film, error)
	EpisodesGetAllBySeries(
		ctx context.Context,
		seriesID int,
		queryOptions query.SortOrderOptions,
	) (episodes []*models.Film, total int, err error)
	EpisodesGetAllBySeason(
		ctx context.Context,
		seriesID int,
		seasonNumber int,
		queryOptions query.SortOrderOptions,
	) (episodes []*models.Film, total int, err error)
	EpisodePut(
		ctx context.Context,
		seriesID, seasonNumber, episodeNumber int,
		contributorID int,
		req *dto.EpisodePutRequest,
	) error
	EpisodesPutAllBySeason(
		ctx context.Context,
		seriesID int,
		seasonNumber int,
		contributorID int,
		req *dto.EpisodesPutAllBySeasonRequest,
	) error
	EpisodeUpdate(
		ctx context.Context,
		seriesID, seasonNumber, episodeNumber int,
		contributorID int,
		req *dto.EpisodeUpdateRequest,
	) error
	EpisodeInvalidate(
		ctx context.Context,
		seriesID, seasonNumber, episodeNumber int,
		contributorID int,
		req *dto.InvalidationRequest,
	) error
	EpisodesInvalidateAllBySeason(
		ctx context.Context,
		seriesID, seasonNumber,
		contributorID int,
		req *dto.InvalidationRequest,
	) error
	EpisodeAuditsGetAll(
		ctx context.Context,
		seriesID, seasonNumber, episodeNumber int,
		queryOptions query.SortOrderOptions,
	) (audits []*models.FilmsAudit, total int, err error)

	// Watchlist
	WatchlistGet(
		ctx context.Context,
		userID int,
		queryOptions query.WatchlistOptions,
	) (watchlist []*watchlist.Item, total int, err error)
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

type Application struct {
	repository repo.ServiceTx
	token      token.Service
	search     search.Service
	hasher     hasher.Interface
}

var _ Service = (*Application)(nil)

func NewApplication(
	repo repo.ServiceTx,
	tokenService token.Service,
	searchService search.Service,
	hasher hasher.Interface,
) *Application {
	return &Application{
		repository: repo,
		token:      tokenService,
		search:     searchService,
		hasher:     hasher,
	}
}
