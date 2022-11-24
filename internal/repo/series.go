package repo

import (
	"context"
	"database/sql"

	"github.com/aria3ppp/watchlist-server/internal/models"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func (repo *Repository) SeriesGet(
	ctx context.Context,
	id int,
) (*models.Series, error) {
	serie, err := models.Serieses(
		models.SeriesWhere.ID.EQ(id),
	).One(ctx, repo.exec)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoRecord
		}
		return nil, err
	}
	return serie, nil
}

func (repo *Repository) SeriesesGetAll(
	ctx context.Context,
	offset, limit int,
) ([]*models.Series, error) {
	series, err := models.Serieses(
		qm.Offset(offset),
		qm.Limit(limit),
		qm.OrderBy(models.SeriesColumns.ID),
	).All(ctx, repo.exec)
	if err != nil {
		return nil, err
	}
	return series, nil
}

func (repo *Repository) SeriesesCount(ctx context.Context) (int, error) {
	nSerie, err := models.Serieses().Count(ctx, repo.exec)
	return int(nSerie), err
}

func (repo *Repository) SeriesCreate(
	ctx context.Context,
	contributorID int,
	series *models.Series,
) error {
	series.ContributedBy = contributorID
	return series.Insert(ctx, repo.exec, boil.Infer())
}

func (repo *Repository) SeriesUpdate(
	ctx context.Context,
	serieID int,
	contributorID int,
	cols map[string]any,
) error {
	cols[models.SeriesColumns.ContributedBy] = contributorID
	rowsAff, err := models.Serieses(
		models.SeriesWhere.ID.EQ(serieID),
	).UpdateAll(ctx, repo.exec, cols)
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return ErrNoRecord
	}
	return nil
}

func (repo *Repository) SeriesInvalidate(
	ctx context.Context,
	serieID int,
	contributorID int,
	invalidation string,
) error {
	rowsAff, err := models.Serieses(
		models.SeriesWhere.ID.EQ(serieID),
	).UpdateAll(
		ctx,
		repo.exec,
		map[string]any{
			models.SeriesColumns.Invalidation:  invalidation,
			models.SeriesColumns.ContributedBy: contributorID,
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

func (repo *Repository) SeriesAuditsGetAll(
	ctx context.Context,
	id int,
	offset, limit int,
) ([]*models.SeriesesAudit, error) {
	audits, err := models.SeriesesAudits(
		models.SeriesesAuditWhere.ID.EQ(id),
		qm.Offset(offset),
		qm.Limit(limit),
		qm.OrderBy(models.SeriesesAuditColumns.ContributedAt+" desc"),
	).All(ctx, repo.exec)
	if err != nil {
		return nil, err
	}
	return audits, nil
}

func (repo *Repository) SeriesAuditsCount(
	ctx context.Context,
	id int,
) (int, error) {
	auditsCount, err := models.SeriesesAudits(
		models.SeriesesAuditWhere.ID.EQ(id),
	).Count(ctx, repo.exec)
	if err != nil {
		return 0, err
	}
	return int(auditsCount), nil
}
