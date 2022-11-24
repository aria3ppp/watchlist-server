package repo

import (
	"context"
	"database/sql"

	"github.com/aria3ppp/watchlist-server/internal/models"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func (repo *Repository) MovieGet(
	ctx context.Context,
	id int,
) (*models.Film, error) {
	movie, err := models.Films(
		models.FilmWhere.ID.EQ(id),
		models.FilmWhere.SeriesID.IsNull(),
		models.FilmWhere.SeasonNumber.IsNull(),
		models.FilmWhere.EpisodeNumber.IsNull(),
	).One(ctx, repo.exec)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoRecord
		}
		return nil, err
	}
	return movie, nil
}

func (repo *Repository) MoviesGetAll(
	ctx context.Context,
	offset, limit int,
) ([]*models.Film, error) {
	movies, err := models.Films(
		models.FilmWhere.SeriesID.IsNull(),
		models.FilmWhere.SeasonNumber.IsNull(),
		models.FilmWhere.EpisodeNumber.IsNull(),
		qm.Offset(offset),
		qm.Limit(limit),
		qm.OrderBy(models.FilmColumns.ID),
	).All(ctx, repo.exec)
	if err != nil {
		return nil, err
	}
	return movies, nil
}

func (repo *Repository) MoviesCount(ctx context.Context) (int, error) {
	nMovies, err := models.Films(
		models.FilmWhere.SeriesID.IsNull(),
		models.FilmWhere.SeasonNumber.IsNull(),
		models.FilmWhere.EpisodeNumber.IsNull(),
	).Count(ctx, repo.exec)
	return int(nMovies), err
}

func (repo *Repository) MovieCreate(
	ctx context.Context,
	contributorID int,
	movie *models.Film,
) error {
	movie.ContributedBy = contributorID
	return movie.Insert(ctx, repo.exec, boil.Infer())
}

func (repo *Repository) MovieUpdate(
	ctx context.Context,
	movieID int,
	contributorID int,
	cols map[string]any,
) error {
	cols[models.FilmColumns.ContributedBy] = contributorID
	rowsAff, err := models.Films(
		models.FilmWhere.ID.EQ(movieID),
		models.FilmWhere.SeriesID.IsNull(),
		models.FilmWhere.SeasonNumber.IsNull(),
		models.FilmWhere.EpisodeNumber.IsNull(),
	).UpdateAll(ctx, repo.exec, cols)
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return ErrNoRecord
	}
	return nil
}

func (repo *Repository) MovieInvalidate(
	ctx context.Context,
	movieID int,
	contributorID int,
	invalidation string,
) error {
	rowsAff, err := models.Films(
		models.FilmWhere.ID.EQ(movieID),
		models.FilmWhere.SeriesID.IsNull(),
		models.FilmWhere.SeasonNumber.IsNull(),
		models.FilmWhere.EpisodeNumber.IsNull(),
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

func (repo *Repository) MovieAuditsGetAll(
	ctx context.Context,
	id int,
	offset, limit int,
) ([]*models.FilmsAudit, error) {
	audits, err := models.FilmsAudits(
		models.FilmsAuditWhere.ID.EQ(id),
		models.FilmsAuditWhere.SeriesID.IsNull(),
		models.FilmsAuditWhere.SeasonNumber.IsNull(),
		models.FilmsAuditWhere.EpisodeNumber.IsNull(),
		qm.Offset(offset),
		qm.Limit(limit),
		qm.OrderBy(models.FilmsAuditColumns.ContributedAt+" desc"),
	).All(ctx, repo.exec)
	if err != nil {
		return nil, err
	}
	return audits, nil
}

func (repo *Repository) MovieAuditsCount(
	ctx context.Context,
	id int,
) (int, error) {
	auditsCount, err := models.FilmsAudits(
		models.FilmsAuditWhere.ID.EQ(id),
		models.FilmsAuditWhere.SeriesID.IsNull(),
		models.FilmsAuditWhere.SeasonNumber.IsNull(),
		models.FilmsAuditWhere.EpisodeNumber.IsNull(),
	).Count(ctx, repo.exec)
	if err != nil {
		return 0, err
	}
	return int(auditsCount), nil
}
