package repo

import (
	"context"
	"database/sql"

	"github.com/aria3ppp/watchlist-server/internal/models"
	"github.com/aria3ppp/watchlist-server/internal/query"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// func (repo *Repo) EpisodeGetByID(
// 	ctx context.Context,
// 	id int,
// ) (*models.Film, error) {
// 	episode, err := models.Films(
// 		models.FilmWhere.ID.EQ(id),
// 		models.FilmWhere.SeriesID.IsNotNull(),
// 		models.FilmWhere.SeasonNumber.IsNotNull(),
// 		models.FilmWhere.EpisodeNumber.IsNotNull(),
// 	).One(ctx, repo.exec)
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			return nil, ErrNoRecord
// 		}
// 		return nil, err
// 	}
// 	return episode, nil
// }

func (repo *Repository) EpisodeGet(
	ctx context.Context,
	seriesID, seasonNumber, episodeNumber int,
) (*models.Film, error) {
	episode, err := models.Films(
		models.FilmWhere.SeriesID.EQ(null.IntFrom(seriesID)),
		models.FilmWhere.SeasonNumber.EQ(null.IntFrom(seasonNumber)),
		models.FilmWhere.EpisodeNumber.EQ(null.IntFrom(episodeNumber)),
	).One(ctx, repo.exec)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoRecord
		}
		return nil, err
	}
	return episode, nil
}

func (repo *Repository) EpisodesGetAllBySeries(
	ctx context.Context,
	seriesID int,
	queryOptions query.SortOrderOptions,
) ([]*models.Film, error) {
	episodes, err := models.Films(
		models.FilmWhere.SeriesID.EQ(null.IntFrom(seriesID)),
		models.FilmWhere.SeasonNumber.IsNotNull(),
		models.FilmWhere.EpisodeNumber.IsNotNull(),
		qm.Offset(queryOptions.Offset),
		qm.Limit(queryOptions.Limit),
		qm.OrderBy(models.FilmColumns.SeasonNumber+" "+queryOptions.SortOrder),
		qm.OrderBy(models.FilmColumns.EpisodeNumber),
	).All(ctx, repo.exec)
	if err != nil {
		return nil, err
	}
	return episodes, nil
}

func (repo *Repository) EpisodesGetAllBySeason(
	ctx context.Context,
	seriesID int,
	seasonNumber int,
	queryOptions query.SortOrderOptions,
) ([]*models.Film, error) {
	episodes, err := models.Films(
		models.FilmWhere.SeriesID.EQ(null.IntFrom(seriesID)),
		models.FilmWhere.SeasonNumber.EQ(null.IntFrom(seasonNumber)),
		models.FilmWhere.EpisodeNumber.IsNotNull(),
		qm.Offset(queryOptions.Offset),
		qm.Limit(queryOptions.Limit),
		qm.OrderBy(models.FilmColumns.EpisodeNumber+" "+queryOptions.SortOrder),
	).All(ctx, repo.exec)
	if err != nil {
		return nil, err
	}
	return episodes, nil
}

////////////////////////////////////////////////////////////////////////////////

func (repo *Repository) EpisodesCountBySeries(
	ctx context.Context,
	seriesID int,
) (int, error) {
	nEpisodes, err := models.Films(
		models.FilmWhere.SeriesID.EQ(null.IntFrom(seriesID)),
		models.FilmWhere.SeasonNumber.IsNotNull(),
		models.FilmWhere.EpisodeNumber.IsNotNull(),
	).Count(ctx, repo.exec)
	return int(nEpisodes), err
}

func (repo *Repository) EpisodesCountBySeason(
	ctx context.Context,
	seriesID int,
	seasonNumber int,
) (int, error) {
	nEpisodes, err := models.Films(
		models.FilmWhere.SeriesID.EQ(null.IntFrom(seriesID)),
		models.FilmWhere.SeasonNumber.EQ(null.IntFrom(seasonNumber)),
		models.FilmWhere.EpisodeNumber.IsNotNull(),
	).Count(ctx, repo.exec)
	return int(nEpisodes), err
}

////////////////////////////////////////////////////////////////////////////////

func (repo *Repository) EpisodePut(
	ctx context.Context,
	seriesID, seasonNumber, episodeNumber int,
	contributorID int,
	episode *models.Film,
) error {
	episode.SeriesID = null.IntFrom(seriesID)
	episode.SeasonNumber = null.IntFrom(seasonNumber)
	episode.EpisodeNumber = null.IntFrom(episodeNumber)
	episode.ContributedBy = contributorID
	return episode.Upsert(
		ctx,
		repo.exec,
		true, // update on conflict
		[]string{
			// upsert on conflict on unique episode
			models.FilmColumns.SeriesID,
			models.FilmColumns.SeasonNumber,
			models.FilmColumns.EpisodeNumber,
		},
		boil.Infer(),
		boil.Infer(),
	)
}

func (repo *Repository) EpisodeUpdate(
	ctx context.Context,
	seriesID, seasonNumber, episodeNumber int,
	contributorID int,
	cols map[string]any,
) error {
	cols[models.FilmColumns.ContributedBy] = contributorID
	rowsAff, err := models.Films(
		models.FilmWhere.SeriesID.EQ(null.IntFrom(seriesID)),
		models.FilmWhere.SeasonNumber.EQ(null.IntFrom(seasonNumber)),
		models.FilmWhere.EpisodeNumber.EQ(null.IntFrom(episodeNumber)),
	).UpdateAll(ctx, repo.exec, cols)
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return ErrNoRecord
	}
	return nil
}

func (repo *Repository) EpisodesInvalidateAllBySeason(
	ctx context.Context,
	seriesID int,
	seasonNumber int,
	contributorID int,
	invalidation string,
) error {
	rowsAff, err := models.Films(
		models.FilmWhere.SeriesID.EQ(null.IntFrom(seriesID)),
		models.FilmWhere.SeasonNumber.EQ(null.IntFrom(seasonNumber)),
		models.FilmWhere.EpisodeNumber.IsNotNull(),
	).UpdateAll(
		ctx,
		repo.exec,
		map[string]any{
			models.FilmColumns.Invalidation:  invalidation,
			models.FilmColumns.ContributedBy: contributorID,
		},
	)
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return ErrNoRecord
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

// func (repo *Repo) EpisodeAuditsGetAllByID(
// 	ctx context.Context,
// 	id int,
// 	offset, limit int,
// ) ([]*models.FilmsAudit, error) {
// 	audits, err := models.FilmsAudits(
// 		models.FilmsAuditWhere.ID.EQ(id),
// 		models.FilmsAuditWhere.SeriesID.IsNotNull(),
// 		models.FilmsAuditWhere.SeasonNumber.IsNotNull(),
// 		models.FilmsAuditWhere.EpisodeNumber.IsNotNull(),
// 		qm.Offset(offset),
// 		qm.Limit(limit),
// 		qm.OrderBy(models.FilmsAuditColumns.ContributedAt+" desc"),
// 	).All(ctx, repo.exec)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return audits, nil
// }

func (repo *Repository) EpisodeAuditsGetAll(
	ctx context.Context,
	seriesID, seasonNumber, episodeNumber int,
	queryOptions query.SortOrderOptions,
) ([]*models.FilmsAudit, error) {
	audits, err := models.FilmsAudits(
		models.FilmsAuditWhere.SeriesID.EQ(null.IntFrom(seriesID)),
		models.FilmsAuditWhere.SeasonNumber.EQ(null.IntFrom(seasonNumber)),
		models.FilmsAuditWhere.EpisodeNumber.EQ(null.IntFrom(episodeNumber)),
		qm.Offset(queryOptions.Offset),
		qm.Limit(queryOptions.Limit),
		qm.OrderBy(
			models.FilmsAuditColumns.ContributedAt+" "+queryOptions.SortOrder,
		),
	).All(ctx, repo.exec)
	if err != nil {
		return nil, err
	}
	return audits, nil
}

func (repo *Repository) EpisodeAuditsCount(
	ctx context.Context,
	seriesID, seasonNumber, episodeNumber int,
) (int, error) {
	auditsCount, err := models.FilmsAudits(
		models.FilmsAuditWhere.SeriesID.EQ(null.IntFrom(seriesID)),
		models.FilmsAuditWhere.SeasonNumber.EQ(null.IntFrom(seasonNumber)),
		models.FilmsAuditWhere.EpisodeNumber.EQ(null.IntFrom(episodeNumber)),
	).Count(ctx, repo.exec)
	if err != nil {
		return 0, err
	}
	return int(auditsCount), nil
}

////////////////////////////////////////////////////////////////////////////////

func (repo *Repository) EpisodesAuditsGetAllBySeason(
	ctx context.Context,
	seriesID int,
	seasonNumber int,
	offset, limit int,
) ([]*models.FilmsAudit, error) {
	audits, err := models.FilmsAudits(
		models.FilmsAuditWhere.SeriesID.EQ(null.IntFrom(seriesID)),
		models.FilmsAuditWhere.SeasonNumber.EQ(null.IntFrom(seasonNumber)),
		models.FilmsAuditWhere.EpisodeNumber.IsNotNull(),
		qm.Offset(offset),
		qm.Limit(limit),
		qm.OrderBy(models.FilmsAuditColumns.ContributedAt+" desc"),
	).All(ctx, repo.exec)
	if err != nil {
		return nil, err
	}
	return audits, nil
}

func (repo *Repository) EpisodesAuditsCountBySeason(
	ctx context.Context,
	seriesID int,
	seasonNumber int,
) (int, error) {
	auditsCount, err := models.FilmsAudits(
		models.FilmsAuditWhere.SeriesID.EQ(null.IntFrom(seriesID)),
		models.FilmsAuditWhere.SeasonNumber.EQ(null.IntFrom(seasonNumber)),
		models.FilmsAuditWhere.EpisodeNumber.IsNotNull(),
	).Count(ctx, repo.exec)
	if err != nil {
		return 0, err
	}
	return int(auditsCount), nil
}

func (repo *Repository) EpisodesAuditsGetAllBySeries(
	ctx context.Context,
	seriesID int,
	offset, limit int,
) ([]*models.FilmsAudit, error) {
	audits, err := models.FilmsAudits(
		models.FilmsAuditWhere.SeriesID.EQ(null.IntFrom(seriesID)),
		models.FilmsAuditWhere.SeasonNumber.IsNotNull(),
		models.FilmsAuditWhere.EpisodeNumber.IsNotNull(),
		qm.Offset(offset),
		qm.Limit(limit),
		qm.OrderBy(models.FilmsAuditColumns.ContributedAt+" desc"),
	).All(ctx, repo.exec)
	if err != nil {
		return nil, err
	}
	return audits, nil
}

func (repo *Repository) EpisodesAuditsCountBySeries(
	ctx context.Context,
	seriesID int,
) (int, error) {
	auditsCount, err := models.FilmsAudits(
		models.FilmsAuditWhere.SeriesID.EQ(null.IntFrom(seriesID)),
		models.FilmsAuditWhere.SeasonNumber.IsNotNull(),
		models.FilmsAuditWhere.EpisodeNumber.IsNotNull(),
	).Count(ctx, repo.exec)
	if err != nil {
		return 0, err
	}
	return int(auditsCount), nil
}
