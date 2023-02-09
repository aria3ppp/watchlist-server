package server_test

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/aria3ppp/watchlist-server/internal/app"
	"github.com/aria3ppp/watchlist-server/internal/config"
	"github.com/aria3ppp/watchlist-server/internal/dto"
	"github.com/aria3ppp/watchlist-server/internal/models"
	"github.com/aria3ppp/watchlist-server/internal/server/request"
	"github.com/aria3ppp/watchlist-server/internal/testutils"
	"github.com/gavv/httpexpect/v2"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/gofrs/uuid"
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
		Equal(testutils.ErrorMessage(
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
		Equal(testutils.ErrorMessage(
			http.StatusText(http.StatusNotFound),
		))

	// create a new user
	userCreateReq := &dto.UserCreateRequest{
		Email:    "new_email@example.com",
		Password: "new_pa$$W0RD1",
	}
	userID, err := appInstance.UserCreate(ctx, userCreateReq)
	require.NoError(err)

	gotUser, err := appInstance.UserGet(ctx, userID)
	require.NoError(err)

	// get user
	e.Request(method, path).
		WithPath("id", userID).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Equal(&models.User{
			ID:        userID,
			Email:     userCreateReq.Email,
			FirstName: userCreateReq.FirstName,
			LastName:  userCreateReq.LastName,
			Bio:       userCreateReq.Bio,
			Birthdate: userCreateReq.Birthdate,
			Jointime:  gotUser.Jointime,
		})
}

func TestHandleUserCreate(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	server, appInstance, _, teardown := setup()
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/user"
	method := http.MethodPost

	// invalid requeset
	e.Request(method, path).
		WithJSON(dto.UserCreateRequest{}).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
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
		Value("id").Number().Gt(0).Raw()

	userID := int(userIDRaw)

	// check user created
	gotUser, err := appInstance.UserGet(ctx, userID)
	require.NoError(err)

	require.GreaterOrEqual(gotUser.Jointime, createDate)

	require.Equal(&models.User{
		ID:           userID,
		Email:        userCreateReq.Email,
		PasswordHash: gotUser.PasswordHash,
		FirstName:    userCreateReq.FirstName,
		LastName:     userCreateReq.LastName,
		Bio:          userCreateReq.Bio,
		Birthdate:    userCreateReq.Birthdate,
		Jointime:     gotUser.Jointime,
	}, gotUser)

	// email address already taken
	e.Request(method, path).
		WithJSON(dto.UserCreateRequest{
			Email:    userCreateReq.Email,
			Password: userCreateReq.Password,
		}).
		Expect().
		Status(http.StatusConflict).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			http.StatusText(http.StatusConflict),
		))
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
		Equal(testutils.ErrorMessage(
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
		Status(http.StatusNotFound).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			http.StatusText(http.StatusNotFound),
		))

	// incorrect password
	e.Request(method, path).
		WithJSON(dto.UserLoginRequest{
			Email:    defaults.user.email,
			Password: "1nc0RR3ct_pa$$",
		}).
		Expect().
		Status(http.StatusUnauthorized).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			http.StatusText(http.StatusUnauthorized),
		))

	// login user
	respObj := e.Request(method, path).
		WithJSON(dto.UserCreateRequest{
			Email:    defaults.user.email,
			Password: defaults.user.password,
		}).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object()

	respObj.Value("jwt_token").String().NotEmpty()
	respObj.Value("jwt_expires_at").Number().Gt(time.Now().Unix())
	respObj.Value("refresh_token").String().NotEmpty()
	respObj.Value("refresh_expires_at").Number().Gt(time.Now().Unix())
	respObj.Value("user_id").Number().Equal(defaults.user.id)
}

func TestHandleUserLogout(t *testing.T) {
	server, _, defaults, teardown := setup(OptEnableDefaultUser)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/user/{id}/logout"
	method := http.MethodPost

	// invalid id
	e.Request(method, path).
		WithPath("id", -1).
		WithJSON(request.TokenBody{
			Token: defaults.user.refreshToken,
		}).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			validation.Errors{
				"id": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{"threshold": 1},
				),
			}.Error(),
		))

	// user not found
	e.Request(method, path).
		WithPath("id", 999).
		WithJSON(request.TokenBody{
			Token: defaults.user.refreshToken,
		}).
		Expect().
		Status(http.StatusNotFound).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			http.StatusText(http.StatusNotFound),
		))

	// missing token
	e.Request(method, path).
		WithPath("id", defaults.user.id).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			validation.Errors{"token": validation.ErrRequired}.Error(),
		))

	// invalid token
	e.Request(method, path).
		WithPath("id", defaults.user.id).
		WithJSON(request.TokenBody{
			Token: "invalid-refersh-token",
		}).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			validation.Errors{"token": is.ErrUUID}.Error(),
		))

	// token not found
	uuid, err := uuid.NewV4()
	require.NoError(t, err)

	e.Request(method, path).
		WithPath("id", defaults.user.id).
		WithJSON(request.TokenBody{
			Token: uuid.String(),
		}).
		Expect().
		Status(http.StatusNotFound).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			http.StatusText(http.StatusNotFound),
		))

	// logout
	e.Request(method, path).
		WithPath("id", defaults.user.id).
		WithJSON(request.TokenBody{
			Token: defaults.user.refreshToken,
		}).
		Expect().
		Status(http.StatusOK).
		NoContent()
}

func TestHandleUserRefreshToken(t *testing.T) {
	server, _, defaults, teardown := setup(OptEnableDefaultUser)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/user/{id}/refresh"
	method := http.MethodPost

	// invalid id
	e.Request(method, path).
		WithPath("id", -1).
		WithJSON(request.TokenBody{
			Token: defaults.user.refreshToken,
		}).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			validation.Errors{
				"id": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{"threshold": 1},
				),
			}.Error(),
		))

	// user not found
	e.Request(method, path).
		WithPath("id", 999).
		WithJSON(request.TokenBody{
			Token: defaults.user.refreshToken,
		}).
		Expect().
		Status(http.StatusNotFound).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			http.StatusText(http.StatusNotFound),
		))

	// missing token
	e.Request(method, path).
		WithPath("id", defaults.user.id).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			validation.Errors{"token": validation.ErrRequired}.Error(),
		))

	// invalid token
	e.Request(method, path).
		WithPath("id", defaults.user.id).
		WithJSON(request.TokenBody{
			Token: "invalid-refersh-token",
		}).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			validation.Errors{"token": is.ErrUUID}.Error(),
		))

	// token not found
	uuid, err := uuid.NewV4()
	require.NoError(t, err)

	e.Request(method, path).
		WithPath("id", defaults.user.id).
		WithJSON(request.TokenBody{
			Token: uuid.String(),
		}).
		Expect().
		Status(http.StatusNotFound).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			http.StatusText(http.StatusNotFound),
		))

	// refresh token
	respObj := e.Request(method, path).
		WithPath("id", defaults.user.id).
		WithJSON(request.TokenBody{
			Token: defaults.user.refreshToken,
		}).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object()

	respObj.Value("jwt_token").String().NotEmpty()
	respObj.Value("jwt_expires_at").Number().Gt(time.Now().Unix())
}

func TestHandleUserUpdate(t *testing.T) {
	require := require.New(t)

	server, appInstance, defaults, teardown := setup(OptEnableDefaultUser)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/authorized/user"
	method := http.MethodPatch

	// invalid request
	e.Request(method, path).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(dto.UserUpdateRequest{FirstName: null.StringFrom("f")}).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
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
		NoContent()

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
		Equal(testutils.ErrorMessage(
			http.StatusText(http.StatusNotFound),
		))
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
		Equal(testutils.ErrorMessage(
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
		NoContent()

	// check updated email
	gotUser, err := appInstance.UserGet(ctx, defaults.user.id)
	require.NoError(err)
	require.Equal(
		&models.User{
			ID:           defaults.user.id,
			Email:        userEmailUpdateReq.Email,
			PasswordHash: gotUser.PasswordHash,
			FirstName:    defaults.user.reqObject.FirstName,
			LastName:     defaults.user.reqObject.LastName,
			Bio:          defaults.user.reqObject.Bio,
			Birthdate:    defaults.user.reqObject.Birthdate,
			Jointime:     gotUser.Jointime,
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
		Equal(testutils.ErrorMessage(
			http.StatusText(http.StatusNotFound),
		))
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
		Equal(testutils.ErrorMessage(
			validation.Errors{
				"current_password": validation.ErrRequired,
				"new_password":     validation.ErrRequired,
			}.Error(),
		))

	// same password
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
		Equal(testutils.ErrorMessage(
			"same password",
		))

	newPassword := "new_pa$$W0RD1"

	// incorrect password
	e.Request(method, path).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(&dto.UserPasswordUpdateRequest{
			CurrentPassword: "1nc0RR3ct_pa$$",
			NewPassword:     newPassword,
		}).
		Expect().
		Status(http.StatusUnauthorized).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			"incorrect password",
		))

	// update password
	e.Request(method, path).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(&dto.UserPasswordUpdateRequest{
			CurrentPassword: defaults.user.password,
			NewPassword:     newPassword,
		}).
		Expect().
		Status(http.StatusOK).
		NoContent()

	// check password updated
	gotUser, err := appInstance.UserGet(ctx, defaults.user.id)
	require.NoError(err)
	require.Equal(
		&models.User{
			ID:           defaults.user.id,
			Email:        defaults.user.email,
			PasswordHash: gotUser.PasswordHash,
			FirstName:    defaults.user.reqObject.FirstName,
			LastName:     defaults.user.reqObject.LastName,
			Bio:          defaults.user.reqObject.Bio,
			Birthdate:    defaults.user.reqObject.Birthdate,
			Jointime:     gotUser.Jointime,
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
		Equal(testutils.ErrorMessage(
			http.StatusText(http.StatusNotFound),
		))
}

func TestHandleUserDelete(t *testing.T) {
	require := require.New(t)

	server, appInstance, defaults, teardown := setup(OptEnableDefaultUser)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/authorized/user"
	method := http.MethodDelete

	// invalid request
	e.Request(method, path).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(&dto.UserDeleteRequest{}).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			validation.Errors{
				"password": validation.ErrRequired,
			}.Error(),
		))

	// incorrect password
	e.Request(method, path).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(&dto.UserDeleteRequest{Password: "1nc0RR3ct_pa$$"}).
		Expect().
		Status(http.StatusUnauthorized).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			"incorrect password",
		))

	// delete user
	e.Request(method, path).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithJSON(&dto.UserDeleteRequest{Password: defaults.user.password}).
		Expect().
		Status(http.StatusOK).
		NoContent()

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
		Equal(testutils.ErrorMessage(
			http.StatusText(http.StatusNotFound),
		))
}

func TestHandleUserPutAvatar(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	server, appInstance, defaults, teardown := setup(OptEnableDefaultUser)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/authorized/user/avatar"
	method := http.MethodPut

	filename := config.Config.MinIO.Filename.User

	type FileInfo struct {
		path        string
		file        *os.File
		size        int64
		contentType string
	}
	openFile := func(path string) FileInfo {
		file, err := os.Open(path)
		require.NoError(err)
		t.Cleanup(func() { file.Close() })

		stat, err := file.Stat()
		require.NoError(err)

		buf := make([]byte, 512)
		_, err = file.Read(buf)
		require.NoError(err)
		_, err = file.Seek(0, 0)
		require.NoError(err)

		return FileInfo{
			path:        path,
			file:        file,
			size:        stat.Size(),
			contentType: http.DetectContentType(buf),
		}
	}

	file := openFile(filepath.Join("testdata", "gopher-1.webp"))
	overwrittenFile := openFile(filepath.Join("testdata", "gopher-2.png"))
	unsupportedFile := openFile(filepath.Join("testdata", "gopher-3.ico"))

	// user have no avatar

	userGetReq := struct {
		method string
		path   string
	}{
		method: http.MethodGet,
		path:   "/v1/authorized/user/{id}",
	}

	e.Request(userGetReq.method, userGetReq.path).
		WithPath("id", defaults.user.id).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("avatar").Null()

	// missing file
	e.Request(method, path).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithMultipart().
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			"missing file",
		))

	// unsupported file
	e.Request(method, path).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithMultipart().WithFile(filename, unsupportedFile.path).
		Expect().
		Status(http.StatusUnsupportedMediaType).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			"unsupported file type",
		))

	// put avatar
	uri := e.Request(method, path).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithMultipart().WithFile(filename, file.path).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("uri").String().NotEmpty().Raw()

	// now user have avatar
	e.Request(userGetReq.method, userGetReq.path).
		WithPath("id", defaults.user.id).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		ValueEqual("avatar", uri)

	// check file on storage

	storageExpect := httpexpect.New(t, "http://127.0.0.1:9000")

	uriSplit := strings.Split(uri, "?")
	require.Equal(2, len(uriSplit))
	versionIdQuery := strings.Split(uriSplit[1], "=")
	require.Equal(2, len(versionIdQuery))

	resp := storageExpect.Request(http.MethodGet, uriSplit[0]).
		WithQuery(versionIdQuery[0], versionIdQuery[1]).
		Expect().
		Status(http.StatusOK)

	resp.Header("Content-Length").Equal(strconv.FormatInt(file.size, 10))
	resp.Header("Content-Type").Equal(file.contentType)

	// overwrite avatar
	uri = e.Request(method, path).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithMultipart().WithFile(filename, overwrittenFile.path).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("uri").String().NotEmpty().Raw()

	// check overwritten user avatar
	e.Request(userGetReq.method, userGetReq.path).
		WithPath("id", defaults.user.id).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		ValueEqual("avatar", uri)

	// check overwritten file on storage
	uriSplit = strings.Split(uri, "?")
	require.Equal(2, len(uriSplit))
	versionIdQuery = strings.Split(uriSplit[1], "=")
	require.Equal(2, len(versionIdQuery))

	resp = storageExpect.Request(http.MethodGet, uriSplit[0]).
		WithQuery(versionIdQuery[0], versionIdQuery[1]).
		Expect().
		Status(http.StatusOK)

	resp.Header("Content-Length").
		Equal(strconv.FormatInt(overwrittenFile.size, 10))
	resp.Header("Content-Type").Equal(overwrittenFile.contentType)

	// user not found

	err := appInstance.UserDelete(
		ctx,
		defaults.user.id,
		&dto.UserDeleteRequest{Password: defaults.user.password},
	)
	require.NoError(err)

	e.Request(method, path).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		WithMultipart().WithFile(filename, file.path).
		Expect().
		Status(http.StatusNotFound).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			http.StatusText(http.StatusNotFound),
		))
}
