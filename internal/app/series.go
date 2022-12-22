package app

import (
	"context"

	"github.com/aria3ppp/watchlist-server/internal/dto"
	"github.com/aria3ppp/watchlist-server/internal/models"
	"github.com/aria3ppp/watchlist-server/internal/query"
	"github.com/aria3ppp/watchlist-server/internal/repo"
)

func (a *Application) SeriesGet(
	ctx context.Context,
	id int,
) (*models.Series, error) {
	series, err := a.repository.SeriesGet(ctx, id)
	if err != nil {
		if err == repo.ErrNoRecord {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return series, nil
}

func (a *Application) SeriesesGetAll(
	ctx context.Context,
	queryOptions query.Options,
) (series []*models.Series, total int, err error) {
	err = a.repository.Transaction(
		ctx,
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

func (a *Application) SeriesCreate(
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

	err = a.repository.SeriesCreate(ctx, contributorID, insertSeries)
	if err != nil {
		return 0, err
	}

	return insertSeries.ID, nil
}

func (a *Application) SeriesUpdate(
	ctx context.Context,
	seriesID int,
	contributorID int,
	req *dto.SeriesUpdateRequest,
) error {
	columns := seriesUpdateRequestToValidMap(req)

	err := a.repository.SeriesUpdate(ctx, seriesID, contributorID, columns)
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

func (a *Application) SeriesInvalidate(
	ctx context.Context,
	seriesID int,
	contributorID int,
	req *dto.InvalidationRequest,
) error {
	err := a.repository.Transaction(
		ctx,
		func(ctx context.Context, tx repo.Service) error {
			// first invalidate series itself
			err := tx.SeriesInvalidate(
				ctx,
				seriesID,
				contributorID,
				req.Invalidation,
			)
			if err != nil {
				if err == repo.ErrNoRecord {
					return ErrNotFound
				}
				return err
			}
			// then invalidate all the related episodes if any
			err = tx.EpisodesInvalidateAllBySeries(
				ctx,
				seriesID,
				contributorID,
				req.Invalidation,
			)
			if err != nil && err != repo.ErrNoRecord {
				return err
			}
			return nil
		},
	)
	return err
}

func (a *Application) SeriesAuditsGetAll(
	ctx context.Context,
	id int,
	queryOptions query.SortOrderOptions,
) (audits []*models.SeriesesAudit, total int, err error) {
	err = a.repository.Transaction(
		ctx,
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

func (a *Application) SeriesesSearch(
	ctx context.Context,
	queryOptions query.SearchOptions,
) (results []*models.Series, total int, err error) {
	return a.search.SearchSerieses(ctx, queryOptions)
}
