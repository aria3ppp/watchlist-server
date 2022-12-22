package app_test

import (
	"context"
	"errors"
	"testing"
	_ "unsafe"

	"github.com/aria3ppp/watchlist-server/internal/app"
	"github.com/aria3ppp/watchlist-server/internal/dto"
	"github.com/aria3ppp/watchlist-server/internal/hasher"
	"github.com/aria3ppp/watchlist-server/internal/hasher/mock_hasher"
	"github.com/aria3ppp/watchlist-server/internal/models"
	"github.com/aria3ppp/watchlist-server/internal/repo"
	"github.com/aria3ppp/watchlist-server/internal/repo/mock_repo"
	"github.com/aria3ppp/watchlist-server/internal/token"
	"github.com/aria3ppp/watchlist-server/internal/token/mock_token"
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

			app := app.NewApplication(mockRepo, nil, nil, nil)

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
		expHashed = "hashed"
		user      = &models.User{
			Email:          req.Email,
			HashedPassword: expHashed,
			FirstName:      req.FirstName,
			LastName:       req.LastName,
			Bio:            req.Bio,
			Birthdate:      req.Birthdate,
		}
		expNoRecordError             = repo.ErrNoRecord
		expEmailAlreadyUsedError     = app.ErrEmailAlreadyUsed
		expUserGetByEmailError       = errors.New("UserGetByEmail error")
		expGenerateFromPasswordError = errors.New("GenerateFromPassword error")
		expUserCreateError           = errors.New("UserCreate error")
	)

	type UserGetByEmailExp struct {
		user *models.User
		err  error
	}
	type HashPasswordExp struct {
		hashed []byte
		err    error
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
	type HashPassword struct {
		exp HashPasswordExp
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
		hashPassword   HashPassword
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
			hashPassword: HashPassword{
				exp: HashPasswordExp{
					hashed: nil,
					err:    expGenerateFromPasswordError,
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
			hashPassword: HashPassword{
				exp: HashPasswordExp{
					hashed: []byte(expHashed),
					err:    nil,
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
			hashPassword: HashPassword{
				exp: HashPasswordExp{
					hashed: []byte(expHashed),
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
				Transaction(ctx, gomock.Any()).
				Do(func(ctx context.Context, fn func(_ context.Context, _ repo.Service) error) {
					fn(ctx, mockRepo)
				}).
				Return(tc.tx.exp.err)

			userGetByEmailCall := mockRepo.EXPECT().
				UserGetByEmail(ctx, req.Email).
				Return(tc.userGetByEmail.exp.user, tc.userGetByEmail.exp.err).
				After(txCall)

			if tc.userGetByEmail.exp.err == repo.ErrNoRecord {
				hashCall := mockHasher.EXPECT().
					GenerateFromPassword([]byte(req.Password), gomock.Any()).
					Return(tc.hashPassword.exp.hashed, tc.hashPassword.exp.err).
					After(userGetByEmailCall)

				if tc.hashPassword.exp.err == nil {
					mockRepo.EXPECT().
						UserCreate(ctx, user).
						Do(func(_ context.Context, user *models.User) {
							user.ID = tc.exp.userID
						}).
						Return(tc.create.exp.err).
						After(hashCall)
				}
			}

			app := app.NewApplication(mockRepo, nil, nil, mockHasher)

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
			Email:          req.Email,
			HashedPassword: "hashed",
		}
		payload                        = &token.Payload{UserID: 1}
		expNoRecordError               = repo.ErrNoRecord
		expEmailNotFoundError          = app.ErrNotFound
		expIncorrectPassword           = app.ErrIncorrectPassword
		expUserGetByEmailError         = errors.New("UserGetByEmail error")
		expCompareHashAndPasswordError = errors.New(
			"CompareHashAndPassword error",
		)
		expAccessToken               = "access token"
		expRefreshToken              = "refresh token"
		expGenerateAccessTokenError  = errors.New("GenerateAccessToken error")
		expGenerateRefreshTokenError = errors.New("GenerateRefreshToken error")
	)

	type UserGetByEmailExp struct {
		user *models.User
		err  error
	}
	type GenerateTokenExp struct {
		token string
		err   error
	}
	type GenerateToken struct {
		exp GenerateTokenExp
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
	type Exp struct {
		accessToken, refreshToken string
		err                       error
	}
	type TestCase struct {
		name                 string
		userGetByEmail       UserGetByEmail
		compareHash          CompareHash
		generateAccessToken  GenerateToken
		generateRefreshToken GenerateToken
		exp                  Exp
	}

	testCases := []TestCase{
		{
			name: "email not found",
			userGetByEmail: UserGetByEmail{
				exp: UserGetByEmailExp{
					user: nil,
					err:  expNoRecordError,
				},
			},
			exp: Exp{
				accessToken:  "",
				refreshToken: "",
				err:          expEmailNotFoundError,
			},
		},

		{
			name: "UserGetByEmail error",
			userGetByEmail: UserGetByEmail{
				exp: UserGetByEmailExp{
					user: nil,
					err:  expUserGetByEmailError,
				},
			},
			exp: Exp{
				accessToken:  "",
				refreshToken: "",
				err:          expUserGetByEmailError,
			},
		},

		{
			name: "incorrect password",
			userGetByEmail: UserGetByEmail{
				exp: UserGetByEmailExp{
					user: expUser,
					err:  nil,
				},
			},
			compareHash: CompareHash{
				exp: CompareHashExp{
					err: hasher.ErrMismatchedHashAndPassword,
				},
			},
			exp: Exp{
				accessToken:  "",
				refreshToken: "",
				err:          expIncorrectPassword,
			},
		},

		{
			name: "CompareHashAndPassword error",
			userGetByEmail: UserGetByEmail{
				exp: UserGetByEmailExp{
					user: expUser,
					err:  nil,
				},
			},
			compareHash: CompareHash{
				exp: CompareHashExp{
					err: expCompareHashAndPasswordError,
				},
			},
			exp: Exp{
				accessToken:  "",
				refreshToken: "",
				err:          expCompareHashAndPasswordError,
			},
		},

		{
			name: "GenerateAccessToken error",
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
			generateAccessToken: GenerateToken{
				exp: GenerateTokenExp{
					token: "",
					err:   expGenerateAccessTokenError,
				},
			},
			exp: Exp{
				accessToken:  "",
				refreshToken: "",
				err:          expGenerateAccessTokenError,
			},
		},

		{
			name: "GenerateRefreshToken error",
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
			generateAccessToken: GenerateToken{
				exp: GenerateTokenExp{
					token: expAccessToken,
					err:   nil,
				},
			},
			generateRefreshToken: GenerateToken{
				exp: GenerateTokenExp{
					token: "",
					err:   expGenerateRefreshTokenError,
				},
			},
			exp: Exp{
				accessToken:  "",
				refreshToken: "",
				err:          expGenerateRefreshTokenError,
			},
		},

		{
			name: "ok",
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
			generateAccessToken: GenerateToken{
				exp: GenerateTokenExp{
					token: expAccessToken,
					err:   nil,
				},
			},
			generateRefreshToken: GenerateToken{
				exp: GenerateTokenExp{
					token: expRefreshToken,
					err:   nil,
				},
			},
			exp: Exp{
				accessToken:  expAccessToken,
				refreshToken: expRefreshToken,
				err:          nil,
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
			mockTokenService := mock_token.NewMockService(controller)

			userGetByEmailCall := mockRepo.EXPECT().
				UserGetByEmail(ctx, req.Email).
				Do(func(_ context.Context, _ string) {
					expUser.ID = payload.UserID
				}).
				Return(tc.userGetByEmail.exp.user, tc.userGetByEmail.exp.err)

			if tc.userGetByEmail.exp.err == nil {
				compateHashCall := mockHasher.EXPECT().
					CompareHashAndPassword([]byte(expUser.HashedPassword), []byte(req.Password)).
					Return(tc.compareHash.exp.err).
					After(userGetByEmailCall)

				if tc.compareHash.exp.err == nil {
					generateAccessTokenCall := mockTokenService.EXPECT().
						GenerateAccessToken(payload).
						Return(tc.generateAccessToken.exp.token, tc.generateAccessToken.exp.err).
						After(compateHashCall)

					if tc.generateAccessToken.exp.err == nil {
						mockTokenService.EXPECT().
							GenerateRefreshToken(payload).
							Return(tc.generateRefreshToken.exp.token, tc.generateRefreshToken.exp.err).
							After(generateAccessTokenCall)
					}
				}
			}

			app := app.NewApplication(
				mockRepo,
				mockTokenService,
				nil,
				mockHasher,
			)

			tAccess, tRefresh, err := app.UserLogin(ctx, req)
			require.Equal(tc.exp.err, err)
			require.Equal(tc.exp.accessToken, tAccess)
			require.Equal(tc.exp.refreshToken, tRefresh)
		})
	}
}

func TestUserRefreshToken(t *testing.T) {
	t.Parallel()

	var (
		ctx = context.Background()

		expPayload                  = &token.Payload{UserID: 1}
		expTokenInvalidError        = app.ErrTokenInvalid
		refreshToken                = "refresh token"
		expNewAccessToken           = "new access token"
		expValidateTokenError       = errors.New("ValidateToken error")
		expGenerateAccessTokenError = errors.New("GenerateAccessToken error")
	)

	type GenerateAccessTokenExp struct {
		token string
		err   error
	}
	type GenerateAccessToken struct {
		exp GenerateAccessTokenExp
	}
	type ValidateTokenExp struct {
		payload *token.Payload
		err     error
	}
	type ValidateToken struct {
		exp ValidateTokenExp
	}
	type Exp struct {
		accessToken string
		err         error
	}
	type TestCase struct {
		name                string
		validateToken       ValidateToken
		generateAccessToken GenerateAccessToken
		exp                 Exp
	}

	testCases := []TestCase{
		{
			name: "invalid token",
			validateToken: ValidateToken{
				exp: ValidateTokenExp{
					payload: nil,
					err:     token.ErrInvalidToken,
				},
			},
			exp: Exp{
				accessToken: "",
				err:         expTokenInvalidError,
			},
		},

		{
			name: "ValidateToken error",
			validateToken: ValidateToken{
				exp: ValidateTokenExp{
					payload: nil,
					err:     expValidateTokenError,
				},
			},
			exp: Exp{
				accessToken: "",
				err:         expValidateTokenError,
			},
		},

		{
			name: "GenerateAccessToken error",
			validateToken: ValidateToken{
				exp: ValidateTokenExp{
					payload: expPayload,
					err:     nil,
				},
			},
			generateAccessToken: GenerateAccessToken{
				exp: GenerateAccessTokenExp{
					token: "",
					err:   expGenerateAccessTokenError,
				},
			},
			exp: Exp{
				accessToken: "",
				err:         expGenerateAccessTokenError,
			},
		},

		{
			name: "ok",
			validateToken: ValidateToken{
				exp: ValidateTokenExp{
					payload: expPayload,
					err:     nil,
				},
			},
			generateAccessToken: GenerateAccessToken{
				exp: GenerateAccessTokenExp{
					token: expNewAccessToken,
					err:   nil,
				},
			},
			exp: Exp{
				accessToken: expNewAccessToken,
				err:         nil,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require := require.New(t)

			controller := gomock.NewController(t)
			mockTokenService := mock_token.NewMockService(controller)

			validateTokenCall := mockTokenService.EXPECT().
				ValidateToken(refreshToken).
				Return(tc.validateToken.exp.payload, tc.validateToken.exp.err)

			if tc.validateToken.exp.err == nil {
				mockTokenService.EXPECT().
					GenerateAccessToken(expPayload).
					Return(tc.generateAccessToken.exp.token, tc.generateAccessToken.exp.err).
					After(validateTokenCall)
			}

			app := app.NewApplication(nil, mockTokenService, nil, nil)

			accessToken, err := app.UserRefreshToken(ctx, refreshToken)
			require.Equal(tc.exp.err, err)
			require.Equal(tc.exp.accessToken, accessToken)
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

			app := app.NewApplication(mockRepo, nil, nil, nil)

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

			app := app.NewApplication(mockRepo, nil, nil, nil)

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
			ID:             userID,
			Email:          "email",
			HashedPassword: "hash",
		}
		columns = map[string]any{
			models.UserColumns.HashedPassword: expUser.HashedPassword,
		}
		expSameNewPasswordError        = app.ErrSameNewPassword
		expNotFoundError               = app.ErrNotFound
		expUserGetError                = errors.New("UserGet error")
		expIncorrectPasswordError      = app.ErrIncorrectPassword
		expCompareHashAndPasswordError = errors.New(
			"CompareHashAndPassword error",
		)
		expGenerateFromPasswordError = errors.New(
			"GenerateFromPassword error",
		)
		expUserUpdateError = errors.New("UserUpdate error")
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
	type GenerateFromPasswordExp struct {
		hashedPassword string
		err            error
	}
	type GenerateFromPassword struct {
		exp GenerateFromPasswordExp
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
		name                 string
		req                  *dto.UserPasswordUpdateRequest
		tx                   Tx
		userGet              UserGet
		compareHash          CompareHash
		generateFromPassword GenerateFromPassword
		userUpdate           UserUpdate
		exp                  Exp
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
					err: hasher.ErrMismatchedHashAndPassword,
				},
			},
			exp: Exp{
				err: expIncorrectPasswordError,
			},
		},

		{
			name: "CompareHashAndPassword error",
			req:  req,
			tx: Tx{
				exp: TxExp{
					err: expCompareHashAndPasswordError,
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
					err: expCompareHashAndPasswordError,
				},
			},
			exp: Exp{
				err: expCompareHashAndPasswordError,
			},
		},

		{
			name: "GenerateFromPassword error",
			req:  req,
			tx: Tx{
				exp: TxExp{
					err: expGenerateFromPasswordError,
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
			generateFromPassword: GenerateFromPassword{
				exp: GenerateFromPasswordExp{
					hashedPassword: "",
					err:            expGenerateFromPasswordError,
				},
			},
			exp: Exp{
				err: expGenerateFromPasswordError,
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
			generateFromPassword: GenerateFromPassword{
				exp: GenerateFromPasswordExp{
					hashedPassword: expUser.HashedPassword,
					err:            nil,
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
			generateFromPassword: GenerateFromPassword{
				exp: GenerateFromPasswordExp{
					hashedPassword: expUser.HashedPassword,
					err:            nil,
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
					Transaction(ctx, gomock.Any()).
					Do(func(ctx context.Context, fn func(_ context.Context, _ repo.Service) error) {
						fn(ctx, mockRepo)
					}).
					Return(tc.tx.exp.err)

				userGetCall := mockRepo.EXPECT().
					UserGet(ctx, userID).
					Return(tc.userGet.exp.user, tc.userGet.exp.err).
					After(txCall)

				if tc.userGet.exp.err == nil {
					compareHashAndPasswordCall := mockHasher.EXPECT().
						CompareHashAndPassword([]byte(expUser.HashedPassword), []byte(req.CurrentPassword)).
						Return(tc.compareHash.exp.err).
						After(userGetCall)

					if tc.compareHash.exp.err == nil {
						generateFromPasswordCall := mockHasher.EXPECT().
							GenerateFromPassword([]byte(req.NewPassword), gomock.Any()).
							Return([]byte(tc.generateFromPassword.exp.hashedPassword), tc.generateFromPassword.exp.err).
							After(compareHashAndPasswordCall)

						if tc.generateFromPassword.exp.err == nil {
							mockRepo.EXPECT().
								UserUpdate(ctx, userID, columns).
								Return(tc.userUpdate.exp.err).
								After(generateFromPasswordCall)
						}
					}
				}
			}

			app := app.NewApplication(mockRepo, nil, nil, mockHasher)

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
			ID:             userID,
			Email:          "email",
			HashedPassword: "hash",
		}
		expNotFoundError               = app.ErrNotFound
		expUserGetError                = errors.New("UserGet error")
		expIncorrectPasswordError      = app.ErrIncorrectPassword
		expCompareHashAndPasswordError = errors.New(
			"CompareHashAndPassword error",
		)
		expUserDeleteError = errors.New("UserDelete error")
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
					err: hasher.ErrMismatchedHashAndPassword,
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
					err: expCompareHashAndPasswordError,
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
					err: expCompareHashAndPasswordError,
				},
			},
			exp: Exp{
				err: expCompareHashAndPasswordError,
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
				Transaction(ctx, gomock.Any()).
				Do(func(ctx context.Context, fn func(_ context.Context, _ repo.Service) error) {
					fn(ctx, mockRepo)
				}).
				Return(tc.tx.exp.err)

			userGetCall := mockRepo.EXPECT().
				UserGet(ctx, userID).
				Return(tc.userGet.exp.user, tc.userGet.exp.err).
				After(txCall)

			if tc.userGet.exp.err == nil {
				compareHashAndPasswordCall := mockHasher.EXPECT().
					CompareHashAndPassword([]byte(expUser.HashedPassword), []byte(req.Password)).
					Return(tc.compareHash.exp.err).
					After(userGetCall)

				if tc.compareHash.exp.err == nil {
					mockRepo.EXPECT().
						UserDelete(ctx, userID).
						Return(tc.userDelete.exp.err).
						After(compareHashAndPasswordCall)
				}
			}

			app := app.NewApplication(mockRepo, nil, nil, mockHasher)

			err := app.UserDelete(ctx, userID, req)
			require.Equal(tc.exp.err, err)
		})
	}
}
