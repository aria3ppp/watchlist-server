package repo_test

import (
	"context"
	"testing"
	"time"

	"github.com/aria3ppp/watchlist-server/internal/models"
	"github.com/aria3ppp/watchlist-server/internal/repo"
	"github.com/aria3ppp/watchlist-server/internal/testutils"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestTokenGet(t *testing.T) {
	require := require.New(t)

	teardown := setup()
	t.Cleanup(teardown)

	r := repo.NewRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "email"}
	err := r.UserCreate(ctx, user)
	require.NoError(err)

	refreshToken := "refresh-token"

	refreshTokenHash, err := bcrypt.GenerateFromPassword(
		[]byte(refreshToken),
		bcrypt.MinCost,
	)
	require.NoError(err)

	token := &models.Token{
		TokenHash: string(refreshTokenHash),
		UserID:    user.ID,
		// round to microseconds as postgres save time in microsecond precision
		ExpiresAt: time.Now().Add(time.Hour).Round(time.Microsecond),
	}

	// first there's no token

	fetchedToken, err := r.TokenGet(ctx, user.ID, refreshToken)
	require.Equal(repo.ErrNoRecord, err)
	require.Nil(fetchedToken)

	// insert token

	err = r.TokenCreate(ctx, token)
	require.NoError(err)

	// fetch the token

	fetchedToken, err = r.TokenGet(ctx, user.ID, refreshToken)
	require.NoError(err)

	testutils.SetTimeLocation(
		&token.ExpiresAt,
		fetchedToken.ExpiresAt.Location(),
	)

	require.Equal(token, fetchedToken)

	// no expired token

	err = r.TokenUpdate(ctx, token.ID, map[string]any{
		models.TokenColumns.ExpiresAt: time.Now().Add(-time.Hour),
	})
	require.NoError(err)

	token, err = r.TokenGet(ctx, user.ID, refreshToken)
	require.Equal(repo.ErrNoRecord, err)
	require.Nil(token)
}

func TestTokenCreate(t *testing.T) {
	require := require.New(t)

	teardown := setup()
	t.Cleanup(teardown)

	r := repo.NewRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "email"}
	err := r.UserCreate(ctx, user)
	require.NoError(err)

	refreshToken := "refresh-token"

	refreshTokenHash, err := bcrypt.GenerateFromPassword(
		[]byte(refreshToken),
		bcrypt.MinCost,
	)
	require.NoError(err)

	token := &models.Token{
		TokenHash: string(refreshTokenHash),
		UserID:    user.ID,
		// round to microseconds as postgres save time in microsecond precision
		ExpiresAt: time.Now().Add(time.Hour).Round(time.Microsecond),
	}

	// create token

	err = r.TokenCreate(ctx, token)
	require.NoError(err)

	// fetch token created

	fetchedToken, err := r.TokenGet(ctx, user.ID, refreshToken)
	require.NoError(err)

	testutils.SetTimeLocation(
		&token.ExpiresAt,
		fetchedToken.ExpiresAt.Location(),
	)

	require.Equal(token, fetchedToken)

	// create another token

	anotherRefershToken := "another-refersh-token"

	anotherRefreshTokenHash, err := bcrypt.GenerateFromPassword(
		[]byte(anotherRefershToken),
		bcrypt.MinCost,
	)
	require.NoError(err)

	anotherToken := &models.Token{
		TokenHash: string(anotherRefreshTokenHash),
		UserID:    user.ID,
		// round to microseconds as postgres save time in microsecond precision
		ExpiresAt: time.Now().Add(time.Hour * 2).Round(time.Microsecond),
	}

	err = r.TokenCreate(ctx, anotherToken)
	require.NoError(err)

	// fetch another token created

	fetchedAnotherToken, err := r.TokenGet(ctx, user.ID, anotherRefershToken)
	require.NoError(err)

	testutils.SetTimeLocation(
		&anotherToken.ExpiresAt,
		fetchedAnotherToken.ExpiresAt.Location(),
	)

	require.Equal(anotherToken, fetchedAnotherToken)
}

func TestTokenUpdate(t *testing.T) {
	require := require.New(t)

	teardown := setup()
	t.Cleanup(teardown)

	r := repo.NewRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "email"}
	err := r.UserCreate(ctx, user)
	require.NoError(err)

	refreshToken := "refresh-token"

	refreshTokenHash, err := bcrypt.GenerateFromPassword(
		[]byte(refreshToken),
		bcrypt.MinCost,
	)
	require.NoError(err)

	token := &models.Token{
		TokenHash: string(refreshTokenHash),
		UserID:    user.ID,
		// round to microseconds as postgres save time in microsecond precision
		ExpiresAt: time.Now().Add(time.Hour).Round(time.Microsecond),
	}

	newRefreshToken := "new-refresh-token"
	newRefreshTokenHash, err := bcrypt.GenerateFromPassword(
		[]byte(newRefreshToken),
		bcrypt.MinCost,
	)
	require.NoError(err)
	// round to microseconds as postgres save time in microsecond precision
	newExpiresAt := time.Now().Add(time.Hour * 2).Round(time.Microsecond)
	newUser := &models.User{Email: "new user with new email"}
	err = r.UserCreate(ctx, newUser)
	require.NoError(err)

	updateColumns := map[string]any{
		models.TokenColumns.TokenHash: string(newRefreshTokenHash),
		models.TokenColumns.UserID:    newUser.ID,
		models.TokenColumns.ExpiresAt: newExpiresAt,
	}

	// create token

	err = r.TokenCreate(ctx, token)
	require.NoError(err)

	// update token

	err = r.TokenUpdate(ctx, token.ID, updateColumns)
	require.NoError(err)

	// fetch the updated token

	fetchedUpdatedToken, err := r.TokenGet(ctx, newUser.ID, newRefreshToken)
	require.NoError(err)

	testutils.SetTimeLocation(
		&newExpiresAt,
		fetchedUpdatedToken.ExpiresAt.Location(),
	)

	require.Equal(
		&models.Token{
			ID:        token.ID,
			TokenHash: string(newRefreshTokenHash),
			UserID:    newUser.ID,
			ExpiresAt: newExpiresAt,
		},
		fetchedUpdatedToken,
	)
}
