package repo_test

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/aria3ppp/watchlist-server/internal/models"
	"github.com/aria3ppp/watchlist-server/internal/query"
	"github.com/aria3ppp/watchlist-server/internal/repo"
	"github.com/aria3ppp/watchlist-server/internal/testutils"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/null/v8"
)

func TestMovieGet(t *testing.T) {
	require := require.New(t)

	teardown := setup()
	t.Cleanup(teardown)

	r := repo.NewRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "email"}
	err := r.UserCreate(ctx, user)
	require.NoError(err)

	movie := &models.Film{
		Title:        "movie",
		DateReleased: testutils.Date(2000, 1, 1),
	}

	// first there's no movie

	fetchedMovie, err := r.MovieGet(ctx, 1)
	require.Equal(repo.ErrNoRecord, err)
	require.Nil(fetchedMovie)

	// insert movie

	err = r.MovieCreate(ctx, user.ID, movie)
	require.NoError(err)

	// fetch the movie

	fetchedMovie, err = r.MovieGet(ctx, movie.ID)
	require.NoError(err)

	testutils.SetTimeLocation(
		&movie.DateReleased,
		fetchedMovie.DateReleased.Location(),
	)

	require.Equal(movie, fetchedMovie)
}

func TestMoviesGetAll(t *testing.T) {
	require := require.New(t)

	teardown := setup()
	t.Cleanup(teardown)

	r := repo.NewRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "email"}
	err := r.UserCreate(ctx, user)
	require.NoError(err)

	queryOptions := query.Options{
		Offset:    0,
		Limit:     math.MaxInt,
		SortField: models.FilmColumns.ID,
		SortOrder: "asc",
	}

	// first there's no movie

	fetchedMovies, err := r.MoviesGetAll(ctx, queryOptions)
	require.NoError(err)
	require.Equal(0, len(fetchedMovies))

	// insert movies

	movies := []*models.Film{
		{
			Title:        "m1",
			DateReleased: testutils.Date(2000, 1, 1),
		},
		{
			Title:        "m2",
			DateReleased: testutils.Date(2001, 1, 1),
		},
		{
			Title:        "m3",
			DateReleased: testutils.Date(2002, 1, 1),
		},
		{
			Title:        "m4",
			DateReleased: testutils.Date(2003, 1, 1),
		},
		{
			Title:        "m5",
			DateReleased: testutils.Date(2004, 1, 1),
		},
	}

	for _, m := range movies {
		err = r.MovieCreate(ctx, user.ID, m)
		require.NoError(err)
	}

	// fetch movies

	fetchedMovies, err = r.MoviesGetAll(ctx, queryOptions)
	require.NoError(err)
	require.Equal(len(movies), len(fetchedMovies))

	// MoviesGetAll records are order by id by default
	for i, fm := range fetchedMovies {
		testutils.SetTimeLocation(
			&movies[i].DateReleased,
			fm.DateReleased.Location(),
		)
		require.Equal(movies[i], fm)
	}
}

func TestMoviesCount(t *testing.T) {
	require := require.New(t)

	teardown := setup()
	t.Cleanup(teardown)

	r := repo.NewRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "email"}
	err := r.UserCreate(ctx, user)
	require.NoError(err)

	// first there's no movie

	nMovies, err := r.MoviesCount(ctx)
	require.NoError(err)
	require.Equal(0, nMovies)

	// insert movies

	movies := []*models.Film{
		{Title: "m1"},
		{Title: "m2"},
		{Title: "m3"},
		{Title: "m4"},
		{Title: "m5"},
	}

	for _, m := range movies {
		err = r.MovieCreate(ctx, user.ID, m)
		require.NoError(err)
	}

	// count movies

	nMovies, err = r.MoviesCount(ctx)
	require.NoError(err)
	require.Equal(len(movies), nMovies)
}

func TestMovieCreate(t *testing.T) {
	require := require.New(t)

	teardown := setup()
	t.Cleanup(teardown)

	r := repo.NewRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "email"}
	err := r.UserCreate(ctx, user)
	require.NoError(err)

	movie := &models.Film{
		Title:        "movie",
		DateReleased: testutils.Date(2000, 1, 1),
	}

	// first there's no movie

	nMovies, err := r.MoviesCount(ctx)
	require.NoError(err)
	require.Equal(0, nMovies)

	// create a movie

	err = r.MovieCreate(
		ctx,
		user.ID,
		movie,
	)
	require.NoError(err)

	// fetch the movie

	fetchedMovie, err := r.MovieGet(ctx, movie.ID)
	require.NoError(err)

	testutils.SetTimeLocation(
		&movie.DateReleased,
		fetchedMovie.DateReleased.Location(),
	)

	require.Equal(movie, fetchedMovie)
}

func TestMovieUpdate(t *testing.T) {
	require := require.New(t)

	teardown := setup()
	t.Cleanup(teardown)

	r := repo.NewRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "email"}
	err := r.UserCreate(ctx, user)
	require.NoError(err)

	movie := &models.Film{
		Title:        "movie",
		DateReleased: testutils.Date(2000, 1, 1),
	}

	newTitle := "new movie title"
	newDescriptions := "new movie descriptions"
	newDateReleased := testutils.Date(1990, 1, 1)
	newDuration := 3600 * 2

	updateColumns := map[string]any{
		models.FilmColumns.Title:        newTitle,
		models.FilmColumns.Descriptions: newDescriptions,
		models.FilmColumns.DateReleased: newDateReleased,
		models.FilmColumns.Duration:     newDuration,
	}

	// first there's no movie

	err = r.MovieUpdate(ctx, movie.ID, user.ID, updateColumns)
	require.Equal(repo.ErrNoRecord, err)

	// add movie

	err = r.MovieCreate(ctx, user.ID, movie)
	require.NoError(err)

	// update movie

	outdatedMovie := movie

	err = r.MovieUpdate(ctx, movie.ID, user.ID, updateColumns)
	require.NoError(err)

	// fetch the updated movie

	fetchedUpdatedMovie, err := r.MovieGet(ctx, movie.ID)
	require.NoError(err)

	testutils.SetTimeLocation(
		&newDateReleased,
		fetchedUpdatedMovie.DateReleased.Location(),
	)

	require.Equal(
		&models.Film{
			ID:            movie.ID,
			Title:         newTitle,
			Descriptions:  null.StringFrom(newDescriptions),
			DateReleased:  newDateReleased,
			Duration:      null.IntFrom(newDuration),
			ContributedBy: user.ID,
			// let's imagine we don't care about ContributedAt field for a moment :|
			ContributedAt: fetchedUpdatedMovie.ContributedAt,
			Invalidation:  outdatedMovie.Invalidation,
		},
		fetchedUpdatedMovie,
	)

	// check outdated movie audited

	audits, err := r.MovieAuditsGetAll(
		ctx,
		movie.ID,
		query.SortOrderOptions{
			Offset:    0,
			Limit:     math.MaxInt,
			SortOrder: "desc",
		},
	)
	require.NoError(err)
	require.Equal(1, len(audits))

	testutils.SetTimeLocation(
		&outdatedMovie.DateReleased,
		audits[0].DateReleased.Location(),
	)

	require.Equal(
		outdatedMovie,
		&models.Film{
			ID:            audits[0].ID,
			Title:         audits[0].Title,
			Descriptions:  audits[0].Descriptions,
			DateReleased:  audits[0].DateReleased,
			Duration:      audits[0].Duration,
			ContributedBy: audits[0].ContributedBy,
			ContributedAt: audits[0].ContributedAt,
			Invalidation:  audits[0].Invalidation,
		},
	)
}

////////////////////////////////////////////////////////////////////////////////

func TestMovieAuditsGetAll(t *testing.T) {
	require := require.New(t)

	teardown := setup()
	t.Cleanup(teardown)

	r := repo.NewRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "email"}
	err := r.UserCreate(ctx, user)
	require.NoError(err)

	movie := &models.Film{
		Title:        "movie",
		DateReleased: testutils.Date(2000, 1, 1),
	}
	err = r.MovieCreate(ctx, user.ID, movie)
	require.NoError(err)

	queryOptions := query.SortOrderOptions{
		Offset:    0,
		Limit:     math.MaxInt,
		SortOrder: "desc",
	}

	// first there's no audits

	audits, err := r.MovieAuditsGetAll(ctx, movie.ID, queryOptions)
	require.NoError(err)
	require.Equal(0, len(audits))

	// audits for movie

	movieNewUpdates := []map[string]any{
		{models.FilmColumns.Title: "nt1"},
		{models.FilmColumns.Descriptions: "nd2"},
		{
			models.FilmColumns.DateReleased: testutils.Date(2002, 11, 13),
		},

		{
			models.FilmColumns.Duration: 2 * 60,
		},

		{
			models.FilmColumns.Title:        "nt5",
			models.FilmColumns.Descriptions: "nd5",
			models.FilmColumns.DateReleased: testutils.Date(2004, 3, 6),
			models.FilmColumns.Duration:     12 * 60,
		},
		{models.FilmColumns.Title: "nt6"},
	}

	for _, mnu := range movieNewUpdates {
		err = r.MovieUpdate(
			ctx,
			movie.ID,
			user.ID,
			mnu,
		)
		require.NoError(err)
	}

	// fetch audits

	audits, err = r.MovieAuditsGetAll(ctx, movie.ID, queryOptions)
	require.NoError(err)
	require.Equal(len(movieNewUpdates), len(audits))

	testutils.SetTimeLocation(
		&movie.DateReleased,
		audits[len(audits)-1].DateReleased.Location(),
	)

	require.Equal(
		movie,
		&models.Film{
			ID:            audits[len(audits)-1].ID,
			Title:         audits[len(audits)-1].Title,
			Descriptions:  audits[len(audits)-1].Descriptions,
			DateReleased:  audits[len(audits)-1].DateReleased,
			Duration:      audits[len(audits)-1].Duration,
			SeriesID:      audits[len(audits)-1].SeriesID,
			SeasonNumber:  audits[len(audits)-1].SeasonNumber,
			EpisodeNumber: audits[len(audits)-1].EpisodeNumber,
			ContributedBy: audits[len(audits)-1].ContributedBy,
			ContributedAt: audits[len(audits)-1].ContributedAt,
			Invalidation:  audits[len(audits)-1].Invalidation,
		},
	)

	for i, a := range audits[:len(audits)-1] {
		auditedMovie := &models.FilmsAudit{}
		auditedMovie.ID = movie.ID

		snu := movieNewUpdates[len(movieNewUpdates)-i-2]

		if titleString, exists := snu[models.FilmColumns.Title]; exists {
			title, isString := titleString.(string)
			require.True(isString)
			auditedMovie.Title = title
		} else {
			auditedMovie.Title = audits[i+1].Title
		}
		if descriptionsString, exists := snu[models.FilmColumns.Descriptions]; exists {
			descriptions, isString := descriptionsString.(string)
			require.True(isString)
			auditedMovie.Descriptions = null.StringFrom(descriptions)
		} else {
			auditedMovie.Descriptions = audits[i+1].Descriptions
		}
		if dateReleasedTime, exists := snu[models.FilmColumns.DateReleased]; exists {
			dateReleased, isTime := dateReleasedTime.(time.Time)
			require.True(isTime)
			auditedMovie.DateReleased = dateReleased
		} else {
			auditedMovie.DateReleased = audits[i+1].DateReleased
		}
		if durationInt, exists := snu[models.FilmColumns.Duration]; exists {
			duration, isInt := durationInt.(int)
			require.True(isInt)
			auditedMovie.Duration = null.IntFrom(duration)
		} else {
			auditedMovie.Duration = audits[i+1].Duration
		}
		auditedMovie.SeriesID = null.Int{}
		auditedMovie.SeasonNumber = null.Int{}
		auditedMovie.EpisodeNumber = null.Int{}
		auditedMovie.Invalidation = null.String{}
		auditedMovie.ContributedBy = user.ID
		auditedMovie.ContributedAt = a.ContributedAt

		testutils.SetTimeLocation(
			&auditedMovie.DateReleased,
			a.DateReleased.Location(),
		)

		require.Equal(auditedMovie, a)

	}
}

func TestMovieAuditsCount(t *testing.T) {
	require := require.New(t)

	teardown := setup()
	t.Cleanup(teardown)

	r := repo.NewRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "email"}
	err := r.UserCreate(ctx, user)
	require.NoError(err)

	movie := &models.Film{Title: "movie"}
	err = r.MovieCreate(
		ctx,
		user.ID,
		movie,
	)
	require.NoError(err)

	// first there's no audits

	auditsCount, err := r.MovieAuditsCount(ctx, movie.ID)
	require.NoError(err)
	require.Equal(0, auditsCount)

	// audits for movie

	movieNewVersions := []*models.Film{
		{Title: "a1"},
		{Title: "a2"},
		{Title: "a3"},
		{Title: "a4"},
		{Title: "a5"},
	}

	for _, mnv := range movieNewVersions {
		err = r.MovieUpdate(
			ctx,
			movie.ID,
			user.ID,
			map[string]any{
				models.FilmColumns.Title: mnv.Title,
			},
		)
		require.NoError(err)
	}

	// count audits

	auditsCount, err = r.MovieAuditsCount(ctx, movie.ID)
	require.NoError(err)
	require.Equal(len(movieNewVersions), auditsCount)
}
