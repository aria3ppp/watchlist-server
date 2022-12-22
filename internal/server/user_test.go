package server_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/aria3ppp/watchlist-server/internal/app"
	"github.com/aria3ppp/watchlist-server/internal/config"
	"github.com/aria3ppp/watchlist-server/internal/dto"
	"github.com/aria3ppp/watchlist-server/internal/models"
	"github.com/aria3ppp/watchlist-server/internal/server/response"
	"github.com/aria3ppp/watchlist-server/internal/testutils"
	"github.com/gavv/httpexpect/v2"
	validation "github.com/go-ozzo/ozzo-validation/v4"
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
		Equal(response.Error(
			response.StatusInvalidURLParameter,
			validation.Errors{
				"id": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{"threshold": 1},
				),
			}.Error(),
		))

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
		Jointime:  gotUser.Jointime,
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

	// invalid requeset
	e.Request(method, path).
		WithJSON(dto.UserCreateRequest{}).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		Equal(response.Error(
			response.StatusInvalidRequest,
			validation.Errors{
				"email":    validation.ErrRequired,
				"password": validation.ErrRequired,
			}.Error(),
		))

	// create user
	userCreateReq := &dto.UserCreateRequest{
		Email:    "aria3ppp@gamil.com",
		Password: "pa$$W0RD1",
	}

	createDate := time.Now()

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

	require.GreaterOrEqual(gotUser.Jointime, createDate)

	require.Equal(&models.User{
		ID:             userID,
		Email:          userCreateReq.Email,
		HashedPassword: gotUser.HashedPassword,
		FirstName:      userCreateReq.FirstName,
		LastName:       userCreateReq.LastName,
		Bio:            userCreateReq.Bio,
		Birthdate:      userCreateReq.Birthdate,
		Jointime:       gotUser.Jointime,
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

func TestHandleUserLogin(t *testing.T) {
	server, _, defaults, teardown := setup(OptEnableDefaultUser)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/user/login"
	method := http.MethodPost

	// invalid request
	e.Request(method, path).
		WithJSON(dto.UserLoginRequest{}).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		Equal(response.Error(
			response.StatusInvalidRequest,
			validation.Errors{
				"email":    validation.ErrRequired,
				"password": validation.ErrRequired,
			}.Error(),
		))

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

func TestHandleUserUpdate(t *testing.T) {
	require := require.New(t)

	server, appInstance, defaults, teardown := setup(OptEnableDefaultUser)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/authorized/user/"
	method := http.MethodPatch

	// invalid request
	e.Request(method, path).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(dto.UserUpdateRequest{FirstName: null.StringFrom("f")}).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		Equal(response.Error(
			response.StatusInvalidRequest,
			validation.Errors{
				"first_name": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.FirstName.MinLength,
						"max": config.Config.Validation.User.FirstName.MaxLength,
					},
				),
			}.Error(),
		))

	timeNow := time.Now()

	// update user
	userUpdateReq := &dto.UserUpdateRequest{
		FirstName: null.StringFrom("updated_first_name"),
		LastName:  null.StringFrom("updated_last_name"),
		Bio:       null.StringFrom("updated_bio"),
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

func TestHandleUserEmailUpdate(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	server, appInstance, defaults, teardown := setup(OptEnableDefaultUser)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/authorized/user/email"
	method := http.MethodPut

	// invalid request
	e.Request(method, path).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(dto.UserEmailUpdateRequest{}).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		Equal(response.Error(
			response.StatusInvalidRequest,
			validation.Errors{
				"email": validation.ErrRequired,
			}.Error(),
		))

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
			Jointime:       gotUser.Jointime,
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

func TestHandleUserPasswordUpdate(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	server, appInstance, defaults, teardown := setup(OptEnableDefaultUser)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/authorized/user/password"
	method := http.MethodPut

	// invalid request
	e.Request(method, path).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(&dto.UserPasswordUpdateRequest{}).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		Equal(response.Error(
			response.StatusInvalidRequest,
			validation.Errors{
				"current_password": validation.ErrRequired,
				"new_password":     validation.ErrRequired,
			}.Error(),
		))

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

	newPassword := "new_pa$$W0RD1"

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
			Jointime:       gotUser.Jointime,
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

func TestHandleUserDelete(t *testing.T) {
	require := require.New(t)

	server, appInstance, defaults, teardown := setup(OptEnableDefaultUser)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/authorized/user/"
	method := http.MethodDelete

	// invalid request
	e.Request(method, path).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(&dto.UserDeleteRequest{}).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		Equal(response.Error(
			response.StatusInvalidRequest,
			validation.Errors{
				"password": validation.ErrRequired,
			}.Error(),
		))

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
