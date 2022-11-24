package server_test

import (
	"net/http"
	"testing"

	"github.com/aria3ppp/watchlist-server/internal/server/response"
	"github.com/gavv/httpexpect/v2"
	"github.com/labstack/echo/v4"
)

func TestE2EAuthorization(t *testing.T) {
	server, _, defaults, teardown := setup(OptEnableDefaultUser)
	t.Cleanup(teardown)

	e := httpexpect.New(t, server.URL)
	path := "/v1/authorized/user/"
	method := http.MethodGet

	// missing token
	e.Request(method, path).
		Expect().
		Status(http.StatusUnauthorized).
		JSON().
		Object().
		Equal(response.Error(response.StatusTokenMissingOrMalformed))

	// invalid token
	e.Request(method, path).
		WithHeader(echo.HeaderAuthorization, "Bearer invalid_token").
		Expect().
		Status(http.StatusUnauthorized).
		JSON().
		Object().
		Equal(response.Error(response.StatusTokenInvalid))

	// successful authorization
	e.Request(method, path+"{id}").
		WithPath("id", defaults.user.id).
		WithHeader(echo.HeaderAuthorization, defaults.user.refreshAuth).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		ValueEqual("status", response.StatusOK.String())
}
