package repo

import (
	"context"

	"github.com/aria3ppp/watchlist-server/internal/models"
)

func (repo *Repository) FilmExists(ctx context.Context, filmID int) error {
	exists, err := models.Films(
		models.FilmWhere.ID.EQ(filmID),
	).Exists(ctx, repo.exec)
	if err != nil {
		return err
	}
	if !exists {
		return ErrNoRecord
	}
	return nil
}
