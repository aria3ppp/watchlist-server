package app

import (
	"context"
	"io"

	"github.com/aria3ppp/watchlist-server/internal/dto"
	"github.com/aria3ppp/watchlist-server/internal/models"
	"github.com/aria3ppp/watchlist-server/internal/query"
	"github.com/aria3ppp/watchlist-server/internal/repo"
	"github.com/aria3ppp/watchlist-server/internal/storage"
)

func (a *Application) MovieGet(
	ctx context.Context,
	id int,
) (*models.Film, error) {
	movie, err := a.repository.MovieGet(ctx, id)
	if err != nil {
		if err == repo.ErrNoRecord {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return movie, nil
}

func (a *Application) MoviesGetAll(
	ctx context.Context,
	queryOptions query.Options,
) (movies []*models.Film, total int, err error) {
	err = a.repository.Transaction(
		ctx,
		func(ctx context.Context, tx repo.Service) error {
			var err error
			movies, err = tx.MoviesGetAll(ctx, queryOptions)
			if err != nil {
				return err
			}
			total, err = tx.MoviesCount(ctx)
			return err
		},
	)
	if err != nil {
		return nil, 0, err
	}
	return movies, total, nil
}

func (a *Application) MovieCreate(
	ctx context.Context,
	contributorID int,
	req *dto.MovieCreateRequest,
) (movieID int, err error) {
	insertMovie := &models.Film{
		Title:        req.Title,
		Descriptions: req.Descriptions,
		DateReleased: req.DateReleased,
		Duration:     req.Duration,
	}

	err = a.repository.MovieCreate(ctx, contributorID, insertMovie)
	if err != nil {
		return 0, err
	}

	return insertMovie.ID, nil
}

func (a *Application) MovieUpdate(
	ctx context.Context,
	id int,
	contributorID int,
	req *dto.MovieUpdateRequest,
) error {
	columns := movieUpdateRequestToValidMap(req)

	err := a.repository.MovieUpdate(ctx, id, contributorID, columns)
	if err != nil {
		if err == repo.ErrNoRecord {
			return ErrNotFound
		}
		return err
	}

	return nil
}

func movieUpdateRequestToValidMap(
	req *dto.MovieUpdateRequest,
) map[string]any {
	m := make(map[string]any)
	if req.Title.Valid {
		m[models.FilmColumns.Title] = req.Title.String
	}
	if req.Descriptions.Valid {
		m[models.FilmColumns.Descriptions] = req.Descriptions.String
	}
	if req.DateReleased.Valid {
		m[models.FilmColumns.DateReleased] = req.DateReleased.Time
	}
	if req.Duration.Valid {
		m[models.FilmColumns.Duration] = req.Duration.Int
	}
	return m
}

func (a *Application) MovieInvalidate(
	ctx context.Context,
	id int,
	contributorID int,
	req *dto.InvalidationRequest,
) error {
	err := a.repository.MovieUpdate(
		ctx,
		id,
		contributorID,
		map[string]any{
			models.FilmColumns.Invalidation: req.Invalidation,
		},
	)
	if err != nil {
		if err == repo.ErrNoRecord {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (a *Application) MovieAuditsGetAll(
	ctx context.Context,
	id int,
	queryOptions query.SortOrderOptions,
) (audits []*models.FilmsAudit, total int, err error) {
	err = a.repository.Transaction(
		ctx,
		func(ctx context.Context, tx repo.Service) error {
			// first check the movie exists
			_, err := tx.MovieGet(ctx, id)
			if err != nil {
				if err == repo.ErrNoRecord {
					return ErrNotFound
				}
				return err
			}
			// fetch audits
			audits, err = tx.MovieAuditsGetAll(ctx, id, queryOptions)
			if err != nil {
				return err
			}
			// count total audits
			total, err = tx.MovieAuditsCount(ctx, id)
			return err
		},
	)
	if err != nil {
		return nil, 0, err
	}
	return audits, total, nil
}

func (a *Application) MoviesSearch(
	ctx context.Context,
	queryOptions query.SearchOptions,
) (results []*models.Film, total int, err error) {
	return a.search.SearchMovies(ctx, queryOptions)
}

func (a *Application) MoviePutPoster(
	ctx context.Context,
	id int,
	contributorID int,
	poster io.Reader,
	options *storage.PutOptions,
) (uri string, err error) {
	// put file
	uri, err = a.storage.PutFile(ctx, poster, options)
	if err != nil {
		return "", err
	}
	// update movie poster
	err = a.repository.MovieUpdate(ctx, id, contributorID, map[string]any{
		models.FilmColumns.Poster: uri,
	})
	if err != nil {
		// TODO: transactional approach is to delete file in storage service on failure
		if err == repo.ErrNoRecord {
			return "", ErrNotFound
		}
		return "", err
	}
	return uri, nil
}
