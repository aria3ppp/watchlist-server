package app

import (
	"context"
	"io"
	"time"

	"github.com/aria3ppp/watchlist-server/internal/auth"
	"github.com/aria3ppp/watchlist-server/internal/dto"
	"github.com/aria3ppp/watchlist-server/internal/hasher"
	"github.com/aria3ppp/watchlist-server/internal/models"
	"github.com/aria3ppp/watchlist-server/internal/repo"
	"github.com/aria3ppp/watchlist-server/internal/storage"
)

func (app *Application) UserGet(
	ctx context.Context,
	id int,
) (*models.User, error) {
	user, err := app.repo.UserGet(ctx, id)
	if err != nil {
		if err == repo.ErrNoRecord {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return user, nil
}

// -----------------------------------------------------------------------------

func (app *Application) UserCreate(
	ctx context.Context,
	req *dto.UserCreateRequest,
) (userID int, err error) {
	err = app.repo.Tx(
		ctx,
		nil,
		func(ctx context.Context, tx repo.Service) error {
			// check email have not been used
			_, err := tx.UserGetByEmail(ctx, req.Email)
			if err == nil {
				return ErrUsedEmail
			}
			if err != repo.ErrNoRecord {
				return err
			}

			// hash the request password
			passwordHash, err := app.hasher.GenerateHash([]byte(req.Password))
			if err != nil {
				return err
			}

			// create the user
			insertUser := &models.User{
				Email:        req.Email,
				PasswordHash: string(passwordHash),
				FirstName:    req.FirstName,
				LastName:     req.LastName,
				Bio:          req.Bio,
				Birthdate:    req.Birthdate,
			}

			if err := tx.UserCreate(ctx, insertUser); err != nil {
				return err
			}

			// set the user id
			userID = insertUser.ID

			return nil
		},
	)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

//------------------------------------------------------------------------------

func (app *Application) UserLogin(
	ctx context.Context,
	req *dto.UserLoginRequest,
) (resp *dto.UserLoginResponse, err error) {
	err = app.repo.Tx(
		ctx,
		nil,
		func(ctx context.Context, tx repo.Service) error {
			// get user by provided email address
			user, err := tx.UserGetByEmail(ctx, req.Email)
			if err != nil {
				if err == repo.ErrNoRecord {
					return ErrNotFound
				}
				return err
			}

			// check provided password matches user password
			err = app.hasher.CompareHash(
				[]byte(user.PasswordHash),
				[]byte(req.Password),
			)
			if err != nil {
				if err == hasher.ErrMismatchedHash {
					return ErrIncorrectPassword
				}
				return err
			}

			// generate token
			jwtToken, jwtTokenExpiresAt, err := app.auth.GenerateJwtToken(
				&auth.Payload{UserID: user.ID},
			)
			if err != nil {
				return err
			}
			refreshToken, refreshTokenExpiresAt, err := app.auth.GenerateRefreshToken()
			if err != nil {
				return err
			}

			// hash and then save the refresh token
			refreshTokenHash, err := app.hasher.GenerateHash(
				[]byte(refreshToken),
			)
			if err != nil {
				return err
			}

			err = tx.TokenCreate(ctx, &models.Token{
				TokenHash: string(refreshTokenHash),
				UserID:    user.ID,
				ExpiresAt: refreshTokenExpiresAt,
			})
			if err != nil {
				return err
			}

			// set response
			resp = &dto.UserLoginResponse{
				UserRefreshResponse: dto.UserRefreshResponse{
					JwtToken:     jwtToken,
					JwtExpiresAt: jwtTokenExpiresAt.Unix(),
				},
				RefreshToken:     refreshToken,
				RefreshExpiresAt: refreshTokenExpiresAt.Unix(),
				UserID:           user.ID,
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

//------------------------------------------------------------------------------

func (app *Application) UserLogout(
	ctx context.Context,
	userID int,
	refreshToken string,
) error {
	err := app.repo.Tx(
		ctx,
		nil,
		func(ctx context.Context, tx repo.Service) error {
			// get token
			token, err := tx.TokenGet(ctx, userID, refreshToken)
			if err != nil {
				if err == repo.ErrNoRecord {
					return ErrNotFound
				}
				return err
			}
			// invalidate token
			err = tx.TokenUpdate(ctx, token.ID, map[string]any{
				models.TokenColumns.ExpiresAt: time.Now(),
			})
			if err != nil {
				// repo.ErrNoRecord have been checked before at top?
				return err
			}
			return nil
		},
	)
	return err
}

// -----------------------------------------------------------------------------

func (app *Application) UserRefreshToken(
	ctx context.Context,
	userID int,
	refreshToken string,
) (resp *dto.UserRefreshResponse, err error) {
	// check token exists
	token, err := app.repo.TokenGet(ctx, userID, refreshToken)
	if err != nil {
		if err == repo.ErrNoRecord {
			return nil, ErrNotFound
		}
		return nil, err
	}
	// create the new jwt token
	jwtToken, expiresAt, err := app.auth.GenerateJwtToken(
		&auth.Payload{UserID: token.UserID},
	)
	if err != nil {
		return nil, err
	}
	// return jwt token
	return &dto.UserRefreshResponse{
		JwtToken:     jwtToken,
		JwtExpiresAt: expiresAt.Unix(),
	}, nil
}

//------------------------------------------------------------------------------

func (app *Application) UserUpdate(
	ctx context.Context,
	userID int,
	req *dto.UserUpdateRequest,
) error {
	// build user columns to update
	columns := userUpdateRequestToValidMap(req)

	// update user
	if err := app.repo.UserUpdate(ctx, userID, columns); err != nil {
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

func (app *Application) UserEmailUpdate(
	ctx context.Context,
	userID int,
	req *dto.UserEmailUpdateRequest,
) error {
	// update email
	if err := app.repo.UserUpdate(
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

func (app *Application) UserPasswordUpdate(
	ctx context.Context,
	userID int,
	req *dto.UserPasswordUpdateRequest,
) error {
	// check new password is not the same as current one
	if req.NewPassword == req.CurrentPassword {
		return ErrSamePassword
	}

	err := app.repo.Tx(
		ctx,
		nil,
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
			err = app.hasher.CompareHash(
				[]byte(user.PasswordHash),
				[]byte(req.CurrentPassword),
			)
			if err != nil {
				if err == hasher.ErrMismatchedHash {
					return ErrIncorrectPassword
				}
				return err
			}

			// hash new password
			newPasswordHash, err := app.hasher.GenerateHash(
				[]byte(req.NewPassword),
			)
			if err != nil {
				return err
			}

			// replace new password
			if err = tx.UserUpdate(
				ctx,
				userID,
				map[string]any{
					models.UserColumns.PasswordHash: string(newPasswordHash),
				},
			); err != nil {
				// TODO: as we running in transaction; user existance have already been checked by UserGet?
				// TODO: looks like we should run at least in RepeatableRead isolation level to ensure there's no phantom read phenomena
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

func (app *Application) UserDelete(
	ctx context.Context,
	userID int,
	req *dto.UserDeleteRequest,
) error {
	err := app.repo.Tx(
		ctx,
		nil,
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
			err = app.hasher.CompareHash(
				[]byte(user.PasswordHash),
				[]byte(req.Password),
			)
			if err != nil {
				if err == hasher.ErrMismatchedHash {
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

func (app *Application) UserPutAvatar(
	ctx context.Context,
	userID int,
	avatar io.Reader,
	options *storage.PutOptions,
) (uri string, err error) {
	// put file
	uri, err = app.storage.PutFile(ctx, avatar, options)
	if err != nil {
		return "", err
	}
	// update user avatar
	err = app.repo.UserUpdate(ctx, userID, map[string]any{
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
