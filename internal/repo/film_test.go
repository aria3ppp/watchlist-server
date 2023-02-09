package repo_test

import (
	"context"
	"testing"

	"github.com/aria3ppp/watchlist-server/internal/models"
	"github.com/aria3ppp/watchlist-server/internal/repo"
	"github.com/aria3ppp/watchlist-server/internal/testutils"
	"github.com/stretchr/testify/require"
)

func TestFilmExists(t *testing.T) {
	require := require.New(t)

	teardown := setup()
	t.Cleanup(teardown)

	r := repo.NewRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "email"}
	err := r.UserCreate(ctx, user)
	require.NoError(err)

	// not exists

	err = r.FilmExists(ctx, 999)
	require.Equal(repo.ErrNoRecord, err)

	// movie exists

	movie := &models.Film{
		Title:        "movie",
		DateReleased: testutils.Date(2000, 1, 1),
	}
	err = r.MovieCreate(ctx, user.ID, movie)
	require.NoError(err)

	err = r.FilmExists(ctx, movie.ID)
	require.NoError(err)

	// episode exists

	series := &models.Series{
		Title:       "series",
		DateStarted: testutils.Date(2000, 1, 1),
	}
	err = r.SeriesCreate(ctx, user.ID, series)
	require.NoError(err)

	episode := &models.Film{
		Title:        "episode",
		DateReleased: testutils.Date(2000, 1, 1),
	}
	err = r.EpisodePut(ctx, series.ID, 1, 1, user.ID, episode)
	require.NoError(err)

	err = r.FilmExists(ctx, episode.ID)
	require.NoError(err)
}
