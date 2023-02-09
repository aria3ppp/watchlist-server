package request_test

import (
	"testing"

	"github.com/aria3ppp/watchlist-server/internal/server/request"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
)

func TestTokenBody_Validate(t *testing.T) {
	uuid, err := uuid.NewV4()
	require.NoError(t, err)

	testCases := []struct {
		name     string
		params   request.TokenBody
		expError error
	}{
		{
			name:   "tc1",
			params: request.TokenBody{},
			expError: validation.Errors{
				"token": validation.ErrRequired,
			},
		},
		{
			name: "tc2",
			params: request.TokenBody{
				Token: "invlid-uuid-string",
			},
			expError: validation.Errors{
				"token": is.ErrUUID,
			},
		},
		{
			name: "tc3",
			params: request.TokenBody{
				Token: uuid.String(),
			},
			expError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)
			require.Equal(tc.expError, tc.params.Validate())
		})
	}
}
