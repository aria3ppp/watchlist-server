package repo_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aria3ppp/watchlist-server/internal/models"
	"github.com/aria3ppp/watchlist-server/internal/repo"
	"github.com/stretchr/testify/require"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func TestTransactionOK(t *testing.T) {
	require := require.New(t)

	teardown := setup()
	t.Cleanup(teardown)

	r := repo.NewRepository(db)
	ctx := context.Background()

	users := []*models.User{
		{
			Email:        "fdfjshfksjd@fdkjjd.com",
			PasswordHash: "fdagafghfdah",
		},
		{
			Email:        "afyuehf@dfhk.com",
			PasswordHash: "hsghgfagdfad",
		},
		{
			Email:        "fhskgksj@kfdfj.com",
			PasswordHash: "hgshfafadfd",
		},
		{
			Email:        "igrhhgurhgur@wewef.com",
			PasswordHash: "hdgddfagfg",
		},
		{
			Email:        "qwdadasd@dkeffjejf.com",
			PasswordHash: "adfaasfdgafg",
		},
	}

	err := r.Tx(
		ctx,
		nil,
		func(ctx context.Context, tx repo.Service) error {
			for _, user := range users {
				err := tx.UserCreate(ctx, user)
				require.NoError(err)
				require.NotEqual(user.ID, 0)
			}
			return nil
		},
	)

	require.NoError(err)

	nUsers, err := r.UsersCount(ctx)
	require.NoError(err)
	require.Equal(len(users), nUsers)
}

func TestTransactionFailed(t *testing.T) {
	require := require.New(t)

	teardown := setup()
	t.Cleanup(teardown)

	r := repo.NewRepository(db)
	ctx := context.Background()

	users := []*models.User{
		{
			Email:        "fdfjshfksjd@fdkjjd.com",
			PasswordHash: "dfhshsdfdshfd",
		},
		{
			Email:        "afyuehf@dfhk.com",
			PasswordHash: "dhjsghsfsgfd",
		},
		{
			Email:        "fhskgksj@kfdfj.com",
			PasswordHash: "fgafgdfdgqdsf",
		},
		{
			Email:        "igrhhgurhgur@wewef.com",
			PasswordHash: "gfdggfagagad",
		},
		{
			Email:        "qwdadasd@dkeffjejf.com",
			PasswordHash: "hdfhdghdhdsfd",
		},
	}

	expError := errors.New("expected_error")

	err := r.Tx(
		ctx,
		nil,
		func(ctx context.Context, tx repo.Service) error {
			for _, user := range users {
				err := tx.UserCreate(ctx, user)
				require.NoError(err)
				require.NotEqual(user.ID, 0)
			}
			return expError
		},
	)

	require.Equal(expError, err)

	nUsers, err := r.UsersCount(ctx)
	require.NoError(err)
	require.Equal(0, nUsers)
}

func TestTransactionPanic(t *testing.T) {
	require := require.New(t)

	teardown := setup()
	t.Cleanup(teardown)

	r := repo.NewRepository(db)
	ctx := context.Background()

	users := []*models.User{
		{
			Email:        "fdfjshfksjd@fdkjjd.com",
			PasswordHash: "dfhshsdfdshfd",
		},
		{
			Email:        "afyuehf@dfhk.com",
			PasswordHash: "dhjsghsfsgfd",
		},
		{
			Email:        "fhskgksj@kfdfj.com",
			PasswordHash: "fgafgdfdgqdsf",
		},
		{
			Email:        "igrhhgurhgur@wewef.com",
			PasswordHash: "gfdggfagagad",
		},
		{
			Email:        "qwdadasd@dkeffjejf.com",
			PasswordHash: "hdfhdghdhdsfd",
		},
	}

	expErr := errors.New("panic")

	require.PanicsWithError(expErr.Error(), func() {
		r.Tx(
			ctx,
			nil,
			func(ctx context.Context, tx repo.Service) error {
				for _, user := range users {
					err := tx.UserCreate(ctx, user)
					require.NoError(err)
					require.NotEqual(user.ID, 0)
				}
				panic(expErr)
			},
		)
	})

	// inserted users were rollback
	nUsers, err := r.UsersCount(ctx)
	require.NoError(err)
	require.Equal(0, nUsers)
}
