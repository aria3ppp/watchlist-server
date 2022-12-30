package app

import (
	"context"

	"github.com/aria3ppp/watchlist-server/internal/dto"
	"github.com/aria3ppp/watchlist-server/internal/models"
	"github.com/aria3ppp/watchlist-server/internal/query"
	"github.com/aria3ppp/watchlist-server/internal/repo"
)

func (a *Application) EpisodeGet(
	ctx context.Context,
	seriesID, seasonNumber, episodeNumber int,
) (*models.Film, error) {
	episode, err := a.repository.EpisodeGet(
		ctx,
		seriesID,
		seasonNumber,
		episodeNumber,
	)
	if err != nil {
		if err == repo.ErrNoRecord {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return episode, nil
}

func (a *Application) EpisodesGetAllBySeries(
	ctx context.Context,
	seriesID int,
	queryOptions query.SortOrderOptions,
) (episodes []*models.Film, total int, err error) {
	err = a.repository.Transaction(
		ctx,
		func(ctx context.Context, tx repo.Service) error {
			_, err := tx.SeriesGet(ctx, seriesID)
			if err != nil {
				if err == repo.ErrNoRecord {
					return ErrNotFound
				}
				return err
			}
			episodes, err = tx.EpisodesGetAllBySeries(
				ctx,
				seriesID,
				queryOptions,
			)
			if err != nil {
				return err
			}
			total, err = tx.EpisodesCountBySeries(ctx, seriesID)
			return err
		},
	)
	if err != nil {
		return nil, 0, err
	}
	return episodes, total, nil
}

func (a *Application) EpisodesGetAllBySeason(
	ctx context.Context,
	seriesID int,
	seasonNumber int,
	queryOptions query.SortOrderOptions,
) (episodes []*models.Film, total int, err error) {
	err = a.repository.Transaction(
		ctx,
		func(ctx context.Context, tx repo.Service) error {
			_, err := tx.SeriesGet(ctx, seriesID)
			if err != nil {
				if err == repo.ErrNoRecord {
					return ErrNotFound
				}
				return err
			}
			episodes, err = tx.EpisodesGetAllBySeason(
				ctx,
				seriesID,
				seasonNumber,
				queryOptions,
			)
			if err != nil {
				return err
			}
			total, err = tx.EpisodesCountBySeason(ctx, seriesID, seasonNumber)
			return err
		},
	)
	if err != nil {
		return nil, 0, err
	}
	return episodes, total, nil
}

//------------------------------------------------------------------------------

func (a *Application) EpisodePut(
	ctx context.Context,
	seriesID, seasonNumber, episodeNumber int,
	contributorID int,
	req *dto.EpisodePutRequest,
) error {
	err := a.repository.Transaction(
		ctx,
		func(ctx context.Context, tx repo.Service) error {
			// first check seris exists
			_, err := tx.SeriesGet(ctx, seriesID)
			if err != nil {
				if err == repo.ErrNoRecord {
					return ErrNotFound
				}
				return err
			}
			// then put episode
			err = a.repository.EpisodePut(
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
			)
			return err
		},
	)
	return err
}

func (a *Application) EpisodesPutAllBySeason(
	ctx context.Context,
	seriesID int,
	seasonNumber int,
	contributorID int,
	req *dto.EpisodesPutAllBySeasonRequest,
) error {
	err := a.repository.Transaction(
		ctx,
		func(ctx context.Context, tx repo.Service) error {
			// check series id exists
			if _, err := tx.SeriesGet(ctx, seriesID); err != nil {
				if err == repo.ErrNoRecord {
					return ErrNotFound
				}
				return err
			}
			// replace episodes
			for i, e := range req.Episodes {
				episodeNumber := i + 1
				err := tx.EpisodePut(
					ctx,
					seriesID,
					seasonNumber,
					episodeNumber,
					contributorID,
					&models.Film{
						Title:        e.Title,
						Descriptions: e.Descriptions,
						DateReleased: e.DateReleased,
						Duration:     e.Duration,
					},
				)
				if err != nil {
					return err
				}
			}

			return nil
		},
	)
	return err
}

//------------------------------------------------------------------------------

func (a *Application) EpisodeUpdate(
	ctx context.Context,
	seriesID, seasonNumber, episodeNumber int,
	contributorID int,
	req *dto.EpisodeUpdateRequest,
) error {
	columns := episodeUpdateRequestToValidMap(req)

	err := a.repository.EpisodeUpdate(
		ctx,
		seriesID,
		seasonNumber,
		episodeNumber,
		contributorID,
		columns,
	)
	if err != nil {
		if err == repo.ErrNoRecord {
			return ErrNotFound
		}
		return err
	}

	return nil
}

func episodeUpdateRequestToValidMap(
	req *dto.EpisodeUpdateRequest,
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

func (a *Application) EpisodeInvalidate(
	ctx context.Context,
	seriesID, seasonNumber, episodeNumber int,
	contributorID int,
	req *dto.InvalidationRequest,
) error {
	err := a.repository.EpisodeUpdate(
		ctx,
		seriesID,
		seasonNumber,
		episodeNumber,
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

func (a *Application) EpisodesInvalidateAllBySeason(
	ctx context.Context,
	seriesID, seasonNumber,
	contributorID int,
	req *dto.InvalidationRequest,
) error {
	err := a.repository.EpisodesInvalidateAllBySeason(
		ctx,
		seriesID,
		seasonNumber,
		contributorID,
		req.Invalidation,
	)
	if err != nil {
		if err == repo.ErrNoRecord {
			return ErrNotFound
		}
		return err
	}
	return nil
}

//------------------------------------------------------------------------------

func (a *Application) EpisodeAuditsGetAll(
	ctx context.Context,
	seriesID, seasonNumber, episodeNumber int,
	queryOptions query.SortOrderOptions,
) (audits []*models.FilmsAudit, total int, err error) {
	err = a.repository.Transaction(
		ctx,
		func(ctx context.Context, tx repo.Service) error {
			// first check the episode exists
			_, err := tx.EpisodeGet(ctx, seriesID, seasonNumber, episodeNumber)
			if err != nil {
				if err == repo.ErrNoRecord {
					return ErrNotFound
				}
				return err
			}
			// fetch audits
			audits, err = tx.EpisodeAuditsGetAll(
				ctx,
				seriesID,
				seasonNumber,
				episodeNumber,
				queryOptions,
			)
			if err != nil {
				return err
			}
			// count total audits
			total, err = tx.EpisodeAuditsCount(
				ctx,
				seriesID,
				seasonNumber,
				episodeNumber,
			)
			return err
		},
	)
	if err != nil {
		return nil, 0, err
	}
	return audits, total, nil
}
