package server_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/aria3ppp/watchlist-server/internal/app"
	"github.com/aria3ppp/watchlist-server/internal/config"
	"github.com/aria3ppp/watchlist-server/internal/dto"
	"github.com/aria3ppp/watchlist-server/internal/models"
	"github.com/aria3ppp/watchlist-server/internal/server/response"
	"github.com/aria3ppp/watchlist-server/internal/testutils"
	"github.com/aria3ppp/watchlist-server/internal/validator"
	"github.com/gavv/httpexpect/v2"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/null/v8"
)

func TestHandleUserGet(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	server, appInstance, defaults, teardown := setup(OptEnableDefaultUser)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/authorized/user/{id}"
	method := http.MethodGet

	// invalid id
	e.Request(method, path).
		WithPath("id", -1).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		Equal(response.Error(response.StatusInvalidURLParameter))

	// user not found
	e.Request(method, path).
		WithPath("id", 999).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusNotFound).
		JSON().
		Object().
		Equal(response.Error(response.StatusNotFound))

	// create a new user
	userCreateReq := &dto.UserCreateRequest{
		Email:    "new_email@example.com",
		Password: "new_pa$$W0RD1",
	}
	userID, err := appInstance.UserCreate(ctx, userCreateReq)
	require.NoError(err)

	gotUser, err := appInstance.UserGet(ctx, userID)
	require.NoError(err)

	payload := &models.User{
		ID:        userID,
		Email:     userCreateReq.Email,
		FirstName: userCreateReq.FirstName,
		LastName:  userCreateReq.LastName,
		Bio:       userCreateReq.Bio,
		Birthdate: userCreateReq.Birthdate,
		Joindate:  gotUser.Joindate,
	}

	// get user
	e.Request(method, path).
		WithPath("id", userID).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Equal(response.OK(payload))
}

func TestHandleUserCreate(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	server, appInstance, _, teardown := setup()
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/user/"
	method := http.MethodPost

	// create user
	userCreateReq := &dto.UserCreateRequest{
		Email:    "aria3ppp@gamil.com",
		Password: "pa$$W0RD1",
	}

	// As database date is in utc so it should compare in the same timezone
	createDate := testutils.Date(time.Now().UTC().Date())

	userIDRaw := e.Request(method, path).
		WithJSON(userCreateReq).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		ValueEqual("status", response.StatusOK.String()).
		Value("payload").Number().Gt(0).Raw()

	userID := int(userIDRaw)

	// check user created
	gotUser, err := appInstance.UserGet(ctx, userID)
	require.NoError(err)

	require.GreaterOrEqual(gotUser.Joindate, createDate)

	require.Equal(&models.User{
		ID:             userID,
		Email:          userCreateReq.Email,
		HashedPassword: gotUser.HashedPassword,
		FirstName:      userCreateReq.FirstName,
		LastName:       userCreateReq.LastName,
		Bio:            userCreateReq.Bio,
		Birthdate:      userCreateReq.Birthdate,
		Joindate:       gotUser.Joindate,
	}, gotUser)

	// email address already taken
	e.Request(method, path).
		WithJSON(dto.UserCreateRequest{
			Email:    userCreateReq.Email,
			Password: userCreateReq.Password,
		}).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		Equal(response.Error(response.StatusEmailAlreadyUsed))
}

func TestHandleUserCreate_ValidateRequest(t *testing.T) {
	server, _, _, teardown := setup()
	t.Cleanup(teardown)

	path := "/v1/user/"
	method := http.MethodPost

	timeNow := time.Now()

	type Length struct {
		shorterThanMin string
		longerThanMax  string
	}
	testDatas := struct {
		email struct {
			length       Length
			validValue   string
			invalidValue string
		}
		password struct {
			length       Length
			validValue   string
			invalidValue string
		}
		firstname struct {
			length Length
		}
		lastname struct {
			length Length
		}
		bio struct {
			length Length
		}
		birthdate struct {
			lesserThanMinValue  time.Time
			greaterThanMaxValue time.Time
		}
	}{
		email: struct {
			length       Length
			validValue   string
			invalidValue string
		}{
			length: Length{
				shorterThanMin: "email",
				longerThanMax: testutils.GenerateStringLongerThanMaxLength(
					config.Config.Validation.User.Email.MaxLength,
				),
			},
			validValue:   "email@example.com",
			invalidValue: "invalidemailaddr",
		},
		password: struct {
			length       Length
			validValue   string
			invalidValue string
		}{
			length: Length{
				shorterThanMin: "passwd",
				longerThanMax: testutils.GenerateStringLongerThanMaxLength(
					config.Config.Validation.User.Password.MaxLength,
				),
			},
			validValue:   "pa$$W0RD0",
			invalidValue: "password",
		},
		firstname: struct{ length Length }{
			length: Length{
				shorterThanMin: "f",
				longerThanMax: testutils.GenerateStringLongerThanMaxLength(
					config.Config.Validation.User.FirstName.MaxLength,
				),
			},
		},
		lastname: struct{ length Length }{
			length: Length{
				shorterThanMin: "l",
				longerThanMax: testutils.GenerateStringLongerThanMaxLength(
					config.Config.Validation.User.LastName.MaxLength,
				),
			},
		},
		bio: struct{ length Length }{
			length: Length{
				shorterThanMin: "b",
				longerThanMax: testutils.GenerateStringLongerThanMaxLength(
					config.Config.Validation.User.Bio.MaxLength,
				),
			},
		},
		birthdate: struct {
			lesserThanMinValue  time.Time
			greaterThanMaxValue time.Time
		}{
			lesserThanMinValue: testutils.Date(
				config.Config.Validation.User.Birthdate.MinValue.Year-1,
				time.Month(
					config.Config.Validation.User.Birthdate.MinValue.Month,
				),
				config.Config.Validation.User.Birthdate.MinValue.Day,
			),
			greaterThanMaxValue: testutils.Date(
				timeNow.Year(), timeNow.Month(), timeNow.Day(),
			).Add(time.Hour),
		},
	}

	testCases := []struct {
		name      string
		req       dto.UserCreateRequest
		expErrors validation.Errors
	}{
		{
			name: "tc1",
			req:  dto.UserCreateRequest{},
			expErrors: validation.Errors{
				"email":    validation.ErrRequired,
				"password": validation.ErrRequired,
			},
		},

		{
			name: "tc2",
			req: dto.UserCreateRequest{
				Email:     "",
				Password:  "",
				FirstName: null.StringFrom(""),
				LastName:  null.StringFrom(""),
				Bio:       null.StringFrom(""),
				Birthdate: null.TimeFrom(time.Time{}),
			},
			// reauired if submitted (null.Valid == true)
			expErrors: validation.Errors{
				"email":      validation.ErrRequired,
				"password":   validation.ErrRequired,
				"first_name": validation.ErrRequired,
				"last_name":  validation.ErrRequired,
				"bio":        validation.ErrRequired,
				"birthdate":  validation.ErrRequired,
			},
		},

		{
			name: "tc3",
			req: dto.UserCreateRequest{
				Email:    testDatas.email.length.shorterThanMin,
				Password: testDatas.password.length.shorterThanMin,
			},
			expErrors: validation.Errors{
				"email": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.Email.MinLength,
						"max": config.Config.Validation.User.Email.MaxLength,
					},
				),
				"password": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.Password.MinLength,
						"max": config.Config.Validation.User.Password.MaxLength,
					},
				),
			},
		},

		{
			name: "tc4",
			req: dto.UserCreateRequest{
				Email:    testDatas.email.length.longerThanMax,
				Password: testDatas.password.length.longerThanMax,
			},
			expErrors: validation.Errors{
				"email": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.Email.MinLength,
						"max": config.Config.Validation.User.Email.MaxLength,
					},
				),
				"password": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.Password.MinLength,
						"max": config.Config.Validation.User.Password.MaxLength,
					},
				),
			},
		},

		{
			name: "tc5",
			req: dto.UserCreateRequest{
				Email:    testDatas.email.invalidValue,
				Password: testDatas.password.invalidValue,
			},
			expErrors: validation.Errors{
				"email": is.ErrEmail,
				"password": validator.ErrPasswordInvalid.SetParams(
					map[string]any{
						"num":     config.Config.Validation.User.Password.RequiredNumbers,
						"lower":   config.Config.Validation.User.Password.RequiredLowerLetters,
						"upper":   config.Config.Validation.User.Password.RequiredUpperLetters,
						"special": config.Config.Validation.User.Password.RequiredSpecialChars,
					},
				),
			},
		},

		{
			name: "tc6",
			req: dto.UserCreateRequest{
				Email:    testDatas.email.validValue,
				Password: testDatas.password.validValue,
				FirstName: null.StringFrom(
					testDatas.firstname.length.shorterThanMin,
				),
				LastName: null.StringFrom(
					testDatas.lastname.length.shorterThanMin,
				),
				Bio: null.StringFrom(testDatas.bio.length.shorterThanMin),
				Birthdate: null.TimeFrom(
					testDatas.birthdate.lesserThanMinValue,
				),
			},
			expErrors: validation.Errors{
				"first_name": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.FirstName.MinLength,
						"max": config.Config.Validation.User.FirstName.MaxLength,
					},
				),
				"last_name": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.LastName.MinLength,
						"max": config.Config.Validation.User.LastName.MaxLength,
					},
				),
				"bio": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.Bio.MinLength,
						"max": config.Config.Validation.User.Bio.MaxLength,
					},
				),
				"birthdate": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{
						"threshold": testutils.Date(
							config.Config.Validation.User.Birthdate.MinValue.Year,
							time.Month(
								config.Config.Validation.User.Birthdate.MinValue.Month,
							),
							config.Config.Validation.User.Birthdate.MinValue.Day,
						),
					},
				),
			},
		},

		{
			name: "tc7",
			req: dto.UserCreateRequest{
				Email:    testDatas.email.validValue,
				Password: testDatas.password.validValue,
				FirstName: null.StringFrom(
					testDatas.firstname.length.longerThanMax,
				),
				LastName: null.StringFrom(
					testDatas.lastname.length.longerThanMax,
				),
				Bio: null.StringFrom(testDatas.bio.length.longerThanMax),
				Birthdate: null.TimeFrom(
					testDatas.birthdate.greaterThanMaxValue,
				),
			},
			expErrors: validation.Errors{
				"first_name": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.FirstName.MinLength,
						"max": config.Config.Validation.User.FirstName.MaxLength,
					},
				),
				"last_name": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.LastName.MinLength,
						"max": config.Config.Validation.User.LastName.MaxLength,
					},
				),
				"bio": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.Bio.MinLength,
						"max": config.Config.Validation.User.Bio.MaxLength,
					},
				),
				"birthdate": validation.ErrMaxLessEqualThanRequired.SetParams(
					map[string]any{
						"threshold": testutils.Date(
							time.Now().Year(),
							time.Now().Month(),
							time.Now().Day(),
						),
					},
				),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			e := httpexpect.New(t, server.URL)

			e.Request(method, path).
				WithJSON(tc.req).
				Expect().
				Status(http.StatusBadRequest).
				JSON().
				Equal(response.Error(
					response.StatusInvalidRequest,
					tc.expErrors.Error(),
				))
		})
	}
}

func TestHandleUserUpdate(t *testing.T) {
	require := require.New(t)

	server, appInstance, defaults, teardown := setup(OptEnableDefaultUser)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/authorized/user/"
	method := http.MethodPatch

	timeNow := time.Now()

	// update user
	userUpdateReq := &dto.UserUpdateRequest{
		FirstName: null.StringFrom("updated_first_name"),
		LastName:  null.StringFrom("update_last_name"),
		Bio:       null.StringFrom("update_bio"),
		Birthdate: null.TimeFrom(
			testutils.Date(
				timeNow.Year(),
				timeNow.Month(),
				timeNow.Day(),
			),
		),
	}

	e.Request(method, path).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(userUpdateReq).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Equal(response.OK(nil))

	// check updated fileds
	updatedUser, err := appInstance.UserGet(
		context.Background(),
		defaults.user.id,
	)
	require.NoError(err)
	if userUpdateReq.FirstName.Valid {
		require.Equal(userUpdateReq.FirstName, updatedUser.FirstName)
	}
	if userUpdateReq.LastName.Valid {
		require.Equal(userUpdateReq.LastName, updatedUser.LastName)
	}
	if userUpdateReq.Bio.Valid {
		require.Equal(userUpdateReq.Bio, updatedUser.Bio)
	}
	if userUpdateReq.Birthdate.Valid {
		testutils.SetTimeLocation(
			&userUpdateReq.Birthdate.Time,
			updatedUser.Birthdate.Time.Location(),
		)
		require.Equal(userUpdateReq.Birthdate, updatedUser.Birthdate)
	}

	// user not found
	err = appInstance.UserDelete(
		context.Background(),
		defaults.user.id,
		&dto.UserDeleteRequest{
			Password: defaults.user.password,
		},
	)
	require.NoError(err)

	e.Request(method, path).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(&dto.UserUpdateRequest{}).
		Expect().
		Status(http.StatusNotFound).
		JSON().
		Object().
		Equal(response.Error(response.StatusNotFound))
}

func TestHandleUserUpdate_ValidateRequest(t *testing.T) {
	server, _, defaults, teardown := setup(OptEnableDefaultUser)
	t.Cleanup(teardown)

	path := "/v1/authorized/user/"
	method := http.MethodPatch

	timeNow := time.Now()

	type Length struct {
		shorterThanMin string
		longerThanMax  string
	}
	testDatas := struct {
		firstname struct {
			length Length
		}
		lastname struct {
			length Length
		}
		bio struct {
			length Length
		}
		birthdate struct {
			lesserThanMinValue  time.Time
			greaterThanMaxValue time.Time
		}
	}{
		firstname: struct{ length Length }{
			length: Length{
				shorterThanMin: "f",
				longerThanMax: testutils.GenerateStringLongerThanMaxLength(
					config.Config.Validation.User.FirstName.MaxLength,
				),
			},
		},
		lastname: struct{ length Length }{
			length: Length{
				shorterThanMin: "l",
				longerThanMax: testutils.GenerateStringLongerThanMaxLength(
					config.Config.Validation.User.LastName.MaxLength,
				),
			},
		},
		bio: struct{ length Length }{
			length: Length{
				shorterThanMin: "b",
				longerThanMax: testutils.GenerateStringLongerThanMaxLength(
					config.Config.Validation.User.Bio.MaxLength,
				),
			},
		},
		birthdate: struct {
			lesserThanMinValue  time.Time
			greaterThanMaxValue time.Time
		}{
			lesserThanMinValue: testutils.Date(
				config.Config.Validation.User.Birthdate.MinValue.Year-1,
				time.Month(
					config.Config.Validation.User.Birthdate.MinValue.Month,
				),
				config.Config.Validation.User.Birthdate.MinValue.Day,
			),
			greaterThanMaxValue: testutils.Date(
				timeNow.Year(), timeNow.Month(), timeNow.Day(),
			).Add(time.Hour),
		},
	}

	testCases := []struct {
		name      string
		req       dto.UserUpdateRequest
		expErrors validation.Errors
	}{
		{
			name: "tc1",
			req: dto.UserUpdateRequest{
				FirstName: null.StringFrom(""),
				LastName:  null.StringFrom(""),
				Bio:       null.StringFrom(""),
				Birthdate: null.TimeFrom(time.Time{}),
			},
			// reauired if submitted (null.Valid == true)
			expErrors: validation.Errors{
				"first_name": validation.ErrRequired,
				"last_name":  validation.ErrRequired,
				"bio":        validation.ErrRequired,
				"birthdate":  validation.ErrRequired,
			},
		},

		{
			name: "tc2",
			req: dto.UserUpdateRequest{
				FirstName: null.StringFrom(
					testDatas.firstname.length.shorterThanMin,
				),
				LastName: null.StringFrom(
					testDatas.lastname.length.shorterThanMin,
				),
				Bio: null.StringFrom(testDatas.bio.length.shorterThanMin),
				Birthdate: null.TimeFrom(
					testDatas.birthdate.lesserThanMinValue,
				),
			},
			expErrors: validation.Errors{
				"first_name": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.FirstName.MinLength,
						"max": config.Config.Validation.User.FirstName.MaxLength,
					},
				),
				"last_name": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.LastName.MinLength,
						"max": config.Config.Validation.User.LastName.MaxLength,
					},
				),
				"bio": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.Bio.MinLength,
						"max": config.Config.Validation.User.Bio.MaxLength,
					},
				),
				"birthdate": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{
						"threshold": testutils.Date(
							config.Config.Validation.User.Birthdate.MinValue.Year,
							time.Month(
								config.Config.Validation.User.Birthdate.MinValue.Month,
							),
							config.Config.Validation.User.Birthdate.MinValue.Day,
						),
					},
				),
			},
		},

		{
			name: "tc3",
			req: dto.UserUpdateRequest{
				FirstName: null.StringFrom(
					testDatas.firstname.length.longerThanMax,
				),
				LastName: null.StringFrom(
					testDatas.lastname.length.longerThanMax,
				),
				Bio: null.StringFrom(testDatas.bio.length.longerThanMax),
				Birthdate: null.TimeFrom(
					testDatas.birthdate.greaterThanMaxValue,
				),
			},
			expErrors: validation.Errors{
				"first_name": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.FirstName.MinLength,
						"max": config.Config.Validation.User.FirstName.MaxLength,
					},
				),
				"last_name": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.LastName.MinLength,
						"max": config.Config.Validation.User.LastName.MaxLength,
					},
				),
				"bio": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.Bio.MinLength,
						"max": config.Config.Validation.User.Bio.MaxLength,
					},
				),
				"birthdate": validation.ErrMaxLessEqualThanRequired.SetParams(
					map[string]any{
						"threshold": testutils.Date(
							time.Now().Year(),
							time.Now().Month(),
							time.Now().Day(),
						),
					},
				),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			e := httpexpect.New(t, server.URL)

			e.Request(method, path).
				WithHeader(echo.HeaderAuthorization, defaults.user.auth).
				WithJSON(tc.req).
				Expect().
				Status(http.StatusBadRequest).
				JSON().
				Equal(response.Error(
					response.StatusInvalidRequest,
					tc.expErrors.Error(),
				))
		})
	}
}

func TestHandleUserDelete(t *testing.T) {
	require := require.New(t)

	server, appInstance, defaults, teardown := setup(OptEnableDefaultUser)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/authorized/user/"
	method := http.MethodDelete

	// incorrect password
	e.Request(method, path).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(&dto.UserDeleteRequest{Password: "1nc0RR3ct_pa$$"}).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		Equal(response.Error(response.StatusIncorrectPassword))

	// delete user
	e.Request(method, path).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(&dto.UserDeleteRequest{Password: defaults.user.password}).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Equal(response.OK(nil))

	// check deleted
	userAfterDelete, err := appInstance.UserGet(
		context.Background(),
		defaults.user.id,
	)
	require.Nil(userAfterDelete)
	require.Equal(app.ErrNotFound, err)

	// user not found
	e.Request(method, path).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(&dto.UserDeleteRequest{Password: defaults.user.password}).
		Expect().
		Status(http.StatusNotFound).
		JSON().
		Object().
		Equal(response.Error(response.StatusNotFound))
}

func TestHandleUserDelete_ValidateRequest(t *testing.T) {
	server, _, defaults, teardown := setup(OptEnableDefaultUser)
	t.Cleanup(teardown)

	path := "/v1/authorized/user/"
	method := http.MethodDelete

	testCases := []struct {
		name      string
		req       dto.UserDeleteRequest
		expErrors validation.Errors
	}{
		{
			name: "tc1",
			req:  dto.UserDeleteRequest{},
			expErrors: validation.Errors{
				"password": validation.ErrRequired,
			},
		},
		{
			name: "tc2",
			req: dto.UserDeleteRequest{
				Password: "passwd",
			},
			expErrors: validation.Errors{
				"password": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.Password.MinLength,
						"max": config.Config.Validation.User.Password.MaxLength,
					},
				),
			},
		},

		{
			name: "tc3",
			req: dto.UserDeleteRequest{
				Password: testutils.GenerateStringLongerThanMaxLength(
					config.Config.Validation.User.Password.MaxLength,
				),
			},
			expErrors: validation.Errors{
				"password": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.Password.MinLength,
						"max": config.Config.Validation.User.Password.MaxLength,
					},
				),
			},
		},

		{
			name: "tc4",
			req: dto.UserDeleteRequest{
				Password: "invalid_password",
			},
			expErrors: map[string]error{
				"password": validator.ErrPasswordInvalid.SetParams(
					map[string]any{
						"num":     config.Config.Validation.User.Password.RequiredNumbers,
						"lower":   config.Config.Validation.User.Password.RequiredLowerLetters,
						"upper":   config.Config.Validation.User.Password.RequiredUpperLetters,
						"special": config.Config.Validation.User.Password.RequiredSpecialChars,
					},
				),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			e := httpexpect.New(t, server.URL)

			e.Request(method, path).
				WithHeader(echo.HeaderAuthorization, defaults.user.auth).
				WithJSON(tc.req).
				Expect().
				Status(http.StatusBadRequest).
				JSON().
				Equal(response.Error(
					response.StatusInvalidRequest,
					tc.expErrors.Error(),
				))
		})
	}
}

func TestHandleUserEmailUpdate(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	server, appInstance, defaults, teardown := setup(OptEnableDefaultUser)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/authorized/user/email"
	method := http.MethodPut

	// update email
	userEmailUpdateReq := &dto.UserEmailUpdateRequest{Email: "email@gmail.com"}

	e.Request(method, path).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(userEmailUpdateReq).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Equal(response.OK(nil))

	// check updated email
	gotUser, err := appInstance.UserGet(ctx, defaults.user.id)
	require.NoError(err)
	require.Equal(
		&models.User{
			ID:             defaults.user.id,
			Email:          userEmailUpdateReq.Email,
			HashedPassword: gotUser.HashedPassword,
			FirstName:      defaults.user.reqObject.FirstName,
			LastName:       defaults.user.reqObject.LastName,
			Bio:            defaults.user.reqObject.Bio,
			Birthdate:      defaults.user.reqObject.Birthdate,
			Joindate:       gotUser.Joindate,
		},
		gotUser,
	)

	// user not found
	err = appInstance.UserDelete(
		context.Background(),
		defaults.user.id,
		&dto.UserDeleteRequest{Password: defaults.user.password},
	)
	require.NoError(err)

	e.Request(method, path).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(userEmailUpdateReq).
		Expect().
		Status(http.StatusNotFound).
		JSON().
		Object().
		Equal(response.Error(response.StatusNotFound))
}

func TestHandleUserEmailUpdate_ValidateRequest(t *testing.T) {
	server, _, defaults, teardown := setup(OptEnableDefaultUser)
	t.Cleanup(teardown)

	path := "/v1/authorized/user/email"
	method := http.MethodPut

	testCases := []struct {
		name      string
		req       dto.UserEmailUpdateRequest
		expErrors validation.Errors
	}{
		{
			name: "tc1",
			req:  dto.UserEmailUpdateRequest{},
			expErrors: validation.Errors{
				"email": validation.ErrRequired,
			},
		},

		{
			name: "tc2",
			req: dto.UserEmailUpdateRequest{
				Email: "e",
			},
			expErrors: validation.Errors{
				"email": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.Email.MinLength,
						"max": config.Config.Validation.User.Email.MaxLength,
					},
				),
			},
		},

		{
			name: "tc3",
			req: dto.UserEmailUpdateRequest{
				Email: testutils.GenerateStringLongerThanMaxLength(
					config.Config.Validation.User.Email.MaxLength,
				),
			},
			expErrors: validation.Errors{
				"email": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.Email.MinLength,
						"max": config.Config.Validation.User.Email.MaxLength,
					},
				),
			},
		},

		{
			name: "tc4",
			req: dto.UserEmailUpdateRequest{
				Email: "invalidemailaddr",
			},
			expErrors: validation.Errors{
				"email": is.ErrEmail,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			e := httpexpect.New(t, server.URL)

			e.Request(method, path).
				WithHeader(echo.HeaderAuthorization, defaults.user.auth).
				WithJSON(tc.req).
				Expect().
				Status(http.StatusBadRequest).
				JSON().
				Object().
				Equal(response.Error(
					response.StatusInvalidRequest,
					tc.expErrors.Error(),
				))
		})
	}
}

func TestHandleUserPasswordUpdate(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	server, appInstance, defaults, teardown := setup(OptEnableDefaultUser)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/authorized/user/password"
	method := http.MethodPut

	newPassword := "new_pa$$W0RD1"

	// same password as before
	e.Request(method, path).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(&dto.UserPasswordUpdateRequest{
			CurrentPassword: defaults.user.password,
			NewPassword:     defaults.user.password,
		}).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		Equal(response.Error(response.StatusSameNewPassword))

	// incorrect password
	e.Request(method, path).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(&dto.UserPasswordUpdateRequest{
			CurrentPassword: "1nc0RR3ct_pa$$",
			NewPassword:     newPassword,
		}).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		Equal(response.Error(response.StatusIncorrectPassword))

	// update password
	e.Request(method, path).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(&dto.UserPasswordUpdateRequest{
			CurrentPassword: defaults.user.password,
			NewPassword:     newPassword,
		}).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Equal(response.OK(nil))

	// check password updated
	gotUser, err := appInstance.UserGet(ctx, defaults.user.id)
	require.NoError(err)
	require.Equal(
		&models.User{
			ID:             defaults.user.id,
			Email:          defaults.user.email,
			HashedPassword: gotUser.HashedPassword,
			FirstName:      defaults.user.reqObject.FirstName,
			LastName:       defaults.user.reqObject.LastName,
			Bio:            defaults.user.reqObject.Bio,
			Birthdate:      defaults.user.reqObject.Birthdate,
			Joindate:       gotUser.Joindate,
		},
		gotUser,
	)

	// user not found
	err = appInstance.UserDelete(
		context.Background(),
		defaults.user.id,
		&dto.UserDeleteRequest{Password: newPassword},
	)
	require.NoError(err)

	e.Request(method, path).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(&dto.UserPasswordUpdateRequest{
			CurrentPassword: defaults.user.password,
			NewPassword:     newPassword,
		}).
		Expect().
		Status(http.StatusNotFound).
		JSON().
		Object().
		Equal(response.Error(response.StatusNotFound))
}

func TestHandleUserPasswordUpdate_ValidateRequest(t *testing.T) {
	server, _, defaults, teardown := setup(OptEnableDefaultUser)
	t.Cleanup(teardown)

	path := "/v1/authorized/user/password"
	method := http.MethodPut

	testCases := []struct {
		name      string
		req       dto.UserPasswordUpdateRequest
		expErrors validation.Errors
	}{
		{
			name: "tc1",
			req:  dto.UserPasswordUpdateRequest{},
			expErrors: validation.Errors{
				"current_password": validation.ErrRequired,
				"new_password":     validation.ErrRequired,
			},
		},

		{
			name: "tc2",
			req: dto.UserPasswordUpdateRequest{
				CurrentPassword: "cpwd",
				NewPassword:     "npwd",
			},
			expErrors: validation.Errors{
				"current_password": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.Password.MinLength,
						"max": config.Config.Validation.User.Password.MaxLength,
					},
				),
				"new_password": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.Password.MinLength,
						"max": config.Config.Validation.User.Password.MaxLength,
					},
				),
			},
		},

		{
			name: "tc3",
			req: dto.UserPasswordUpdateRequest{
				CurrentPassword: testutils.GenerateStringLongerThanMaxLength(
					config.Config.Validation.User.Password.MaxLength,
				),
				NewPassword: testutils.GenerateStringLongerThanMaxLength(
					config.Config.Validation.User.Password.MaxLength,
				),
			},
			expErrors: validation.Errors{
				"current_password": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.Password.MinLength,
						"max": config.Config.Validation.User.Password.MaxLength,
					},
				),
				"new_password": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.Password.MinLength,
						"max": config.Config.Validation.User.Password.MaxLength,
					},
				),
			},
		},

		{
			name: "tc4",
			req: dto.UserPasswordUpdateRequest{
				CurrentPassword: "current_invalid_password",
				NewPassword:     "new_invalid_password",
			},
			expErrors: validation.Errors{
				"current_password": validator.ErrPasswordInvalid.SetParams(
					map[string]any{
						"num":     config.Config.Validation.User.Password.RequiredNumbers,
						"lower":   config.Config.Validation.User.Password.RequiredLowerLetters,
						"upper":   config.Config.Validation.User.Password.RequiredUpperLetters,
						"special": config.Config.Validation.User.Password.RequiredSpecialChars,
					},
				),
				"new_password": validator.ErrPasswordInvalid.SetParams(
					map[string]any{
						"num":     config.Config.Validation.User.Password.RequiredNumbers,
						"lower":   config.Config.Validation.User.Password.RequiredLowerLetters,
						"upper":   config.Config.Validation.User.Password.RequiredUpperLetters,
						"special": config.Config.Validation.User.Password.RequiredSpecialChars,
					},
				),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			e := httpexpect.New(t, server.URL)

			e.Request(method, path).
				WithHeader(echo.HeaderAuthorization, defaults.user.auth).
				WithJSON(tc.req).
				Expect().
				Status(http.StatusBadRequest).
				JSON().
				Object().
				Equal(response.Error(response.StatusInvalidRequest, tc.expErrors.Error()))
		})
	}
}

func TestHandleUserLogin(t *testing.T) {
	server, _, defaults, teardown := setup(OptEnableDefaultUser)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/user/login"
	method := http.MethodPost

	// email not found
	e.Request(method, path).
		WithJSON(dto.UserLoginRequest{
			Email:    "email@notfound.com",
			Password: defaults.user.password,
		}).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		Equal(response.Error(response.StatusEmailNotFound))

	// incorrect password
	e.Request(method, path).
		WithJSON(dto.UserLoginRequest{
			Email:    defaults.user.email,
			Password: "1nc0RR3ct_pa$$",
		}).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		Equal(response.Error(response.StatusIncorrectPassword))

	// login user
	payloadObj := e.Request(method, path).
		WithJSON(dto.UserCreateRequest{
			Email:    defaults.user.email,
			Password: defaults.user.password,
		}).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		ValueEqual("status", response.StatusOK.String()).
		Value("payload").
		Object()

	payloadObj.Value("access_token").String().NotEmpty()
	payloadObj.Value("refresh_token").String().NotEmpty()
}

func TestHandleUserLogin_ValidateRequest(t *testing.T) {
	server, _, _, teardown := setup()
	t.Cleanup(teardown)

	path := "/v1/user/login"
	method := http.MethodPost

	type Length struct {
		shorterThanMin string
		longerThanMax  string
	}
	testDatas := struct {
		email struct {
			length       Length
			validValue   string
			invalidValue string
		}
		password struct {
			length       Length
			validValue   string
			invalidValue string
		}
	}{
		email: struct {
			length       Length
			validValue   string
			invalidValue string
		}{
			length: Length{
				shorterThanMin: "email",
				longerThanMax: testutils.GenerateStringLongerThanMaxLength(
					config.Config.Validation.User.Email.MaxLength,
				),
			},
			validValue:   "email@example.com",
			invalidValue: "invalidemailaddr",
		},
		password: struct {
			length       Length
			validValue   string
			invalidValue string
		}{
			length: Length{
				shorterThanMin: "passwd",
				longerThanMax: testutils.GenerateStringLongerThanMaxLength(
					config.Config.Validation.User.Password.MaxLength,
				),
			},
			validValue:   "pa$$W0RD0",
			invalidValue: "password",
		},
	}

	testCases := []struct {
		name      string
		req       dto.UserLoginRequest
		expErrors validation.Errors
	}{
		{
			name: "tc1",
			req:  dto.UserLoginRequest{},
			expErrors: validation.Errors{
				"email":    validation.ErrRequired,
				"password": validation.ErrRequired,
			},
		},

		{
			name: "tc2",
			req: dto.UserLoginRequest{
				Email:    testDatas.email.length.shorterThanMin,
				Password: testDatas.password.length.shorterThanMin,
			},
			expErrors: validation.Errors{
				"email": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.Email.MinLength,
						"max": config.Config.Validation.User.Email.MaxLength,
					},
				),
				"password": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.Password.MinLength,
						"max": config.Config.Validation.User.Password.MaxLength,
					},
				),
			},
		},

		{
			name: "tc3",
			req: dto.UserLoginRequest{
				Email:    testDatas.email.length.longerThanMax,
				Password: testDatas.password.length.longerThanMax,
			},
			expErrors: validation.Errors{
				"email": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.Email.MinLength,
						"max": config.Config.Validation.User.Email.MaxLength,
					},
				),
				"password": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.Password.MinLength,
						"max": config.Config.Validation.User.Password.MaxLength,
					},
				),
			},
		},

		{
			name: "tc4",
			req: dto.UserLoginRequest{
				Email:    testDatas.email.invalidValue,
				Password: testDatas.password.invalidValue,
			},
			expErrors: validation.Errors{
				"email": errors.New(is.ErrEmail.Message()),
				"password": validator.ErrPasswordInvalid.SetParams(
					map[string]any{
						"num":     config.Config.Validation.User.Password.RequiredNumbers,
						"lower":   config.Config.Validation.User.Password.RequiredLowerLetters,
						"upper":   config.Config.Validation.User.Password.RequiredUpperLetters,
						"special": config.Config.Validation.User.Password.RequiredSpecialChars,
					},
				),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			e := httpexpect.New(t, server.URL)

			e.Request(method, path).
				WithJSON(tc.req).
				Expect().
				Status(http.StatusBadRequest).
				JSON().
				Object().
				Equal(response.Error(response.StatusInvalidRequest, tc.expErrors.Error()))
		})
	}
}

func TestHandleUserRefreshToken(t *testing.T) {
	server, _, defaults, teardown := setup(OptEnableDefaultUser)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/user/refresh"
	method := http.MethodGet

	// missing token
	e.Request(method, path).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		Equal(response.Error(response.StatusTokenMissingOrMalformed))

	// invalid token
	e.Request(method, path).
		WithHeader(echo.HeaderAuthorization, "Bearer invalid_token").
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		Equal(response.Error(response.StatusTokenInvalid))

	// refresh token
	e.Request(method, path).
		WithHeader(echo.HeaderAuthorization, defaults.user.refreshAuth).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		ValueEqual("status", response.StatusOK.String()).
		Value("payload").
		String().NotEmpty()
}
