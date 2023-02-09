package server_test

import (
	"net/http"
	"testing"

	"github.com/aria3ppp/watchlist-server/internal/models"
	"github.com/aria3ppp/watchlist-server/internal/testutils"
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
		Equal(testutils.ErrorMessage(
			"missing or malformed jwt",
		))

	// invalid token
	e.Request(method, path).
		WithHeader(echo.HeaderAuthorization, "Bearer invalid_token").
		Expect().
		Status(http.StatusUnauthorized).
		JSON().
		Object().
		Equal(testutils.ErrorMessage(
			"invalid or expired jwt",
		))

	// successful authorization
	e.Request(method, path+"{id}").
		WithPath("id", defaults.user.id).
		WithHeader(echo.HeaderAuthorization, defaults.user.auth).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		ValueEqual(models.UserColumns.Email, defaults.user.email)
}
