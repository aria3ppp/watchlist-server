package app_test

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"
	_ "unsafe"

	"github.com/aria3ppp/watchlist-server/internal/app"
	"github.com/aria3ppp/watchlist-server/internal/auth"
	"github.com/aria3ppp/watchlist-server/internal/auth/mock_auth"
	"github.com/aria3ppp/watchlist-server/internal/dto"
	"github.com/aria3ppp/watchlist-server/internal/hasher"
	"github.com/aria3ppp/watchlist-server/internal/hasher/mock_hasher"
	"github.com/aria3ppp/watchlist-server/internal/models"
	"github.com/aria3ppp/watchlist-server/internal/repo"
	"github.com/aria3ppp/watchlist-server/internal/repo/mock_repo"
	"github.com/aria3ppp/watchlist-server/internal/storage"
	"github.com/aria3ppp/watchlist-server/internal/storage/mock_storage"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/null/v8"
)

func TestUserGet(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var (
		id       = 1
		expError = errors.New("error")
		expUser  = &models.User{Email: "email"}
	)

	type GetExp struct {
		series *models.User
		err    error
	}
	type Get struct {
		exp GetExp
	}
	type Exp struct {
		user *models.User
		err  error
	}
	type TestCase struct {
		name string
		get  Get
		exp  Exp
	}

	testCases := []TestCase{
		{
			name: "error",
			get: Get{
				exp: GetExp{
					series: nil,
					err:    expError,
				},
			},
			exp: Exp{
				user: nil,
				err:  expError,
			},
		},
		{
			name: "not found",
			get: Get{
				exp: GetExp{
					series: nil,
					err:    repo.ErrNoRecord,
				},
			},
			exp: Exp{
				user: nil,
				err:  app.ErrNotFound,
			},
		},
		{
			name: "ok",
			get: Get{
				exp: GetExp{
					series: expUser,
					err:    nil,
				},
			},
			exp: Exp{
				user: expUser,
				err:  nil,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require := require.New(t)

			controller := gomock.NewController(t)
			mockRepo := mock_repo.NewMockServiceTx(controller)

			mockRepo.EXPECT().
				UserGet(ctx, id).
				Return(tc.get.exp.series, tc.get.exp.err)

			app := app.NewApplication(mockRepo, nil, nil, nil, nil)

			user, err := app.UserGet(ctx, id)
			require.Equal(tc.exp.err, err)
			require.Equal(tc.exp.user, user)
		})
	}
}

func TestUserCreate(t *testing.T) {
	t.Parallel()

	var (
		ctx = context.Background()

		req = &dto.UserCreateRequest{
			Email: "email",
		}
		expHash = "hash"
		user    = &models.User{
			Email:        req.Email,
			PasswordHash: expHash,
			FirstName:    req.FirstName,
			LastName:     req.LastName,
			Bio:          req.Bio,
			Birthdate:    req.Birthdate,
		}
		expNoRecordError             = repo.ErrNoRecord
		expEmailAlreadyUsedError     = app.ErrUsedEmail
		expUserGetByEmailError       = errors.New("UserGetByEmail error")
		expGenerateFromPasswordError = errors.New("GenerateFromPassword error")
		expUserCreateError           = errors.New("UserCreate error")
	)

	type UserGetByEmailExp struct {
		user *models.User
		err  error
	}
	type PasswordHashExp struct {
		hash []byte
		err  error
	}
	type TxExp struct {
		err error
	}
	type Tx struct {
		exp TxExp
	}
	type UserGetByEmail struct {
		exp UserGetByEmailExp
	}
	type PasswordHash struct {
		exp PasswordHashExp
	}
	type CreateExp struct {
		err error
	}
	type Create struct {
		exp CreateExp
	}
	type Exp struct {
		userID int
		err    error
	}
	type TestCase struct {
		name           string
		tx             Tx
		userGetByEmail UserGetByEmail
		passwordHash   PasswordHash
		create         Create
		exp            Exp
	}

	testCases := []TestCase{
		{
			name: "email address already been used",
			tx: Tx{
				exp: TxExp{
					err: expEmailAlreadyUsedError,
				},
			},
			userGetByEmail: UserGetByEmail{
				exp: UserGetByEmailExp{
					user: &models.User{},
					err:  nil,
				},
			},
			exp: Exp{
				userID: 0,
				err:    expEmailAlreadyUsedError,
			},
		},

		{
			name: "UserGetByEmail error",
			tx: Tx{
				exp: TxExp{
					err: expUserGetByEmailError,
				},
			},
			userGetByEmail: UserGetByEmail{
				exp: UserGetByEmailExp{
					user: nil,
					err:  expUserGetByEmailError,
				},
			},
			exp: Exp{
				userID: 0,
				err:    expUserGetByEmailError,
			},
		},

		{
			name: "GenerateFromPassword error",
			tx: Tx{
				exp: TxExp{
					err: expGenerateFromPasswordError,
				},
			},
			userGetByEmail: UserGetByEmail{
				exp: UserGetByEmailExp{
					user: nil,
					err:  expNoRecordError,
				},
			},
			passwordHash: PasswordHash{
				exp: PasswordHashExp{
					hash: nil,
					err:  expGenerateFromPasswordError,
				},
			},
			exp: Exp{
				userID: 0,
				err:    expGenerateFromPasswordError,
			},
		},

		{
			name: "UserCreate error",
			tx: Tx{
				exp: TxExp{
					err: expUserCreateError,
				},
			},
			userGetByEmail: UserGetByEmail{
				exp: UserGetByEmailExp{
					user: nil,
					err:  expNoRecordError,
				},
			},
			passwordHash: PasswordHash{
				exp: PasswordHashExp{
					hash: []byte(expHash),
					err:  nil,
				},
			},
			create: Create{
				exp: CreateExp{
					err: expUserCreateError,
				},
			},
			exp: Exp{
				userID: 0,
				err:    expUserCreateError,
			},
		},

		{
			name: "ok",
			tx: Tx{
				exp: TxExp{
					err: nil,
				},
			},
			userGetByEmail: UserGetByEmail{
				exp: UserGetByEmailExp{
					user: nil,
					err:  expNoRecordError,
				},
			},
			passwordHash: PasswordHash{
				exp: PasswordHashExp{
					hash: []byte(expHash),
				},
			},
			create: Create{
				exp: CreateExp{
					err: nil,
				},
			},
			exp: Exp{
				userID: 1,
				err:    nil,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require := require.New(t)

			controller := gomock.NewController(t)
			mockRepo := mock_repo.NewMockServiceTx(controller)
			mockHasher := mock_hasher.NewMockInterface(controller)

			txCall := mockRepo.EXPECT().
				Tx(ctx, nil, gomock.Any()).
				Do(func(ctx context.Context, opts *sql.TxOptions, fn func(_ context.Context, _ repo.Service) error) {
					fn(ctx, mockRepo)
				}).
				Return(tc.tx.exp.err)

			userGetByEmailCall := mockRepo.EXPECT().
				UserGetByEmail(ctx, req.Email).
				Return(tc.userGetByEmail.exp.user, tc.userGetByEmail.exp.err).
				After(txCall)

			if tc.userGetByEmail.exp.err == repo.ErrNoRecord {
				hashCall := mockHasher.EXPECT().
					GenerateHash([]byte(req.Password)).
					Return(tc.passwordHash.exp.hash, tc.passwordHash.exp.err).
					After(userGetByEmailCall)

				if tc.passwordHash.exp.err == nil {
					mockRepo.EXPECT().
						UserCreate(ctx, user).
						Do(func(_ context.Context, user *models.User) {
							user.ID = tc.exp.userID
						}).
						Return(tc.create.exp.err).
						After(hashCall)
				}
			}

			app := app.NewApplication(mockRepo, nil, nil, mockHasher, nil)

			userID, err := app.UserCreate(ctx, req)
			require.Equal(tc.exp.err, err)
			require.Equal(tc.exp.userID, userID)
		})
	}
}

func TestUserLogin(t *testing.T) {
	t.Parallel()

	var (
		ctx = context.Background()

		req = &dto.UserLoginRequest{
			Email:    "email",
			Password: "pass",
		}
		expUser = &models.User{
			ID:           1,
			Email:        req.Email,
			PasswordHash: "hash",
		}
		payload                = &auth.Payload{UserID: 1}
		expNoRecordError       = repo.ErrNoRecord
		expEmailNotFoundError  = app.ErrNotFound
		expIncorrectPassword   = app.ErrIncorrectPassword
		expUserGetByEmailError = errors.New("UserGetByEmail error")
		expCompareHashError    = errors.New("CompareHash error")
		expJwtToken            = "jwt token"
		expJwtExpiresAt        = time.Now().Add(time.Minute * 10)
		expJwtTokenHash        = "jwt token hash"
		expRefreshToken        = "refresh token"
		expRefreshExpiresAt    = time.Now().Add(time.Hour * 200)
		expResp                = &dto.UserLoginResponse{
			UserRefreshResponse: dto.UserRefreshResponse{
				JwtToken:     expJwtToken,
				JwtExpiresAt: expJwtExpiresAt.Unix(),
			},
			RefreshToken:     expRefreshToken,
			RefreshExpiresAt: expRefreshExpiresAt.Unix(),
			UserID:           expUser.ID,
		}
		expGenerateJwtTokenError     = errors.New("GenerateJwtToken error")
		expGenerateRefreshTokenError = errors.New("GenerateRefreshToken error")
		expGenerateHashError         = errors.New("GenerateHash error")
		expTokenCreateError          = errors.New("TokenCreate error")
	)

	type TxExp struct {
		err error
	}
	type Tx struct {
		exp TxExp
	}
	type UserGetByEmailExp struct {
		user *models.User
		err  error
	}
	type GenerateRefreshTokenExp struct {
		token     string
		expiresAt time.Time
		err       error
	}
	type GenerateRefreshToken struct {
		exp GenerateRefreshTokenExp
	}
	type GenerateJwtTokenExp struct {
		token     string
		expiresAt time.Time
		err       error
	}
	type GenerateJwtToken struct {
		exp GenerateJwtTokenExp
	}
	type UserGetByEmail struct {
		exp UserGetByEmailExp
	}
	type CompareHashExp struct {
		err error
	}
	type CompareHash struct {
		exp CompareHashExp
	}
	type GenerateHashExp struct {
		hash []byte
		err  error
	}
	type GenerateHash struct {
		exp GenerateHashExp
	}
	type TokenCreateExp struct {
		err error
	}
	type TokenCreate struct {
		exp TokenCreateExp
	}
	type Exp struct {
		resp *dto.UserLoginResponse
		err  error
	}
	type TestCase struct {
		name                 string
		tx                   Tx
		userGetByEmail       UserGetByEmail
		compareHash          CompareHash
		generateJwtToken     GenerateJwtToken
		generateRefreshToken GenerateRefreshToken
		generateHash         GenerateHash
		tokenCreate          TokenCreate
		exp                  Exp
	}

	testCases := []TestCase{
		{
			name: "email not found",
			tx: Tx{
				exp: TxExp{
					err: expEmailNotFoundError,
				},
			},
			userGetByEmail: UserGetByEmail{
				exp: UserGetByEmailExp{
					user: nil,
					err:  expNoRecordError,
				},
			},
			exp: Exp{
				resp: nil,
				err:  expEmailNotFoundError,
			},
		},

		{
			name: "UserGetByEmail error",
			tx: Tx{
				exp: TxExp{
					err: expUserGetByEmailError,
				},
			},
			userGetByEmail: UserGetByEmail{
				exp: UserGetByEmailExp{
					user: nil,
					err:  expUserGetByEmailError,
				},
			},
			exp: Exp{
				resp: nil,
				err:  expUserGetByEmailError,
			},
		},

		{
			name: "incorrect password",
			tx: Tx{
				exp: TxExp{
					err: expIncorrectPassword,
				},
			},
			userGetByEmail: UserGetByEmail{
				exp: UserGetByEmailExp{
					user: expUser,
					err:  nil,
				},
			},
			compareHash: CompareHash{
				exp: CompareHashExp{
					err: hasher.ErrMismatchedHash,
				},
			},
			exp: Exp{
				resp: nil,
				err:  expIncorrectPassword,
			},
		},

		{
			name: "CompareHash error",
			tx: Tx{
				exp: TxExp{
					err: expCompareHashError,
				},
			},
			userGetByEmail: UserGetByEmail{
				exp: UserGetByEmailExp{
					user: expUser,
					err:  nil,
				},
			},
			compareHash: CompareHash{
				exp: CompareHashExp{
					err: expCompareHashError,
				},
			},
			exp: Exp{
				resp: nil,
				err:  expCompareHashError,
			},
		},

		{
			name: "GenerateJwtToken error",
			tx: Tx{
				exp: TxExp{
					err: expGenerateJwtTokenError,
				},
			},
			userGetByEmail: UserGetByEmail{
				exp: UserGetByEmailExp{
					user: expUser,
					err:  nil,
				},
			},
			compareHash: CompareHash{
				exp: CompareHashExp{
					err: nil,
				},
			},
			generateJwtToken: GenerateJwtToken{
				exp: GenerateJwtTokenExp{
					token:     "",
					expiresAt: time.Time{},
					err:       expGenerateJwtTokenError,
				},
			},
			exp: Exp{
				resp: nil,
				err:  expGenerateJwtTokenError,
			},
		},

		{
			name: "GenerateRefreshToken error",
			tx: Tx{
				exp: TxExp{
					err: expGenerateRefreshTokenError,
				},
			},
			userGetByEmail: UserGetByEmail{
				exp: UserGetByEmailExp{
					user: expUser,
					err:  nil,
				},
			},
			compareHash: CompareHash{
				exp: CompareHashExp{
					err: nil,
				},
			},
			generateJwtToken: GenerateJwtToken{
				exp: GenerateJwtTokenExp{
					token:     expJwtToken,
					expiresAt: expJwtExpiresAt,
					err:       nil,
				},
			},
			generateRefreshToken: GenerateRefreshToken{
				exp: GenerateRefreshTokenExp{
					token:     "",
					expiresAt: time.Time{},
					err:       expGenerateRefreshTokenError,
				},
			},
			exp: Exp{
				resp: nil,
				err:  expGenerateRefreshTokenError,
			},
		},

		{
			name: "GenerateHash error",
			tx: Tx{
				exp: TxExp{
					err: expGenerateHashError,
				},
			},
			userGetByEmail: UserGetByEmail{
				exp: UserGetByEmailExp{
					user: expUser,
					err:  nil,
				},
			},
			compareHash: CompareHash{
				exp: CompareHashExp{
					err: nil,
				},
			},
			generateJwtToken: GenerateJwtToken{
				exp: GenerateJwtTokenExp{
					token:     expJwtToken,
					expiresAt: expJwtExpiresAt,
					err:       nil,
				},
			},
			generateRefreshToken: GenerateRefreshToken{
				exp: GenerateRefreshTokenExp{
					token:     expRefreshToken,
					expiresAt: expRefreshExpiresAt,
					err:       nil,
				},
			},
			generateHash: GenerateHash{
				exp: GenerateHashExp{
					hash: nil,
					err:  expGenerateHashError,
				},
			},
			exp: Exp{
				resp: nil,
				err:  expGenerateHashError,
			},
		},

		{
			name: "TokenCreate error",
			tx: Tx{
				exp: TxExp{
					err: expTokenCreateError,
				},
			},
			userGetByEmail: UserGetByEmail{
				exp: UserGetByEmailExp{
					user: expUser,
					err:  nil,
				},
			},
			compareHash: CompareHash{
				exp: CompareHashExp{
					err: nil,
				},
			},
			generateJwtToken: GenerateJwtToken{
				exp: GenerateJwtTokenExp{
					token:     expJwtToken,
					expiresAt: expJwtExpiresAt,
					err:       nil,
				},
			},
			generateRefreshToken: GenerateRefreshToken{
				exp: GenerateRefreshTokenExp{
					token:     expRefreshToken,
					expiresAt: expRefreshExpiresAt,
					err:       nil,
				},
			},
			generateHash: GenerateHash{
				exp: GenerateHashExp{
					hash: []byte(expJwtTokenHash),
					err:  nil,
				},
			},
			tokenCreate: TokenCreate{
				exp: TokenCreateExp{
					err: expTokenCreateError,
				},
			},
			exp: Exp{
				resp: nil,
				err:  expTokenCreateError,
			},
		},

		{
			name: "ok",
			tx: Tx{
				exp: TxExp{
					err: nil,
				},
			},
			userGetByEmail: UserGetByEmail{
				exp: UserGetByEmailExp{
					user: expUser,
					err:  nil,
				},
			},
			compareHash: CompareHash{
				exp: CompareHashExp{
					err: nil,
				},
			},
			generateJwtToken: GenerateJwtToken{
				exp: GenerateJwtTokenExp{
					token:     expJwtToken,
					expiresAt: expJwtExpiresAt,
					err:       nil,
				},
			},
			generateRefreshToken: GenerateRefreshToken{
				exp: GenerateRefreshTokenExp{
					token:     expRefreshToken,
					expiresAt: expRefreshExpiresAt,
					err:       nil,
				},
			},
			generateHash: GenerateHash{
				exp: GenerateHashExp{
					hash: []byte(expJwtTokenHash),
					err:  nil,
				},
			},
			tokenCreate: TokenCreate{
				exp: TokenCreateExp{
					err: nil,
				},
			},
			exp: Exp{
				resp: expResp,
				err:  nil,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require := require.New(t)

			controller := gomock.NewController(t)
			mockRepo := mock_repo.NewMockServiceTx(controller)
			mockHasher := mock_hasher.NewMockInterface(controller)
			mockAuthInterface := mock_auth.NewMockInterface(controller)

			txCall := mockRepo.EXPECT().
				Tx(ctx, nil, gomock.Any()).
				Do(func(ctx context.Context, opts *sql.TxOptions, fn func(_ context.Context, _ repo.Service) error) {
					fn(ctx, mockRepo)
				}).
				Return(tc.tx.exp.err)

			userGetByEmailCall := mockRepo.EXPECT().
				UserGetByEmail(ctx, req.Email).
				Do(func(_ context.Context, _ string) {
					expUser.ID = payload.UserID
				}).
				Return(tc.userGetByEmail.exp.user, tc.userGetByEmail.exp.err).
				After(txCall)

			if tc.userGetByEmail.exp.err == nil {
				compateHashCall := mockHasher.EXPECT().
					CompareHash([]byte(expUser.PasswordHash), []byte(req.Password)).
					Return(tc.compareHash.exp.err).
					After(userGetByEmailCall)

				if tc.compareHash.exp.err == nil {
					generateJwtTokenCall := mockAuthInterface.EXPECT().
						GenerateJwtToken(payload).
						Return(tc.generateJwtToken.exp.token, tc.generateJwtToken.exp.expiresAt, tc.generateJwtToken.exp.err).
						After(compateHashCall)

					if tc.generateJwtToken.exp.err == nil {
						generateRefreshTokenCall := mockAuthInterface.EXPECT().
							GenerateRefreshToken().
							Return(tc.generateRefreshToken.exp.token, tc.generateRefreshToken.exp.expiresAt, tc.generateRefreshToken.exp.err).
							After(generateJwtTokenCall)

						if tc.generateRefreshToken.exp.err == nil {
							mockHasher.EXPECT().
								GenerateHash([]byte(tc.generateRefreshToken.exp.token)).
								Return(tc.generateHash.exp.hash, tc.generateHash.exp.err).
								After(generateRefreshTokenCall)

							if tc.generateHash.exp.err == nil {
								mockRepo.EXPECT().
									TokenCreate(ctx, &models.Token{
										TokenHash: string(
											tc.generateHash.exp.hash,
										),
										UserID:    tc.userGetByEmail.exp.user.ID,
										ExpiresAt: tc.generateRefreshToken.exp.expiresAt,
									}).
									Return(tc.tokenCreate.exp.err)
							}
						}
					}
				}
			}

			app := app.NewApplication(
				mockRepo,
				mockAuthInterface,
				nil,
				mockHasher,
				nil,
			)

			resp, err := app.UserLogin(ctx, req)
			require.Equal(tc.exp.err, err)
			require.Equal(tc.exp.resp, resp)
		})
	}
}

func TestUserLogout(t *testing.T) {
	t.Parallel()

	var (
		ctx = context.Background()

		userID              = 1
		refreshToken        = "refresh token"
		expToken            = &models.Token{UserID: userID}
		expTokenGetError    = errors.New("TokenGet error")
		expTokenUpdateError = errors.New("TokenUpdate error")
	)

	type TxExp struct {
		err error
	}
	type Tx struct {
		exp TxExp
	}
	type TokenUpdateExp struct {
		err error
	}
	type TokenUpdate struct {
		exp TokenUpdateExp
	}
	type TokenGetExp struct {
		token *models.Token
		err   error
	}
	type TokenGet struct {
		exp TokenGetExp
	}
	type Exp struct {
		err error
	}
	type TestCase struct {
		name        string
		tx          Tx
		tokenGet    TokenGet
		tokenUpdate TokenUpdate
		exp         Exp
	}

	testCases := []TestCase{
		{
			name: "not found",
			tx: Tx{
				exp: TxExp{
					err: app.ErrNotFound,
				},
			},
			tokenGet: TokenGet{
				exp: TokenGetExp{
					token: nil,
					err:   repo.ErrNoRecord,
				},
			},
			exp: Exp{
				err: app.ErrNotFound,
			},
		},

		{
			name: "TokenGet error",
			tx: Tx{
				exp: TxExp{
					err: expTokenGetError,
				},
			},
			tokenGet: TokenGet{
				exp: TokenGetExp{
					token: nil,
					err:   expTokenGetError,
				},
			},
			exp: Exp{
				err: expTokenGetError,
			},
		},

		{
			name: "TokenUpdate error",
			tx: Tx{
				exp: TxExp{
					err: expTokenUpdateError,
				},
			},
			tokenGet: TokenGet{
				exp: TokenGetExp{
					token: expToken,
					err:   nil,
				},
			},
			tokenUpdate: TokenUpdate{
				exp: TokenUpdateExp{
					err: expTokenUpdateError,
				},
			},
			exp: Exp{
				err: expTokenUpdateError,
			},
		},

		{
			name: "ok",
			tx: Tx{
				exp: TxExp{
					err: nil,
				},
			},
			tokenGet: TokenGet{
				exp: TokenGetExp{
					token: expToken,
					err:   nil,
				},
			},
			tokenUpdate: TokenUpdate{
				exp: TokenUpdateExp{
					err: nil,
				},
			},
			exp: Exp{
				err: nil,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require := require.New(t)

			controller := gomock.NewController(t)
			mockRepo := mock_repo.NewMockServiceTx(controller)

			txCall := mockRepo.EXPECT().
				Tx(ctx, nil, gomock.Any()).
				Do(func(ctx context.Context, opts *sql.TxOptions, fn func(_ context.Context, _ repo.Service) error) {
					fn(ctx, mockRepo)
				}).
				Return(tc.tx.exp.err)

			tokenGetCall := mockRepo.EXPECT().
				TokenGet(ctx, userID, refreshToken).
				Return(tc.tokenGet.exp.token, tc.tokenGet.exp.err).
				After(txCall)

			if tc.tokenGet.exp.err == nil {
				mockRepo.EXPECT().
					TokenUpdate(ctx, expToken.ID, gomock.Any()).
					Return(tc.tokenUpdate.exp.err).
					After(tokenGetCall)
			}

			app := app.NewApplication(
				mockRepo,
				nil,
				nil,
				nil,
				nil,
			)

			err := app.UserLogout(ctx, userID, refreshToken)
			require.Equal(tc.exp.err, err)
		})
	}
}

func TestUserRefreshToken(t *testing.T) {
	t.Parallel()

	var (
		ctx = context.Background()

		userID             = 1
		refreshToken       = "refresh token"
		expToken           = &models.Token{UserID: userID}
		expNewJwtToken     = "new jwt token"
		expNewJwtExpiresAt = time.Now().Add(time.Minute * 60)
		expResp            = &dto.UserRefreshResponse{
			JwtToken:     expNewJwtToken,
			JwtExpiresAt: expNewJwtExpiresAt.Unix(),
		}
		expTokenGetError         = errors.New("TokenGet error")
		expGenerateJwtTokenError = errors.New("GenerateJwtToken error")
	)

	type GenerateJwtTokenExp struct {
		token     string
		expiresAt time.Time
		err       error
	}
	type GenerateJwtToken struct {
		exp GenerateJwtTokenExp
	}
	type TokenGetExp struct {
		token *models.Token
		err   error
	}
	type TokenGet struct {
		exp TokenGetExp
	}
	type Exp struct {
		resp *dto.UserRefreshResponse
		err  error
	}
	type TestCase struct {
		name             string
		tokenGet         TokenGet
		generateJwtToken GenerateJwtToken
		exp              Exp
	}

	testCases := []TestCase{
		{
			name: "not found",
			tokenGet: TokenGet{
				exp: TokenGetExp{
					token: nil,
					err:   repo.ErrNoRecord,
				},
			},
			exp: Exp{
				resp: nil,
				err:  app.ErrNotFound,
			},
		},

		{
			name: "TokenGet error",
			tokenGet: TokenGet{
				exp: TokenGetExp{
					token: nil,
					err:   expTokenGetError,
				},
			},
			exp: Exp{
				resp: nil,
				err:  expTokenGetError,
			},
		},

		{
			name: "GenerateJwtToken error",
			tokenGet: TokenGet{
				exp: TokenGetExp{
					token: expToken,
					err:   nil,
				},
			},
			generateJwtToken: GenerateJwtToken{
				exp: GenerateJwtTokenExp{
					token:     "",
					expiresAt: time.Time{},
					err:       expGenerateJwtTokenError,
				},
			},
			exp: Exp{
				resp: nil,
				err:  expGenerateJwtTokenError,
			},
		},

		{
			name: "ok",
			tokenGet: TokenGet{
				exp: TokenGetExp{
					token: expToken,
					err:   nil,
				},
			},
			generateJwtToken: GenerateJwtToken{
				exp: GenerateJwtTokenExp{
					token:     expNewJwtToken,
					expiresAt: expNewJwtExpiresAt,
					err:       nil,
				},
			},
			exp: Exp{
				resp: expResp,
				err:  nil,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require := require.New(t)

			controller := gomock.NewController(t)
			mockRepo := mock_repo.NewMockServiceTx(controller)
			mockAuthInterface := mock_auth.NewMockInterface(controller)

			tokenGetCall := mockRepo.EXPECT().
				TokenGet(ctx, userID, refreshToken).
				Return(tc.tokenGet.exp.token, tc.tokenGet.exp.err)

			if tc.tokenGet.exp.err == nil {
				mockAuthInterface.EXPECT().
					GenerateJwtToken(&auth.Payload{UserID: expToken.UserID}).
					Return(tc.generateJwtToken.exp.token, tc.generateJwtToken.exp.expiresAt, tc.generateJwtToken.exp.err).
					After(tokenGetCall)
			}

			app := app.NewApplication(
				mockRepo,
				mockAuthInterface,
				nil,
				nil,
				nil,
			)

			resp, err := app.UserRefreshToken(ctx, userID, refreshToken)
			require.Equal(tc.exp.err, err)
			require.Equal(tc.exp.resp, resp)
		})
	}
}

//go:linkname userUpdateRequestToValidMap github.com/aria3ppp/watchlist-server/internal/app.userUpdateRequestToValidMap
func userUpdateRequestToValidMap(*dto.UserUpdateRequest) map[string]any

func TestUserUpdate(t *testing.T) {
	t.Parallel()

	var (
		ctx = context.Background()

		userID = 1
		req    = &dto.UserUpdateRequest{
			FirstName: null.StringFrom("first name"),
		}
		columns            = userUpdateRequestToValidMap(req)
		expNotFoundError   = app.ErrNotFound
		expUserUpdateError = errors.New("UserUpdate error")
	)

	type UserUpdateExp struct {
		err error
	}
	type UserUpdate struct {
		exp UserUpdateExp
	}
	type Exp struct {
		err error
	}
	type TestCase struct {
		name       string
		userUpdate UserUpdate
		exp        Exp
	}

	testCases := []TestCase{
		{
			name: "user not found",
			userUpdate: UserUpdate{
				exp: UserUpdateExp{
					err: repo.ErrNoRecord,
				},
			},
			exp: Exp{
				err: expNotFoundError,
			},
		},

		{
			name: "UserUpdate error",
			userUpdate: UserUpdate{
				exp: UserUpdateExp{
					err: expUserUpdateError,
				},
			},
			exp: Exp{
				err: expUserUpdateError,
			},
		},

		{
			name: "ok",
			userUpdate: UserUpdate{
				exp: UserUpdateExp{
					err: nil,
				},
			},
			exp: Exp{
				err: nil,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require := require.New(t)

			controller := gomock.NewController(t)
			mockRepo := mock_repo.NewMockServiceTx(controller)

			mockRepo.EXPECT().
				UserUpdate(ctx, userID, columns).
				Return(tc.userUpdate.exp.err)

			app := app.NewApplication(mockRepo, nil, nil, nil, nil)

			err := app.UserUpdate(ctx, userID, req)
			require.Equal(tc.exp.err, err)
		})
	}
}

func TestUserEmailUpdate(t *testing.T) {
	t.Parallel()

	var (
		ctx = context.Background()

		userID = 1
		req    = &dto.UserEmailUpdateRequest{
			Email: "email",
		}
		columns = map[string]any{
			models.UserColumns.Email: req.Email,
		}
		expNotFoundError   = app.ErrNotFound
		expUserUpdateError = errors.New("UserUpdate error")
	)

	type UserUpdateExp struct {
		err error
	}
	type UserUpdate struct {
		exp UserUpdateExp
	}
	type Exp struct {
		err error
	}
	type TestCase struct {
		name       string
		userUpdate UserUpdate
		exp        Exp
	}

	testCases := []TestCase{
		{
			name: "user not found",
			userUpdate: UserUpdate{
				exp: UserUpdateExp{
					err: repo.ErrNoRecord,
				},
			},
			exp: Exp{
				err: expNotFoundError,
			},
		},

		{
			name: "UserUpdate error",
			userUpdate: UserUpdate{
				exp: UserUpdateExp{
					err: expUserUpdateError,
				},
			},
			exp: Exp{
				err: expUserUpdateError,
			},
		},

		{
			name: "ok",
			userUpdate: UserUpdate{
				exp: UserUpdateExp{
					err: nil,
				},
			},
			exp: Exp{
				err: nil,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require := require.New(t)

			controller := gomock.NewController(t)
			mockRepo := mock_repo.NewMockServiceTx(controller)

			mockRepo.EXPECT().
				UserUpdate(ctx, userID, columns).
				Return(tc.userUpdate.exp.err)

			app := app.NewApplication(mockRepo, nil, nil, nil, nil)

			err := app.UserEmailUpdate(ctx, userID, req)
			require.Equal(tc.exp.err, err)
		})
	}
}

func TestUserPasswordUpdate(t *testing.T) {
	t.Parallel()

	var (
		ctx = context.Background()

		samePassword = ""
		req          = &dto.UserPasswordUpdateRequest{
			CurrentPassword: "pass",
			NewPassword:     "new pass",
		}
		userID  = 1
		expUser = &models.User{
			ID:           userID,
			Email:        "email",
			PasswordHash: "hash",
		}
		columns = map[string]any{
			models.UserColumns.PasswordHash: expUser.PasswordHash,
		}
		expSameNewPasswordError   = app.ErrSamePassword
		expNotFoundError          = app.ErrNotFound
		expUserGetError           = errors.New("UserGet error")
		expIncorrectPasswordError = app.ErrIncorrectPassword
		expCompareHashError       = errors.New("CompareHash error")
		expGenerateHashError      = errors.New("GenerateHash error")
		expUserUpdateError        = errors.New("UserUpdate error")
	)

	type UserGetExp struct {
		user *models.User
		err  error
	}
	type UserGet struct {
		exp UserGetExp
	}
	type CompareHashExp struct {
		err error
	}
	type CompareHash struct {
		exp CompareHashExp
	}
	type GenerateHashExp struct {
		passwordHash string
		err          error
	}
	type GenerateHash struct {
		exp GenerateHashExp
	}
	type UserUpdateExp struct {
		err error
	}
	type UserUpdate struct {
		exp UserUpdateExp
	}
	type TxExp struct {
		err error
	}
	type Tx struct {
		exp TxExp
	}
	type Exp struct {
		err error
	}
	type TestCase struct {
		name         string
		req          *dto.UserPasswordUpdateRequest
		tx           Tx
		userGet      UserGet
		compareHash  CompareHash
		generateHash GenerateHash
		userUpdate   UserUpdate
		exp          Exp
	}

	testCases := []TestCase{
		{
			name: "same password",
			req: &dto.UserPasswordUpdateRequest{
				CurrentPassword: samePassword,
				NewPassword:     samePassword,
			},
			exp: Exp{
				err: expSameNewPasswordError,
			},
		},

		{
			name: "user not found",
			req:  req,
			tx: Tx{
				exp: TxExp{
					err: expNotFoundError,
				},
			},
			userGet: UserGet{
				exp: UserGetExp{
					user: nil,
					err:  repo.ErrNoRecord,
				},
			},
			exp: Exp{
				err: expNotFoundError,
			},
		},

		{
			name: "UserGet error",
			req:  req,
			tx: Tx{
				exp: TxExp{
					err: expUserGetError,
				},
			},
			userGet: UserGet{
				exp: UserGetExp{
					user: nil,
					err:  expUserGetError,
				},
			},
			exp: Exp{
				err: expUserGetError,
			},
		},

		{
			name: "incorrect password",
			req:  req,
			tx: Tx{
				exp: TxExp{
					err: expIncorrectPasswordError,
				},
			},
			userGet: UserGet{
				exp: UserGetExp{
					user: expUser,
					err:  nil,
				},
			},
			compareHash: CompareHash{
				exp: CompareHashExp{
					err: hasher.ErrMismatchedHash,
				},
			},
			exp: Exp{
				err: expIncorrectPasswordError,
			},
		},

		{
			name: "CompareHash error",
			req:  req,
			tx: Tx{
				exp: TxExp{
					err: expCompareHashError,
				},
			},
			userGet: UserGet{
				exp: UserGetExp{
					user: expUser,
					err:  nil,
				},
			},
			compareHash: CompareHash{
				exp: CompareHashExp{
					err: expCompareHashError,
				},
			},
			exp: Exp{
				err: expCompareHashError,
			},
		},

		{
			name: "GenerateHash error",
			req:  req,
			tx: Tx{
				exp: TxExp{
					err: expGenerateHashError,
				},
			},
			userGet: UserGet{
				exp: UserGetExp{
					user: expUser,
					err:  nil,
				},
			},
			compareHash: CompareHash{
				exp: CompareHashExp{err: nil},
			},
			generateHash: GenerateHash{
				exp: GenerateHashExp{
					passwordHash: "",
					err:          expGenerateHashError,
				},
			},
			exp: Exp{
				err: expGenerateHashError,
			},
		},

		{
			name: "UserUpdate error",
			req:  req,
			tx: Tx{
				exp: TxExp{
					err: expUserUpdateError,
				},
			},
			userGet: UserGet{
				exp: UserGetExp{
					user: expUser,
					err:  nil,
				},
			},
			compareHash: CompareHash{
				exp: CompareHashExp{
					err: nil,
				},
			},
			generateHash: GenerateHash{
				exp: GenerateHashExp{
					passwordHash: expUser.PasswordHash,
					err:          nil,
				},
			},
			userUpdate: UserUpdate{
				exp: UserUpdateExp{
					err: expUserUpdateError,
				},
			},
			exp: Exp{
				err: expUserUpdateError,
			},
		},

		{
			name: "ok",
			req:  req,
			tx: Tx{
				exp: TxExp{
					err: nil,
				},
			},
			userGet: UserGet{
				exp: UserGetExp{
					user: expUser,
					err:  nil,
				},
			},
			compareHash: CompareHash{
				exp: CompareHashExp{
					err: nil,
				},
			},
			generateHash: GenerateHash{
				exp: GenerateHashExp{
					passwordHash: expUser.PasswordHash,
					err:          nil,
				},
			},
			userUpdate: UserUpdate{
				exp: UserUpdateExp{
					err: nil,
				},
			},
			exp: Exp{
				err: nil,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require := require.New(t)

			controller := gomock.NewController(t)
			mockRepo := mock_repo.NewMockServiceTx(controller)
			mockHasher := mock_hasher.NewMockInterface(controller)

			if tc.req.CurrentPassword != tc.req.NewPassword {
				txCall := mockRepo.EXPECT().
					Tx(ctx, nil, gomock.Any()).
					Do(func(ctx context.Context, opts *sql.TxOptions, fn func(_ context.Context, _ repo.Service) error) {
						fn(ctx, mockRepo)
					}).
					Return(tc.tx.exp.err)

				userGetCall := mockRepo.EXPECT().
					UserGet(ctx, userID).
					Return(tc.userGet.exp.user, tc.userGet.exp.err).
					After(txCall)

				if tc.userGet.exp.err == nil {
					compareHashCall := mockHasher.EXPECT().
						CompareHash([]byte(expUser.PasswordHash), []byte(req.CurrentPassword)).
						Return(tc.compareHash.exp.err).
						After(userGetCall)

					if tc.compareHash.exp.err == nil {
						generateHashCall := mockHasher.EXPECT().
							GenerateHash([]byte(req.NewPassword)).
							Return([]byte(tc.generateHash.exp.passwordHash), tc.generateHash.exp.err).
							After(compareHashCall)

						if tc.generateHash.exp.err == nil {
							mockRepo.EXPECT().
								UserUpdate(ctx, userID, columns).
								Return(tc.userUpdate.exp.err).
								After(generateHashCall)
						}
					}
				}
			}

			app := app.NewApplication(mockRepo, nil, nil, mockHasher, nil)

			err := app.UserPasswordUpdate(ctx, userID, tc.req)
			require.Equal(tc.exp.err, err)
		})
	}
}

func TestUserDelete(t *testing.T) {
	t.Parallel()

	var (
		ctx = context.Background()

		req = &dto.UserDeleteRequest{
			Password: "pass",
		}
		userID  = 1
		expUser = &models.User{
			ID:           userID,
			Email:        "email",
			PasswordHash: "hash",
		}
		expNotFoundError          = app.ErrNotFound
		expUserGetError           = errors.New("UserGet error")
		expIncorrectPasswordError = app.ErrIncorrectPassword
		expCompareHashError       = errors.New("CompareHash error")
		expUserDeleteError        = errors.New("UserDelete error")
	)

	type UserGetExp struct {
		user *models.User
		err  error
	}
	type UserGet struct {
		exp UserGetExp
	}
	type CompareHashExp struct {
		err error
	}
	type CompareHash struct {
		exp CompareHashExp
	}
	type UserDeleteExp struct {
		err error
	}
	type UserDelete struct {
		exp UserDeleteExp
	}
	type TxExp struct {
		err error
	}
	type Tx struct {
		exp TxExp
	}
	type Exp struct {
		err error
	}
	type TestCase struct {
		name        string
		tx          Tx
		userGet     UserGet
		compareHash CompareHash
		userDelete  UserDelete
		exp         Exp
	}

	testCases := []TestCase{
		{
			name: "user not found",
			tx: Tx{
				exp: TxExp{
					err: expNotFoundError,
				},
			},
			userGet: UserGet{
				exp: UserGetExp{
					user: nil,
					err:  repo.ErrNoRecord,
				},
			},
			exp: struct{ err error }{err: expNotFoundError},
		},

		{
			name: "UserGet error",
			tx: Tx{
				exp: TxExp{
					err: expUserGetError,
				},
			},
			userGet: UserGet{
				exp: UserGetExp{
					user: nil,
					err:  expUserGetError,
				},
			},
			exp: Exp{
				err: expUserGetError,
			},
		},

		{
			name: "incorrect password",
			tx: Tx{
				exp: TxExp{
					err: expIncorrectPasswordError,
				},
			},
			userGet: UserGet{
				exp: UserGetExp{
					user: expUser,
					err:  nil,
				},
			},
			compareHash: CompareHash{
				exp: CompareHashExp{
					err: hasher.ErrMismatchedHash,
				},
			},
			exp: Exp{
				err: expIncorrectPasswordError,
			},
		},

		{
			name: "CompareHashAndPassword error",
			tx: Tx{
				exp: TxExp{
					err: expCompareHashError,
				},
			},
			userGet: UserGet{
				exp: UserGetExp{
					user: expUser,
					err:  nil,
				},
			},
			compareHash: CompareHash{
				exp: CompareHashExp{
					err: expCompareHashError,
				},
			},
			exp: Exp{
				err: expCompareHashError,
			},
		},

		{
			name: "UserDelete error",
			tx: Tx{
				exp: TxExp{
					err: expUserDeleteError,
				},
			},
			userGet: UserGet{
				exp: UserGetExp{
					user: expUser,
					err:  nil,
				},
			},
			compareHash: CompareHash{
				exp: CompareHashExp{err: nil},
			},
			userDelete: UserDelete{
				exp: UserDeleteExp{
					err: expUserDeleteError,
				},
			},
			exp: Exp{
				err: expUserDeleteError,
			},
		},

		{
			name: "ok",
			tx: Tx{
				exp: TxExp{
					err: nil,
				},
			},
			userGet: UserGet{
				exp: UserGetExp{
					user: expUser,
					err:  nil,
				},
			},
			compareHash: CompareHash{
				exp: CompareHashExp{err: nil},
			},
			userDelete: UserDelete{
				exp: UserDeleteExp{
					err: nil,
				},
			},
			exp: Exp{
				err: nil,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require := require.New(t)

			controller := gomock.NewController(t)
			mockRepo := mock_repo.NewMockServiceTx(controller)
			mockHasher := mock_hasher.NewMockInterface(controller)

			txCall := mockRepo.EXPECT().
				Tx(ctx, nil, gomock.Any()).
				Do(func(ctx context.Context, opts *sql.TxOptions, fn func(_ context.Context, _ repo.Service) error) {
					fn(ctx, mockRepo)
				}).
				Return(tc.tx.exp.err)

			userGetCall := mockRepo.EXPECT().
				UserGet(ctx, userID).
				Return(tc.userGet.exp.user, tc.userGet.exp.err).
				After(txCall)

			if tc.userGet.exp.err == nil {
				compareHashCall := mockHasher.EXPECT().
					CompareHash([]byte(expUser.PasswordHash), []byte(req.Password)).
					Return(tc.compareHash.exp.err).
					After(userGetCall)

				if tc.compareHash.exp.err == nil {
					mockRepo.EXPECT().
						UserDelete(ctx, userID).
						Return(tc.userDelete.exp.err).
						After(compareHashCall)
				}
			}

			app := app.NewApplication(mockRepo, nil, nil, mockHasher, nil)

			err := app.UserDelete(ctx, userID, req)
			require.Equal(tc.exp.err, err)
		})
	}
}

func TestUserPutAvatar(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var (
		userID  = 1
		avatar  = strings.NewReader("avatar")
		options = &storage.PutOptions{}

		expUri             = "expected uri :/"
		expPutFileError    = errors.New("PutFile error")
		expUserUpdateError = errors.New("UserUpdate error")
	)

	type PutFileExp struct {
		uri string
		err error
	}
	type PutFile struct {
		exp PutFileExp
	}
	type UpdateUserExp struct {
		err error
	}
	type UpdateUser struct {
		exp UpdateUserExp
	}
	type Exp struct {
		uri string
		err error
	}
	testCases := []struct {
		name       string
		putFile    PutFile
		updateUser UpdateUser
		exp        Exp
	}{
		{
			name: "PutFile error",
			putFile: PutFile{
				exp: PutFileExp{
					uri: "",
					err: expPutFileError,
				},
			},
			exp: Exp{
				uri: "",
				err: expPutFileError,
			},
		},
		{
			name: "not found",
			putFile: PutFile{
				exp: PutFileExp{
					uri: expUri,
					err: nil,
				},
			},
			updateUser: UpdateUser{
				exp: UpdateUserExp{
					err: repo.ErrNoRecord,
				},
			},
			exp: Exp{
				uri: "",
				err: app.ErrNotFound,
			},
		},
		{
			name: "UserUpdate error",
			putFile: PutFile{
				exp: PutFileExp{
					uri: expUri,
					err: nil,
				},
			},
			updateUser: UpdateUser{
				exp: UpdateUserExp{
					err: expUserUpdateError,
				},
			},
			exp: Exp{
				uri: "",
				err: expUserUpdateError,
			},
		},
		{
			name: "ok",
			putFile: PutFile{
				exp: PutFileExp{
					uri: expUri,
					err: nil,
				},
			},
			updateUser: UpdateUser{
				exp: UpdateUserExp{
					err: nil,
				},
			},
			exp: Exp{
				uri: expUri,
				err: nil,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require := require.New(t)

			controller := gomock.NewController(t)
			mockStorage := mock_storage.NewMockService(controller)
			mockRepo := mock_repo.NewMockServiceTx(controller)

			putFileCall := mockStorage.EXPECT().
				PutFile(ctx, avatar, options).
				Return(tc.putFile.exp.uri, tc.putFile.exp.err)

			if tc.putFile.exp.err == nil {
				mockRepo.EXPECT().
					UserUpdate(ctx, userID, map[string]any{
						models.UserColumns.Avatar: tc.putFile.exp.uri,
					}).
					Return(tc.updateUser.exp.err).
					After(putFileCall)
			}

			app := app.NewApplication(mockRepo, nil, nil, nil, mockStorage)

			uri, err := app.UserPutAvatar(ctx, userID, avatar, options)
			require.Equal(tc.exp.err, err)
			require.Equal(tc.exp.uri, uri)
		})
	}
}
