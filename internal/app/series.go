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

func (app *Application) SeriesGet(
	ctx context.Context,
	id int,
) (*models.Series, error) {
	series, err := app.repo.SeriesGet(ctx, id)
	if err != nil {
		if err == repo.ErrNoRecord {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return series, nil
}

func (app *Application) SeriesesGetAll(
	ctx context.Context,
	queryOptions query.Options,
) (series []*models.Series, total int, err error) {
	err = app.repo.Tx(
		ctx,
		nil,
		func(ctx context.Context, tx repo.Service) error {
			var err error
			series, err = tx.SeriesesGetAll(ctx, queryOptions)
			if err != nil {
				return err
			}
			total, err = tx.SeriesesCount(ctx)
			return err
		},
	)
	if err != nil {
		return nil, 0, err
	}
	return series, total, nil
}

func (app *Application) SeriesCreate(
	ctx context.Context,
	contributorID int,
	req *dto.SeriesCreateRequest,
) (seriesID int, err error) {
	insertSeries := &models.Series{
		Title:        req.Title,
		Descriptions: req.Descriptions,
		DateStarted:  req.DateStarted,
		DateEnded:    req.DateEnded,
	}

	err = app.repo.SeriesCreate(ctx, contributorID, insertSeries)
	if err != nil {
		return 0, err
	}

	return insertSeries.ID, nil
}

func (app *Application) SeriesUpdate(
	ctx context.Context,
	seriesID int,
	contributorID int,
	req *dto.SeriesUpdateRequest,
) error {
	columns := seriesUpdateRequestToValidMap(req)

	err := app.repo.SeriesUpdate(ctx, seriesID, contributorID, columns)
	if err != nil {
		if err == repo.ErrNoRecord {
			return ErrNotFound
		}
		return err
	}

	return nil
}

func seriesUpdateRequestToValidMap(
	req *dto.SeriesUpdateRequest,
) map[string]any {
	m := make(map[string]any)
	if req.Title.Valid {
		m[models.SeriesColumns.Title] = req.Title.String
	}
	if req.Descriptions.Valid {
		m[models.SeriesColumns.Descriptions] = req.Descriptions.String
	}
	if req.DateStarted.Valid {
		m[models.SeriesColumns.DateStarted] = req.DateStarted.Time
	}
	if req.DateEnded.Valid {
		m[models.SeriesColumns.DateEnded] = req.DateEnded.Time
	}
	return m
}

func (app *Application) SeriesInvalidate(
	ctx context.Context,
	seriesID int,
	contributorID int,
	req *dto.InvalidationRequest,
) error {
	err := app.repo.SeriesUpdate(
		ctx,
		seriesID,
		contributorID,
		map[string]any{
			models.SeriesColumns.Invalidation: req.Invalidation,
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

func (app *Application) SeriesAuditsGetAll(
	ctx context.Context,
	id int,
	queryOptions query.SortOrderOptions,
) (audits []*models.SeriesesAudit, total int, err error) {
	err = app.repo.Tx(
		ctx,
		nil,
		func(ctx context.Context, tx repo.Service) error {
			// first check the series exists
			_, err := tx.SeriesGet(ctx, id)
			if err != nil {
				if err == repo.ErrNoRecord {
					return ErrNotFound
				}
				return err
			}
			// fetch audits
			audits, err = tx.SeriesAuditsGetAll(ctx, id, queryOptions)
			if err != nil {
				return err
			}
			// count total audits
			total, err = tx.SeriesAuditsCount(ctx, id)
			return err
		},
	)
	if err != nil {
		return nil, 0, err
	}
	return audits, total, nil
}

func (app *Application) SeriesesSearch(
	ctx context.Context,
	queryOptions query.SearchOptions,
) (results []*models.Series, total int, err error) {
	return app.search.SearchSerieses(ctx, queryOptions)
}

func (app *Application) SeriesPutPoster(
	ctx context.Context,
	id int,
	contributorID int,
	poster io.Reader,
	options *storage.PutOptions,
) (uri string, err error) {
	// put file
	uri, err = app.storage.PutFile(ctx, poster, options)
	if err != nil {
		return "", err
	}
	// update series poster
	err = app.repo.SeriesUpdate(ctx, id, contributorID, map[string]any{
		models.SeriesColumns.Poster: uri,
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
