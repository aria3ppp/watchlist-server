package repo

import (
	"context"
	"database/sql"

	"github.com/aria3ppp/watchlist-server/internal/models"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func (repo *Repository) UserGet(
	ctx context.Context,
	id int,
) (*models.User, error) {
	user, err := models.Users(models.UserWhere.ID.EQ(id)).One(ctx, repo.exec)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoRecord
		}
		return nil, err
	}
	return user, nil
}

func (repo *Repository) UserGetByEmail(
	ctx context.Context,
	email string,
) (*models.User, error) {
	user, err := models.Users(
		models.UserWhere.Email.EQ(email),
	).One(ctx, repo.exec)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoRecord
		}
		return nil, err
	}
	return user, nil
}

func (repo *Repository) UsersCount(ctx context.Context) (int, error) {
	nUsers, err := models.Users().Count(ctx, repo.exec)
	return int(nUsers), err
}

func (repo *Repository) UserCreate(
	ctx context.Context,
	user *models.User,
) error {
	return user.Insert(ctx, repo.exec, boil.Infer())
}

func (repo *Repository) UserUpdate(
	ctx context.Context,
	id int,
	cols map[string]any,
) error {
	rowsAff, err := models.Users(
		models.UserWhere.ID.EQ(id),
	).UpdateAll(
		ctx,
		repo.exec,
		cols,
	)
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return ErrNoRecord
	}
	return nil
}

func (repo *Repository) UserDelete(ctx context.Context, id int) error {
	rowsAff, err := models.Users(
		models.UserWhere.ID.EQ(id),
	).DeleteAll(ctx, repo.exec)
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return ErrNoRecord
	}
	return nil
}
