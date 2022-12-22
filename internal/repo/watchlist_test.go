package repo_test

import (
	"context"
	"math"
	"testing"
	"time"
	_ "unsafe"

	"github.com/aria3ppp/watchlist-server/internal/models"
	"github.com/aria3ppp/watchlist-server/internal/query"
	"github.com/aria3ppp/watchlist-server/internal/repo"
	"github.com/aria3ppp/watchlist-server/internal/testutils"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/null/v8"
)

func TestWatchlistGet(t *testing.T) {
	require := require.New(t)

	teardown := setup()
	t.Cleanup(teardown)

	r := repo.NewRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "email"}
	err := r.UserCreate(ctx, user)
	require.NoError(err)

	// at first user have no watchlist]

	watchlist, err := r.WatchlistGet(
		ctx,
		user.ID,
		query.WatchlistOptions{
			Offset:           0,
			Limit:            math.MaxInt,
			SortOrder:        "asc",
			WhereTimeWatched: repo.RawSqlWhereTimeWatchedEmptyClause,
		},
	)
	require.NoError(err)
	require.Equal(0, len(watchlist))

	// add films to watchlist

	films := []*models.Film{
		{Title: "f1"},
		{Title: "f2"},
		{Title: "f3"},
		{Title: "f4"},
		{Title: "f5"},
	}

	createTime := time.Now()

	var watchIDs []int
	for _, f := range films {
		// create film
		err = r.MovieCreate(ctx, user.ID, f)
		require.NoError(err)
		// add to watchlist
		id, err := r.WatchlistAdd(ctx, user.ID, f.ID)
		require.NoError(err)
		watchIDs = append(watchIDs, id)
	}

	// get watchlist

	watchlist, err = r.WatchlistGet(ctx, user.ID, query.WatchlistOptions{
		Offset:           0,
		Limit:            math.MaxInt,
		SortOrder:        "asc",
		WhereTimeWatched: repo.RawSqlWhereTimeWatchedIsNull,
	})
	require.NoError(err)
	require.Equal(len(films), len(watchlist))

	for i, wf := range watchlist {
		film := films[i]

		require.GreaterOrEqual(wf.TimeAdded, createTime)
		require.Equal(
			models.Watchfilm{
				ID:          watchIDs[i],
				UserID:      user.ID,
				FilmID:      film.ID,
				TimeAdded:   wf.TimeAdded,
				TimeWatched: null.Time{},
			},
			wf.Watchfilm,
		)

		testutils.SetTimeLocation(
			&film.DateReleased,
			wf.Film.DateReleased.Location(),
		)

		require.Equal(film, &wf.Film)
	}

	// check no watched films
	watchlist, err = r.WatchlistGet(ctx, user.ID, query.WatchlistOptions{
		Offset:           0,
		Limit:            math.MaxInt,
		SortOrder:        "asc",
		WhereTimeWatched: repo.RawSqlWhereTimeWatchedIsNotNull,
	})
	require.NoError(err)
	require.Equal(0, len(watchlist))

	// check reversed order
	watchlist, err = r.WatchlistGet(ctx, user.ID, query.WatchlistOptions{
		Offset:           0,
		Limit:            math.MaxInt,
		SortOrder:        "desc",
		WhereTimeWatched: repo.RawSqlWhereTimeWatchedEmptyClause,
	})
	require.NoError(err)

	for i, wf := range watchlist {
		reverseIndex := len(watchlist) - 1 - i
		require.Equal(watchIDs[reverseIndex], wf.ID)
	}
}

func TestWatchlistCount(t *testing.T) {
	require := require.New(t)

	teardown := setup()
	t.Cleanup(teardown)

	r := repo.NewRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "email"}
	err := r.UserCreate(ctx, user)
	require.NoError(err)

	// at first user have no watchlist

	count, err := r.WatchlistCount(
		ctx,
		user.ID,
		repo.RawSqlWhereTimeWatchedEmptyClause,
	)
	require.NoError(err)
	require.Equal(0, count)

	// add films to watchlist

	films := []*models.Film{
		{Title: "f1"},
		{Title: "f2"},
		{Title: "f3"},
		{Title: "f4"},
		{Title: "f5"},
	}
	watchIDs := make([]int, len(films))

	for i, f := range films {
		// create film
		err = r.MovieCreate(ctx, user.ID, f)
		require.NoError(err)
		// add to watchlist
		watchIDs[i], err = r.WatchlistAdd(ctx, user.ID, f.ID)
		require.NoError(err)
	}

	// count watchlist

	count, err = r.WatchlistCount(
		ctx,
		user.ID,
		repo.RawSqlWhereTimeWatchedIsNull,
	)
	require.NoError(err)
	require.Equal(len(films), count)

	// count watched films

	for _, wid := range watchIDs {
		err := r.WatchlistSetWatched(ctx, user.ID, wid)
		require.NoError(err)
	}

	count, err = r.WatchlistCount(
		ctx,
		user.ID,
		repo.RawSqlWhereTimeWatchedIsNotNull,
	)
	require.NoError(err)
	require.Equal(len(watchIDs), count)

	// zero non-watched films

	count, err = r.WatchlistCount(
		ctx,
		user.ID,
		repo.RawSqlWhereTimeWatchedIsNull,
	)
	require.NoError(err)
	require.Equal(0, count)
}

func TestWatchlistAdd(t *testing.T) {
	require := require.New(t)

	teardown := setup()
	t.Cleanup(teardown)

	r := repo.NewRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "email"}
	err := r.UserCreate(ctx, user)
	require.NoError(err)

	// at first user have no watchlist

	count, err := r.WatchlistCount(
		ctx,
		user.ID,
		repo.RawSqlWhereTimeWatchedEmptyClause,
	)
	require.NoError(err)
	require.Equal(0, count)

	// add a film to watchlist

	film := &models.Film{Title: "film"}

	createTime := time.Now()
	err = r.MovieCreate(ctx, user.ID, film)
	require.NoError(err)

	watchID, err := r.WatchlistAdd(ctx, user.ID, film.ID)
	require.NoError(err)

	// get watchlist film

	watchlist, err := r.WatchlistGet(ctx, user.ID, query.WatchlistOptions{
		Offset:           0,
		Limit:            math.MaxInt,
		SortOrder:        "asc",
		WhereTimeWatched: repo.RawSqlWhereTimeWatchedEmptyClause,
	})
	require.NoError(err)
	require.Equal(1, len(watchlist))

	require.GreaterOrEqual(watchlist[0].TimeAdded, createTime)
	require.Equal(
		models.Watchfilm{
			ID:          watchID,
			UserID:      user.ID,
			FilmID:      watchlist[0].FilmID,
			TimeAdded:   watchlist[0].TimeAdded,
			TimeWatched: null.Time{},
		},
		watchlist[0].Watchfilm,
	)

	testutils.SetTimeLocation(
		&film.DateReleased,
		watchlist[0].Film.DateReleased.Location(),
	)

	require.Equal(film, &watchlist[0].Film)
}

func TestWatchlistDelete(t *testing.T) {
	require := require.New(t)

	teardown := setup()
	t.Cleanup(teardown)

	r := repo.NewRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "email"}
	err := r.UserCreate(ctx, user)
	require.NoError(err)

	// at first user have no watchlist
	count, err := r.WatchlistCount(
		ctx,
		user.ID,
		repo.RawSqlWhereTimeWatchedEmptyClause,
	)
	require.NoError(err)
	require.Equal(0, count)

	// add a film to watchlist
	film := &models.Film{Title: "film"}
	err = r.MovieCreate(ctx, user.ID, film)
	require.NoError(err)

	watchID, err := r.WatchlistAdd(ctx, user.ID, film.ID)
	require.NoError(err)

	// delete watchlist
	err = r.WatchlistDelete(ctx, user.ID, watchID)
	require.NoError(err)

	// no watchlist
	err = r.WatchlistDelete(ctx, user.ID, watchID)
	require.Equal(repo.ErrNoRecord, err)
}

func TestWatchlistSetWatched(t *testing.T) {
	require := require.New(t)

	teardown := setup()
	t.Cleanup(teardown)

	r := repo.NewRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "email"}
	err := r.UserCreate(ctx, user)
	require.NoError(err)

	film := &models.Film{Title: "film"}

	createTime := time.Now()
	err = r.MovieCreate(ctx, user.ID, film)
	require.NoError(err)

	// no watchlist
	watchID := film.ID
	err = r.WatchlistSetWatched(ctx, user.ID, watchID)
	require.Equal(repo.ErrNoRecord, err)

	// add a film to watchlist
	watchID, err = r.WatchlistAdd(ctx, user.ID, film.ID)
	require.NoError(err)

	// get watchlist film
	watchlist, err := r.WatchlistGet(ctx, user.ID, query.WatchlistOptions{
		Offset:           0,
		Limit:            math.MaxInt,
		SortOrder:        "asc",
		WhereTimeWatched: repo.RawSqlWhereTimeWatchedIsNull,
	})
	require.NoError(err)
	require.Equal(1, len(watchlist))

	require.GreaterOrEqual(watchlist[0].TimeAdded, createTime)
	require.Equal(
		models.Watchfilm{
			ID:          watchID,
			UserID:      user.ID,
			FilmID:      watchlist[0].FilmID,
			TimeAdded:   watchlist[0].TimeAdded,
			TimeWatched: null.Time{},
		},
		watchlist[0].Watchfilm,
	)

	testutils.SetTimeLocation(
		&film.DateReleased,
		watchlist[0].Film.DateReleased.Location(),
	)

	require.Equal(film, &watchlist[0].Film)

	watchedTime := time.Now()

	// set watched
	err = r.WatchlistSetWatched(ctx, user.ID, watchID)
	require.NoError(err)

	// check watched films
	watchlist, err = r.WatchlistGet(ctx, user.ID, query.WatchlistOptions{
		Offset:           0,
		Limit:            math.MaxInt,
		SortOrder:        "asc",
		WhereTimeWatched: repo.RawSqlWhereTimeWatchedIsNotNull,
	})
	require.NoError(err)
	require.Equal(1, len(watchlist))

	require.True(watchlist[0].TimeWatched.Valid, "time_watched should be valid")
	require.GreaterOrEqual(watchlist[0].TimeWatched.Time, watchedTime)
	require.Equal(
		models.Watchfilm{
			ID:          watchID,
			UserID:      user.ID,
			FilmID:      watchlist[0].FilmID,
			TimeAdded:   watchlist[0].TimeAdded,
			TimeWatched: null.TimeFrom(watchlist[0].TimeWatched.Time),
		},
		watchlist[0].Watchfilm,
	)

	testutils.SetTimeLocation(
		&film.DateReleased,
		watchlist[0].Film.DateReleased.Location(),
	)

	require.Equal(film, &watchlist[0].Film)

	// check no non-watched films
	watchlist, err = r.WatchlistGet(ctx, user.ID, query.WatchlistOptions{
		Offset:           0,
		Limit:            math.MaxInt,
		SortOrder:        "asc",
		WhereTimeWatched: repo.RawSqlWhereTimeWatchedIsNull,
	})
	require.NoError(err)
	require.Equal(0, len(watchlist))
}
