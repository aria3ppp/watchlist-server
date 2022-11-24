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
			Email:          "fdfjshfksjd@fdkjjd.com",
			HashedPassword: "fdagafghfdah",
		},
		{
			Email:          "afyuehf@dfhk.com",
			HashedPassword: "hsghgfagdfad",
		},
		{
			Email:          "fhskgksj@kfdfj.com",
			HashedPassword: "hgshfafadfd",
		},
		{
			Email:          "igrhhgurhgur@wewef.com",
			HashedPassword: "hdgddfagfg",
		},
		{
			Email:          "qwdadasd@dkeffjejf.com",
			HashedPassword: "adfaasfdgafg",
		},
	}

	err := r.Transaction(
		ctx,
		func(ctx context.Context, repo repo.Service) error {
			for _, user := range users {
				err := repo.UserCreate(ctx, user)
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
			Email:          "fdfjshfksjd@fdkjjd.com",
			HashedPassword: "dfhshsdfdshfd",
		},
		{
			Email:          "afyuehf@dfhk.com",
			HashedPassword: "dhjsghsfsgfd",
		},
		{
			Email:          "fhskgksj@kfdfj.com",
			HashedPassword: "fgafgdfdgqdsf",
		},
		{
			Email:          "igrhhgurhgur@wewef.com",
			HashedPassword: "gfdggfagagad",
		},
		{
			Email:          "qwdadasd@dkeffjejf.com",
			HashedPassword: "hdfhdghdhdsfd",
		},
	}

	expError := errors.New("expected_error")

	err := r.Transaction(
		ctx,
		func(ctx context.Context, repo repo.Service) error {
			for _, user := range users {
				err := repo.UserCreate(ctx, user)
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
			Email:          "fdfjshfksjd@fdkjjd.com",
			HashedPassword: "dfhshsdfdshfd",
		},
		{
			Email:          "afyuehf@dfhk.com",
			HashedPassword: "dhjsghsfsgfd",
		},
		{
			Email:          "fhskgksj@kfdfj.com",
			HashedPassword: "fgafgdfdgqdsf",
		},
		{
			Email:          "igrhhgurhgur@wewef.com",
			HashedPassword: "gfdggfagagad",
		},
		{
			Email:          "qwdadasd@dkeffjejf.com",
			HashedPassword: "hdfhdghdhdsfd",
		},
	}

	expErr := errors.New("panic")

	require.PanicsWithError(expErr.Error(), func() {
		r.Transaction(
			ctx,
			func(ctx context.Context, repo repo.Service) error {
				for _, user := range users {
					err := repo.UserCreate(ctx, user)
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
