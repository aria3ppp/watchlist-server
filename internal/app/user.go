package app

import (
	"context"
	"io"

	"github.com/aria3ppp/watchlist-server/internal/dto"
	"github.com/aria3ppp/watchlist-server/internal/hasher"
	"github.com/aria3ppp/watchlist-server/internal/models"
	"github.com/aria3ppp/watchlist-server/internal/repo"
	"github.com/aria3ppp/watchlist-server/internal/storage"
	"github.com/aria3ppp/watchlist-server/internal/token"
	"golang.org/x/crypto/bcrypt"
)

func (a *Application) UserGet(
	ctx context.Context,
	id int,
) (*models.User, error) {
	user, err := a.repository.UserGet(ctx, id)
	if err != nil {
		if err == repo.ErrNoRecord {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return user, nil
}

//------------------------------------------------------------------------------

func (a *Application) UserCreate(
	ctx context.Context,
	req *dto.UserCreateRequest,
) (userID int, err error) {
	err = a.repository.Transaction(
		ctx,
		func(ctx context.Context, tx repo.Service) error {
			// check email not used already
			_, err := tx.UserGetByEmail(ctx, req.Email)
			if err == nil {
				return ErrEmailAlreadyUsed
			}
			if err != repo.ErrNoRecord {
				return err
			}

			// hash the request password
			hashedPassword, err := a.hasher.GenerateFromPassword(
				[]byte(req.Password),
				bcrypt.DefaultCost,
			)
			if err != nil {
				return err
			}

			// prepare the user
			insertUser := &models.User{
				Email:          req.Email,
				HashedPassword: string(hashedPassword),
				FirstName:      req.FirstName,
				LastName:       req.LastName,
				Bio:            req.Bio,
				Birthdate:      req.Birthdate,
			}

			// create the user
			if err := tx.UserCreate(ctx, insertUser); err != nil {
				return err
			}

			// pass the user id
			userID = insertUser.ID

			return nil
		},
	)

	return
}

//------------------------------------------------------------------------------

func (a *Application) UserLogin(
	ctx context.Context,
	req *dto.UserLoginRequest,
) (accessToken string, refreshToken string, err error) {
	// get user by provided email address
	user, err := a.repository.UserGetByEmail(ctx, req.Email)
	if err != nil {
		if err == repo.ErrNoRecord {
			return "", "", ErrNotFound
		}
		return "", "", err
	}

	// check provided password matches user password
	err = a.hasher.CompareHashAndPassword(
		[]byte(user.HashedPassword),
		[]byte(req.Password),
	)
	if err != nil {
		if err == hasher.ErrMismatchedHashAndPassword {
			return "", "", ErrIncorrectPassword
		}
		return "", "", err
	}

	// generate tokens
	tAccess, err := a.token.GenerateAccessToken(
		&token.Payload{UserID: user.ID},
	)
	if err != nil {
		return "", "", err
	}
	tRefresh, err := a.token.GenerateRefreshToken(
		&token.Payload{UserID: user.ID},
	)
	if err != nil {
		return "", "", err
	}

	return tAccess, tRefresh, nil
}

//------------------------------------------------------------------------------

func (a *Application) UserRefreshToken(
	ctx context.Context,
	refreshToken string,
) (string, error) {
	// validate request refresh token
	payload, err := a.token.ValidateToken(refreshToken)
	if err != nil {
		if err == token.ErrInvalidToken {
			return "", ErrTokenInvalid
		}
		return "", err
	}
	// generate new access token
	return a.token.GenerateAccessToken(payload)
}

//------------------------------------------------------------------------------

func (a *Application) UserUpdate(
	ctx context.Context,
	userID int,
	req *dto.UserUpdateRequest,
) error {
	// build user columns to update
	columns := userUpdateRequestToValidMap(req)

	// update user
	if err := a.repository.UserUpdate(ctx, userID, columns); err != nil {
		if err == repo.ErrNoRecord {
			return ErrNotFound
		}
		return err
	}

	return nil
}

func userUpdateRequestToValidMap(
	req *dto.UserUpdateRequest,
) map[string]any {
	m := make(map[string]any)
	if req.FirstName.Valid {
		m[models.UserColumns.FirstName] = req.FirstName.String
	}
	if req.LastName.Valid {
		m[models.UserColumns.LastName] = req.LastName.String
	}
	if req.Bio.Valid {
		m[models.UserColumns.Bio] = req.Bio.String
	}
	if req.Birthdate.Valid {
		m[models.UserColumns.Birthdate] = req.Birthdate.Time
	}
	return m
}

//------------------------------------------------------------------------------

func (a *Application) UserEmailUpdate(
	ctx context.Context,
	userID int,
	req *dto.UserEmailUpdateRequest,
) error {
	// update email
	if err := a.repository.UserUpdate(
		ctx,
		userID,
		map[string]any{
			models.UserColumns.Email: req.Email,
		},
	); err != nil {
		if err == repo.ErrNoRecord {
			return ErrNotFound
		}
		return err
	}

	return nil
}

//------------------------------------------------------------------------------

func (a *Application) UserPasswordUpdate(
	ctx context.Context,
	userID int,
	req *dto.UserPasswordUpdateRequest,
) error {
	// check new password is not the same as current one
	if req.NewPassword == req.CurrentPassword {
		return ErrSameNewPassword
	}

	err := a.repository.Transaction(
		ctx,
		func(ctx context.Context, tx repo.Service) error {
			// check user with this id exists
			user, err := tx.UserGet(ctx, userID)
			if err != nil {
				if err == repo.ErrNoRecord {
					return ErrNotFound
				}
				return err
			}

			// check current password match
			err = a.hasher.CompareHashAndPassword(
				[]byte(user.HashedPassword),
				[]byte(req.CurrentPassword),
			)
			if err != nil {
				if err == hasher.ErrMismatchedHashAndPassword {
					return ErrIncorrectPassword
				}
				return err
			}

			// hash new password
			hashedNewPassword, err := a.hasher.GenerateFromPassword(
				[]byte(req.NewPassword),
				bcrypt.DefaultCost,
			)
			if err != nil {
				return err
			}

			// replace new password
			if err = tx.UserUpdate(
				ctx,
				userID,
				map[string]any{
					models.UserColumns.HashedPassword: string(hashedNewPassword),
				},
			); err != nil {
				// as we running in transaction; user existance have already been checked by UserGet
				// if err == repo.ErrNoRecord {
				// 	return ErrNotFound
				// }
				return err
			}

			return nil
		},
	)

	return err
}

//------------------------------------------------------------------------------

func (a *Application) UserDelete(
	ctx context.Context,
	userID int,
	req *dto.UserDeleteRequest,
) error {
	err := a.repository.Transaction(
		ctx,
		func(ctx context.Context, tx repo.Service) error {
			// check user with this id exists
			user, err := tx.UserGet(ctx, userID)
			if err != nil {
				if err == repo.ErrNoRecord {
					return ErrNotFound
				}
				return err
			}

			// check password match
			err = a.hasher.CompareHashAndPassword(
				[]byte(user.HashedPassword),
				[]byte(req.Password),
			)
			if err != nil {
				if err == hasher.ErrMismatchedHashAndPassword {
					return ErrIncorrectPassword
				}
				return err
			}

			// delete user
			if err = tx.UserDelete(ctx, userID); err != nil {
				// as we running in transaction; user existance have already been checked by UserGet
				// if err == repo.ErrNoRecord {
				// 	return ErrNotFound
				// }
				return err
			}

			return nil
		},
	)

	return err
}

//------------------------------------------------------------------------------

func (a *Application) UserPutAvatar(
	ctx context.Context,
	userID int,
	avatar io.Reader,
	options *storage.PutOptions,
) (uri string, err error) {
	// put file
	uri, err = a.storage.PutFile(ctx, avatar, options)
	if err != nil {
		return "", err
	}
	// update user avatar
	err = a.repository.UserUpdate(ctx, userID, map[string]any{
		models.UserColumns.Avatar: uri,
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
