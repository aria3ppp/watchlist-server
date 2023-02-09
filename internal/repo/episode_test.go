package repo_test

import (
	"context"
	"math"
	"testing"

	"github.com/aria3ppp/watchlist-server/internal/models"
	"github.com/aria3ppp/watchlist-server/internal/query"
	"github.com/aria3ppp/watchlist-server/internal/repo"
	"github.com/aria3ppp/watchlist-server/internal/testutils"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/null/v8"
)

// func TestEpisodeGetByID(t *testing.T) {
// 	require := require.New(t)

// 	teardown := setup()
// 	t.Cleanup(teardown)

// 	r := repo.NewRepo(db)
// 	ctx := context.Background()

// 	user := &models.User{Email: "email"}
// 	err = r.UserCreate(ctx, user)
// 	require.NoError(err)
// 	series := &models.Series{Title: "series"}
// 	err = r.SeriesCreate(ctx, user.ID, series)
// 	require.NoError(err)

// 	// first there's no episode

// 	fetchedEpisode, err := r.EpisodeGetByID(ctx, 1)
// 	require.Equal(repo.ErrNoRecord, err)
// 	require.Nil(fetchedEpisode)

// 	// insert an episode

// 	episode := &models.Film{
// 		Title:        "episode",
// 		DateReleased: testutils.Date(2000,1,1),
// 	}

// 	err = r.EpisodePut(ctx, series.ID, 1, 1, user.ID, episode)
// 	require.NoError(err)

// 	// fetch the episode

// 	fetchedEpisode, err = r.EpisodeGetByID(ctx, episode.ID)
// 	require.NoError(err)

// 	require.Equal(episode, fetchedEpisode)
// }

func TestEpisodeGet(t *testing.T) {
	require := require.New(t)

	teardown := setup()
	t.Cleanup(teardown)

	r := repo.NewRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "email"}
	err := r.UserCreate(ctx, user)
	require.NoError(err)
	series := &models.Series{Title: "series"}
	err = r.SeriesCreate(ctx, user.ID, series)
	require.NoError(err)

	// first there's no episode

	fetchedEpisode, err := r.EpisodeGet(ctx, series.ID, 1, 1)
	require.Equal(repo.ErrNoRecord, err)
	require.Nil(fetchedEpisode)

	// insert an episode

	episode := &models.Film{
		Title:        "episode",
		DateReleased: testutils.Date(2000, 1, 1),
	}

	seasonNumber := 1
	episodeNumber := 1

	err = r.EpisodePut(
		ctx,
		series.ID,
		seasonNumber,
		episodeNumber,
		user.ID,
		episode,
	)
	require.NoError(err)

	// fetch the episode

	fetchedEpisode, err = r.EpisodeGet(
		ctx,
		series.ID,
		seasonNumber,
		episodeNumber,
	)
	require.NoError(err)

	testutils.SetTimeLocation(
		&episode.DateReleased,
		fetchedEpisode.DateReleased.Location(),
	)

	require.Equal(episode, fetchedEpisode)
}

func TestEpisodesGetAllBySeries(t *testing.T) {
	require := require.New(t)

	teardown := setup()
	t.Cleanup(teardown)

	r := repo.NewRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "email"}
	err := r.UserCreate(ctx, user)
	require.NoError(err)
	series := &models.Series{Title: "series"}
	err = r.SeriesCreate(ctx, user.ID, series)
	require.NoError(err)

	// first there's no episode

	fetchedEpisode, err := r.EpisodesGetAllBySeries(
		ctx,
		series.ID,
		query.SortOrderOptions{
			SortOrder: "asc",
			Offset:    0,
			Limit:     math.MaxInt,
		},
	)
	require.NoError(err)
	require.Equal(0, len(fetchedEpisode))

	// insert episodes

	episodes := []*models.Film{
		{
			Title:        "e1",
			DateReleased: testutils.Date(2000, 1, 1),
		},
		{
			Title:        "e2",
			DateReleased: testutils.Date(2001, 1, 1),
		},
		{
			Title:        "e3",
			DateReleased: testutils.Date(2002, 1, 1),
		},
		{
			Title:        "e4",
			DateReleased: testutils.Date(2003, 1, 1),
		},
		{
			Title:        "e5",
			DateReleased: testutils.Date(2004, 1, 1),
		},
	}

	seasonNumber := 1

	for i, e := range episodes {
		err := r.EpisodePut(ctx, series.ID, seasonNumber, i+1, user.ID, e)
		require.NoError(err)
	}

	// fetch episodes

	fetchedEpisode, err = r.EpisodesGetAllBySeries(
		ctx,
		series.ID,
		query.SortOrderOptions{
			SortOrder: "asc",
			Offset:    0,
			Limit:     math.MaxInt,
		},
	)
	require.NoError(err)
	require.Equal(len(episodes), len(fetchedEpisode))

	// EpisodesGetAllBySeries records are order by id by default
	for i, fe := range fetchedEpisode {
		testutils.SetTimeLocation(
			&episodes[i].DateReleased,
			fe.DateReleased.Location(),
		)
		require.Equal(episodes[i], fe)
	}
}

func TestEpisodesGetAllBySeason(t *testing.T) {
	require := require.New(t)

	teardown := setup()
	t.Cleanup(teardown)

	r := repo.NewRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "email"}
	err := r.UserCreate(ctx, user)
	require.NoError(err)
	series := &models.Series{Title: "series"}
	err = r.SeriesCreate(ctx, user.ID, series)
	require.NoError(err)

	// first there's no episode

	seasonNumber := 1

	fetchedEpisodes, err := r.EpisodesGetAllBySeason(
		ctx,
		series.ID,
		seasonNumber,
		query.SortOrderOptions{
			SortOrder: "asc",
			Offset:    0,
			Limit:     math.MaxInt,
		},
	)
	require.NoError(err)
	require.Equal(0, len(fetchedEpisodes))

	// insert episodes

	episodes := []*models.Film{
		{
			Title:        "e1",
			DateReleased: testutils.Date(2000, 1, 1),
		},
		{
			Title:        "e2",
			DateReleased: testutils.Date(2001, 1, 1),
		},
		{
			Title:        "e3",
			DateReleased: testutils.Date(2002, 1, 1),
		},
		{
			Title:        "e4",
			DateReleased: testutils.Date(2003, 1, 1),
		},
		{
			Title:        "e5",
			DateReleased: testutils.Date(2004, 1, 1),
		},
	}

	for i, e := range episodes {
		err := r.EpisodePut(ctx, series.ID, seasonNumber, i+1, user.ID, e)
		require.NoError(err)
	}

	// fetch episodes

	fetchedEpisodes, err = r.EpisodesGetAllBySeason(
		ctx,
		series.ID,
		seasonNumber,
		query.SortOrderOptions{
			SortOrder: "asc",
			Offset:    0,
			Limit:     math.MaxInt,
		},
	)
	require.NoError(err)
	require.Equal(len(episodes), len(fetchedEpisodes))

	// EpisodesGetAllBySeries records are order by id by default
	for i, fe := range fetchedEpisodes {
		testutils.SetTimeLocation(
			&episodes[i].DateReleased,
			fe.DateReleased.Location(),
		)
		require.Equal(episodes[i], fe)
	}
}

////////////////////////////////////////////////////////////////////////////////

func TestEpisodesCountBySeries(t *testing.T) {
	require := require.New(t)

	teardown := setup()
	t.Cleanup(teardown)

	r := repo.NewRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "email"}
	err := r.UserCreate(ctx, user)
	require.NoError(err)
	series := &models.Series{Title: "series"}
	err = r.SeriesCreate(ctx, user.ID, series)
	require.NoError(err)

	// first there's no episode

	nEpisodes, err := r.EpisodesCountBySeries(ctx, series.ID)
	require.NoError(err)
	require.Equal(0, nEpisodes)

	// insert episodes

	episodes := []*models.Film{
		{
			Title:        "a1",
			DateReleased: testutils.Date(2000, 1, 1),
		},
		{
			Title:        "a2",
			DateReleased: testutils.Date(2001, 1, 1),
		},
		{
			Title:        "a3",
			DateReleased: testutils.Date(2002, 1, 1),
		},
		{
			Title:        "a4",
			DateReleased: testutils.Date(2003, 1, 1),
		},
		{
			Title:        "a5",
			DateReleased: testutils.Date(2004, 1, 1),
		},
	}

	seasonNumber := 1

	for i, e := range episodes {
		err := r.EpisodePut(ctx, series.ID, seasonNumber, i+1, user.ID, e)
		require.NoError(err)
	}

	// count episodes

	nEpisodes, err = r.EpisodesCountBySeries(ctx, series.ID)
	require.NoError(err)
	require.Equal(len(episodes), nEpisodes)
}

func TestEpisodesCountBySeason(t *testing.T) {
	require := require.New(t)

	teardown := setup()
	t.Cleanup(teardown)

	r := repo.NewRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "email"}
	err := r.UserCreate(ctx, user)
	require.NoError(err)
	series := &models.Series{Title: "series"}
	err = r.SeriesCreate(ctx, user.ID, series)
	require.NoError(err)
	seasonNumber := 1

	// first there's no episode

	nEpisodes, err := r.EpisodesCountBySeason(ctx, series.ID, seasonNumber)
	require.NoError(err)
	require.Equal(0, nEpisodes)

	// insert episodes

	episodes := []*models.Film{
		{
			Title:        "a1",
			DateReleased: testutils.Date(2000, 1, 1),
		},
		{
			Title:        "a2",
			DateReleased: testutils.Date(2001, 1, 1),
		},
		{
			Title:        "a3",
			DateReleased: testutils.Date(2002, 1, 1),
		},
		{
			Title:        "a4",
			DateReleased: testutils.Date(2003, 1, 1),
		},
		{
			Title:        "a5",
			DateReleased: testutils.Date(2004, 1, 1),
		},
	}

	for i, e := range episodes {
		err := r.EpisodePut(ctx, series.ID, seasonNumber, i+1, user.ID, e)
		require.NoError(err)
	}

	// count episodes

	nEpisodes, err = r.EpisodesCountBySeason(ctx, series.ID, seasonNumber)
	require.NoError(err)
	require.Equal(len(episodes), nEpisodes)
}

////////////////////////////////////////////////////////////////////////////////

func TestEpisodePut(t *testing.T) {
	require := require.New(t)

	teardown := setup()
	t.Cleanup(teardown)

	r := repo.NewRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "email"}
	err := r.UserCreate(ctx, user)
	require.NoError(err)
	series := &models.Series{Title: "series"}
	err = r.SeriesCreate(ctx, user.ID, series)
	require.NoError(err)
	episode := &models.Film{
		Title:        "episode",
		DateReleased: testutils.Date(2000, 1, 1),
	}

	seasonNumber := 1
	episodeNumber := 1

	// first there's no episode

	nEpisodes, err := r.EpisodesCountBySeason(ctx, series.ID, seasonNumber)
	require.NoError(err)
	require.Equal(0, nEpisodes)

	// put episode

	err = r.EpisodePut(
		ctx,
		series.ID,
		seasonNumber,
		episodeNumber,
		user.ID,
		episode,
	)
	require.NoError(err)

	// fetch the episode

	fetchedEpisodes, err := r.EpisodesGetAllBySeason(
		ctx,
		series.ID,
		seasonNumber,
		query.SortOrderOptions{
			SortOrder: "asc",
			Offset:    0,
			Limit:     math.MaxInt,
		},
	)
	require.NoError(err)
	require.Equal(1, len(fetchedEpisodes))

	testutils.SetTimeLocation(
		&episode.DateReleased,
		fetchedEpisodes[0].DateReleased.Location(),
	)

	require.Equal(episode, fetchedEpisodes[0])

	// replace current episode

	replacedEpisode := fetchedEpisodes[0]

	newEpisode := &models.Film{
		Title:        "new episode title",
		Descriptions: null.StringFrom("new episode descriptions"),
		DateReleased: testutils.Date(1990, 1, 1),
		Duration:     null.IntFrom(3600 * 2),
	}

	err = r.EpisodePut(
		ctx,
		series.ID,
		seasonNumber,
		episodeNumber,
		user.ID,
		newEpisode,
	)
	require.NoError(err)

	fetchedEpisodes, err = r.EpisodesGetAllBySeason(
		ctx,
		series.ID,
		seasonNumber,
		query.SortOrderOptions{
			SortOrder: "asc",
			Offset:    0,
			Limit:     math.MaxInt,
		},
	)
	require.NoError(err)
	require.Equal(1, len(fetchedEpisodes))

	testutils.SetTimeLocation(
		&newEpisode.DateReleased,
		fetchedEpisodes[0].DateReleased.Location(),
	)

	require.Equal(
		&models.Film{
			ID:            replacedEpisode.ID,
			Title:         newEpisode.Title,
			Descriptions:  newEpisode.Descriptions,
			DateReleased:  newEpisode.DateReleased,
			Duration:      newEpisode.Duration,
			SeriesID:      replacedEpisode.SeriesID,
			SeasonNumber:  replacedEpisode.SeasonNumber,
			EpisodeNumber: replacedEpisode.EpisodeNumber,
			ContributedBy: user.ID,
			ContributedAt: newEpisode.ContributedAt,
			Invalidation:  replacedEpisode.Invalidation,
		},
		fetchedEpisodes[0],
	)

	// check replaced episode audited

	audits, err := r.EpisodeAuditsGetAll(
		ctx,
		series.ID,
		seasonNumber,
		episodeNumber,
		query.SortOrderOptions{
			SortOrder: "desc",
			Offset:    0,
			Limit:     math.MaxInt,
		},
	)
	require.NoError(err)
	require.Equal(1, len(audits))

	require.Equal(
		replacedEpisode,
		&models.Film{
			ID:            audits[0].ID,
			Title:         audits[0].Title,
			Descriptions:  audits[0].Descriptions,
			DateReleased:  audits[0].DateReleased,
			Duration:      audits[0].Duration,
			SeriesID:      audits[0].SeriesID,
			SeasonNumber:  audits[0].SeasonNumber,
			EpisodeNumber: audits[0].EpisodeNumber,
			ContributedBy: audits[0].ContributedBy,
			ContributedAt: audits[0].ContributedAt,
			Invalidation:  audits[0].Invalidation,
		},
	)
}

func TestEpisodeUpdate(t *testing.T) {
	require := require.New(t)

	teardown := setup()
	t.Cleanup(teardown)

	r := repo.NewRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "email"}
	err := r.UserCreate(ctx, user)
	require.NoError(err)
	series := &models.Series{Title: "series"}
	err = r.SeriesCreate(ctx, user.ID, series)
	require.NoError(err)
	episode := &models.Film{
		Title:        "episode",
		DateReleased: testutils.Date(2000, 1, 1),
	}

	seasonNumber := 1
	episodeNumber := 1

	newTitle := "new episode title"
	newDescriptions := "new episode descriptions"
	newDateReleased := testutils.Date(1990, 1, 1)
	newDuration := 3600 * 2

	updateColumns := map[string]any{
		models.FilmColumns.Title:        newTitle,
		models.FilmColumns.Descriptions: newDescriptions,
		models.FilmColumns.DateReleased: newDateReleased,
		models.FilmColumns.Duration:     newDuration,
	}

	// first there's no episode

	err = r.EpisodeUpdate(
		ctx,
		series.ID,
		seasonNumber,
		episodeNumber,
		user.ID,
		updateColumns,
	)
	require.Equal(repo.ErrNoRecord, err)

	// add episode

	err = r.EpisodePut(
		ctx,
		series.ID,
		seasonNumber,
		episodeNumber,
		user.ID,
		episode,
	)
	require.NoError(err)

	// update episode

	outdatedEpisode := episode

	err = r.EpisodeUpdate(
		ctx,
		series.ID,
		seasonNumber,
		episodeNumber,
		user.ID,
		updateColumns,
	)
	require.NoError(err)

	// fetch the updated episode

	fetchedUpdatedEpisode, err := r.EpisodeGet(
		ctx,
		series.ID,
		seasonNumber,
		episodeNumber,
	)
	require.NoError(err)

	testutils.SetTimeLocation(
		&newDateReleased,
		fetchedUpdatedEpisode.DateReleased.Location(),
	)

	require.Equal(
		&models.Film{
			ID:            outdatedEpisode.ID,
			Title:         newTitle,
			Descriptions:  null.StringFrom(newDescriptions),
			DateReleased:  newDateReleased,
			Duration:      null.IntFrom(newDuration),
			SeriesID:      episode.SeriesID,
			SeasonNumber:  episode.SeasonNumber,
			EpisodeNumber: episode.EpisodeNumber,
			ContributedBy: user.ID,
			// let's imagine we don't care about ContributedAt field for a moment :|
			ContributedAt: fetchedUpdatedEpisode.ContributedAt,
			Invalidation:  outdatedEpisode.Invalidation,
		},
		fetchedUpdatedEpisode,
	)

	// check outdated episode audited

	audits, err := r.EpisodeAuditsGetAll(
		ctx,
		series.ID,
		seasonNumber,
		episodeNumber,
		query.SortOrderOptions{
			SortOrder: "desc",
			Offset:    0,
			Limit:     math.MaxInt,
		},
	)
	require.NoError(err)
	require.Equal(1, len(audits))

	testutils.SetTimeLocation(
		&outdatedEpisode.DateReleased,
		audits[0].DateReleased.Location(),
	)

	require.Equal(
		outdatedEpisode,
		&models.Film{
			ID:            audits[0].ID,
			Title:         audits[0].Title,
			Descriptions:  audits[0].Descriptions,
			DateReleased:  audits[0].DateReleased,
			Duration:      audits[0].Duration,
			SeriesID:      audits[0].SeriesID,
			SeasonNumber:  audits[0].SeasonNumber,
			EpisodeNumber: audits[0].EpisodeNumber,
			ContributedBy: audits[0].ContributedBy,
			ContributedAt: audits[0].ContributedAt,
			Invalidation:  audits[0].Invalidation,
		},
	)
}

func TestEpisodesInvalidateAllBySeason(t *testing.T) {
	require := require.New(t)

	teardown := setup()
	t.Cleanup(teardown)

	r := repo.NewRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "email"}
	err := r.UserCreate(ctx, user)
	require.NoError(err)
	series := &models.Series{Title: "series"}
	err = r.SeriesCreate(ctx, user.ID, series)
	require.NoError(err)

	seasonNumber := 1

	// first there's no episode

	invalidation := "invalidation"
	err = r.EpisodesInvalidateAllBySeason(
		ctx,
		series.ID,
		seasonNumber,
		user.ID,
		invalidation,
	)
	require.Equal(repo.ErrNoRecord, err)

	// add episodes

	episodes := []*models.Film{
		{
			Title:        "e1",
			DateReleased: testutils.Date(2000, 1, 1),
		},
		{
			Title:        "e2",
			DateReleased: testutils.Date(2001, 1, 1),
		},
		{
			Title:        "e3",
			DateReleased: testutils.Date(2002, 1, 1),
		},
		{
			Title:        "e4",
			DateReleased: testutils.Date(2003, 1, 1),
		},
		{
			Title:        "e5",
			DateReleased: testutils.Date(2004, 1, 1),
		},
	}

	for i, e := range episodes {
		err := r.EpisodePut(ctx, series.ID, seasonNumber, i+1, user.ID, e)
		require.NoError(err)
	}

	episodesBeforeInvalidation := episodes

	// invalidate episodes

	err = r.EpisodesInvalidateAllBySeason(
		ctx,
		series.ID,
		seasonNumber,
		user.ID,
		invalidation,
	)
	require.NoError(err)

	// check invalidated

	invalidatedEpisodes, err := r.EpisodesGetAllBySeason(
		ctx,
		series.ID,
		seasonNumber,
		query.SortOrderOptions{
			SortOrder: "asc",
			Offset:    0,
			Limit:     math.MaxInt,
		},
	)
	require.NoError(err)
	for _, ie := range invalidatedEpisodes {
		require.Equal(null.StringFrom(invalidation), ie.Invalidation)
	}

	// check invalidated episodes audited

	audits, err := r.EpisodesAuditsGetAllBySeason(
		ctx,
		series.ID,
		seasonNumber,
		0,
		math.MaxInt,
	)
	require.NoError(err)
	require.Equal(len(episodesBeforeInvalidation), len(audits))

	// as audits are ordered by ContributedBy DESC so basically they are exactly in opposite order
	for i, de := range episodesBeforeInvalidation {
		testutils.SetTimeLocation(
			&de.DateReleased,
			audits[len(audits)-i-1].DateReleased.Location(),
		)
		require.Equal(
			de,
			&models.Film{
				ID:            audits[len(audits)-i-1].ID,
				Title:         audits[len(audits)-i-1].Title,
				Descriptions:  audits[len(audits)-i-1].Descriptions,
				DateReleased:  audits[len(audits)-i-1].DateReleased,
				Duration:      audits[len(audits)-i-1].Duration,
				SeriesID:      audits[len(audits)-i-1].SeriesID,
				SeasonNumber:  audits[len(audits)-i-1].SeasonNumber,
				EpisodeNumber: audits[len(audits)-i-1].EpisodeNumber,
				ContributedBy: audits[len(audits)-i-1].ContributedBy,
				ContributedAt: audits[len(audits)-i-1].ContributedAt,
				Invalidation:  audits[len(audits)-i-1].Invalidation,
			},
		)
	}
}

////////////////////////////////////////////////////////////////////////////////

func TestEpisodeAuditsGetAll(t *testing.T) {
	require := require.New(t)

	teardown := setup()
	t.Cleanup(teardown)

	r := repo.NewRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "email"}
	err := r.UserCreate(ctx, user)
	require.NoError(err)
	series := &models.Series{Title: "series"}
	err = r.SeriesCreate(ctx, user.ID, series)
	require.NoError(err)

	seasonNumber := 1
	episodeNumber := 1

	episode := &models.Film{
		Title:        "episode",
		DateReleased: testutils.Date(2000, 1, 1),
	}
	err = r.EpisodePut(
		ctx,
		series.ID,
		seasonNumber,
		episodeNumber,
		user.ID,
		episode,
	)
	require.NoError(err)

	// first there's no audits

	audits, err := r.EpisodeAuditsGetAll(
		ctx,
		series.ID,
		seasonNumber,
		episodeNumber,
		query.SortOrderOptions{
			SortOrder: "desc",
			Offset:    0,
			Limit:     math.MaxInt,
		},
	)
	require.NoError(err)
	require.Equal(0, len(audits))

	// audits for episode

	episodeNewVersions := []*models.Film{
		{
			Title:        "a1",
			DateReleased: testutils.Date(2000, 1, 1),
		},
		{
			Title:        "a2",
			DateReleased: testutils.Date(2001, 1, 1),
		},
		{
			Title:        "a3",
			DateReleased: testutils.Date(2002, 1, 1),
		},
		{
			Title:        "a4",
			DateReleased: testutils.Date(2003, 1, 1),
		},
		{
			Title:        "a5",
			DateReleased: testutils.Date(2004, 1, 1),
		},
	}

	for _, env := range episodeNewVersions {
		err := r.EpisodePut(
			ctx,
			series.ID,
			seasonNumber,
			episodeNumber,
			user.ID,
			env,
		)
		require.NoError(err)
	}

	// fetch audits

	audits, err = r.EpisodeAuditsGetAll(
		ctx,
		series.ID,
		seasonNumber,
		episodeNumber,
		query.SortOrderOptions{
			SortOrder: "desc",
			Offset:    0,
			Limit:     math.MaxInt,
		},
	)
	require.NoError(err)
	require.Equal(len(episodeNewVersions), len(audits))

	testutils.SetTimeLocation(
		&episode.DateReleased,
		audits[len(audits)-1].DateReleased.Location(),
	)

	require.Equal(
		episode,
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

	for i, env := range episodeNewVersions[:len(episodeNewVersions)-1] {
		testutils.SetTimeLocation(
			&env.DateReleased,
			audits[len(audits)-i-2].DateReleased.Location(),
		)
		require.Equal(
			env,
			&models.Film{
				ID:            audits[len(audits)-i-2].ID,
				Title:         audits[len(audits)-i-2].Title,
				Descriptions:  audits[len(audits)-i-2].Descriptions,
				DateReleased:  audits[len(audits)-i-2].DateReleased,
				Duration:      audits[len(audits)-i-2].Duration,
				SeriesID:      audits[len(audits)-i-2].SeriesID,
				SeasonNumber:  audits[len(audits)-i-2].SeasonNumber,
				EpisodeNumber: audits[len(audits)-i-2].EpisodeNumber,
				ContributedBy: audits[len(audits)-i-2].ContributedBy,
				ContributedAt: audits[len(audits)-i-2].ContributedAt,
				Invalidation:  audits[len(audits)-i-2].Invalidation,
			},
		)
	}
}

func TestEpisodeAuditsCount(t *testing.T) {
	require := require.New(t)

	teardown := setup()
	t.Cleanup(teardown)

	r := repo.NewRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "email"}
	err := r.UserCreate(ctx, user)
	require.NoError(err)
	series := &models.Series{Title: "series"}
	err = r.SeriesCreate(ctx, user.ID, series)
	require.NoError(err)

	seasonNumber := 1
	episodeNumber := 1

	episode := &models.Film{Title: "episode"}
	err = r.EpisodePut(
		ctx,
		series.ID,
		seasonNumber,
		episodeNumber,
		user.ID,
		episode,
	)
	require.NoError(err)

	// first there's no audits

	auditsCount, err := r.EpisodeAuditsCount(
		ctx,
		series.ID,
		seasonNumber,
		episodeNumber,
	)
	require.NoError(err)
	require.Equal(0, auditsCount)

	// audits for episode

	episodeNewVersions := []*models.Film{
		{
			Title:        "a1",
			DateReleased: testutils.Date(2000, 1, 1),
		},
		{
			Title:        "a2",
			DateReleased: testutils.Date(2001, 1, 1),
		},
		{
			Title:        "a3",
			DateReleased: testutils.Date(2002, 1, 1),
		},
		{
			Title:        "a4",
			DateReleased: testutils.Date(2003, 1, 1),
		},
		{
			Title:        "a5",
			DateReleased: testutils.Date(2004, 1, 1),
		},
	}

	for _, env := range episodeNewVersions {
		err := r.EpisodePut(
			ctx,
			series.ID,
			seasonNumber,
			episodeNumber,
			user.ID,
			env,
		)
		require.NoError(err)
	}

	// count audits

	auditsCount, err = r.EpisodeAuditsCount(
		ctx,
		series.ID,
		seasonNumber,
		episodeNumber,
	)
	require.NoError(err)
	require.Equal(len(episodeNewVersions), auditsCount)
}

func TestEpisodesAuditsGetAllBySeason(t *testing.T) {
	require := require.New(t)

	teardown := setup()
	t.Cleanup(teardown)

	r := repo.NewRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "email"}
	err := r.UserCreate(ctx, user)
	require.NoError(err)
	series := &models.Series{Title: "series"}
	err = r.SeriesCreate(ctx, user.ID, series)
	require.NoError(err)

	seasonNumber := 1
	episodeNumber := 1

	episode := &models.Film{
		Title:        "episode",
		DateReleased: testutils.Date(2000, 1, 1),
	}
	err = r.EpisodePut(
		ctx,
		series.ID,
		seasonNumber,
		episodeNumber,
		user.ID,
		episode,
	)
	require.NoError(err)

	// first there's no audits

	audits, err := r.EpisodesAuditsGetAllBySeason(
		ctx,
		series.ID,
		seasonNumber,
		0,
		math.MaxInt,
	)
	require.NoError(err)
	require.Equal(0, len(audits))

	// audits for episode

	episodeNewVersions := []*models.Film{
		{
			Title:        "a1",
			DateReleased: testutils.Date(2000, 1, 1),
		},
		{
			Title:        "a2",
			DateReleased: testutils.Date(2001, 1, 1),
		},
		{
			Title:        "a3",
			DateReleased: testutils.Date(2002, 1, 1),
		},
		{
			Title:        "a4",
			DateReleased: testutils.Date(2003, 1, 1),
		},
		{
			Title:        "a5",
			DateReleased: testutils.Date(2004, 1, 1),
		},
	}

	for _, env := range episodeNewVersions {
		err := r.EpisodePut(
			ctx,
			series.ID,
			seasonNumber,
			episodeNumber,
			user.ID,
			env,
		)
		require.NoError(err)
	}

	// fetch audits

	audits, err = r.EpisodesAuditsGetAllBySeason(
		ctx,
		series.ID,
		seasonNumber,
		0,
		math.MaxInt,
	)
	require.NoError(err)
	require.Equal(len(episodeNewVersions), len(audits))

	testutils.SetTimeLocation(
		&episode.DateReleased,
		audits[len(audits)-1].DateReleased.Location(),
	)

	require.Equal(
		episode,
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

	for i, env := range episodeNewVersions[:len(episodeNewVersions)-1] {
		testutils.SetTimeLocation(
			&env.DateReleased,
			audits[len(audits)-i-2].DateReleased.Location(),
		)
		require.Equal(
			env,
			&models.Film{
				ID:            audits[len(audits)-i-2].ID,
				Title:         audits[len(audits)-i-2].Title,
				Descriptions:  audits[len(audits)-i-2].Descriptions,
				DateReleased:  audits[len(audits)-i-2].DateReleased,
				Duration:      audits[len(audits)-i-2].Duration,
				SeriesID:      audits[len(audits)-i-2].SeriesID,
				SeasonNumber:  audits[len(audits)-i-2].SeasonNumber,
				EpisodeNumber: audits[len(audits)-i-2].EpisodeNumber,
				ContributedBy: audits[len(audits)-i-2].ContributedBy,
				ContributedAt: audits[len(audits)-i-2].ContributedAt,
				Invalidation:  audits[len(audits)-i-2].Invalidation,
			},
		)
	}
}

func TestEpisodesAuditsCountBySeason(t *testing.T) {
	require := require.New(t)

	teardown := setup()
	t.Cleanup(teardown)

	r := repo.NewRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "email"}
	err := r.UserCreate(ctx, user)
	require.NoError(err)
	series := &models.Series{Title: "series"}
	err = r.SeriesCreate(ctx, user.ID, series)
	require.NoError(err)

	seasonNumber := 1
	episodeNumber := 1

	episode := &models.Film{Title: "episode"}
	err = r.EpisodePut(
		ctx,
		series.ID,
		seasonNumber,
		episodeNumber,
		user.ID,
		episode,
	)
	require.NoError(err)

	// first there's no audits

	auditsCount, err := r.EpisodesAuditsCountBySeason(
		ctx,
		series.ID,
		seasonNumber,
	)
	require.NoError(err)
	require.Equal(0, auditsCount)

	// audits for episode

	episodeNewVersions := []*models.Film{
		{
			Title:        "a1",
			DateReleased: testutils.Date(2000, 1, 1),
		},
		{
			Title:        "a2",
			DateReleased: testutils.Date(2001, 1, 1),
		},
		{
			Title:        "a3",
			DateReleased: testutils.Date(2002, 1, 1),
		},
		{
			Title:        "a4",
			DateReleased: testutils.Date(2003, 1, 1),
		},
		{
			Title:        "a5",
			DateReleased: testutils.Date(2004, 1, 1),
		},
	}

	for _, env := range episodeNewVersions {
		err := r.EpisodePut(
			ctx,
			series.ID,
			seasonNumber,
			episodeNumber,
			user.ID,
			env,
		)
		require.NoError(err)
	}

	// count audits

	auditsCount, err = r.EpisodesAuditsCountBySeason(
		ctx,
		series.ID,
		seasonNumber,
	)
	require.NoError(err)
	require.Equal(len(episodeNewVersions), auditsCount)
}

func TestEpisodesAuditsGetAllBySeries(t *testing.T) {
	require := require.New(t)

	teardown := setup()
	t.Cleanup(teardown)

	r := repo.NewRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "email"}
	err := r.UserCreate(ctx, user)
	require.NoError(err)
	series := &models.Series{Title: "series"}
	err = r.SeriesCreate(ctx, user.ID, series)
	require.NoError(err)

	seasonNumber := 1
	episodeNumber := 1

	episode := &models.Film{
		Title:        "episode",
		DateReleased: testutils.Date(2000, 1, 1),
	}
	err = r.EpisodePut(
		ctx,
		series.ID,
		seasonNumber,
		episodeNumber,
		user.ID,
		episode,
	)
	require.NoError(err)

	// first there's no audits

	audits, err := r.EpisodesAuditsGetAllBySeries(
		ctx,
		series.ID,
		0,
		math.MaxInt,
	)
	require.NoError(err)
	require.Equal(0, len(audits))

	// audits for episode

	episodeNewVersions := []*models.Film{
		{
			Title:        "a1",
			DateReleased: testutils.Date(2000, 1, 1),
		},
		{
			Title:        "a2",
			DateReleased: testutils.Date(2001, 1, 1),
		},
		{
			Title:        "a3",
			DateReleased: testutils.Date(2002, 1, 1),
		},
		{
			Title:        "a4",
			DateReleased: testutils.Date(2003, 1, 1),
		},
		{
			Title:        "a5",
			DateReleased: testutils.Date(2004, 1, 1),
		},
	}

	for _, env := range episodeNewVersions {
		err := r.EpisodePut(
			ctx,
			series.ID,
			seasonNumber,
			episodeNumber,
			user.ID,
			env,
		)
		require.NoError(err)
	}

	// fetch audits

	audits, err = r.EpisodesAuditsGetAllBySeries(
		ctx,
		series.ID,
		0,
		math.MaxInt,
	)
	require.NoError(err)
	require.Equal(len(episodeNewVersions), len(audits))

	testutils.SetTimeLocation(
		&episode.DateReleased,
		audits[len(audits)-1].DateReleased.Location(),
	)

	require.Equal(
		episode,
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

	for i, env := range episodeNewVersions[:len(episodeNewVersions)-1] {
		testutils.SetTimeLocation(
			&env.DateReleased,
			audits[len(audits)-i-2].DateReleased.Location(),
		)
		require.Equal(
			env,
			&models.Film{
				ID:            audits[len(audits)-i-2].ID,
				Title:         audits[len(audits)-i-2].Title,
				Descriptions:  audits[len(audits)-i-2].Descriptions,
				DateReleased:  audits[len(audits)-i-2].DateReleased,
				Duration:      audits[len(audits)-i-2].Duration,
				SeriesID:      audits[len(audits)-i-2].SeriesID,
				SeasonNumber:  audits[len(audits)-i-2].SeasonNumber,
				EpisodeNumber: audits[len(audits)-i-2].EpisodeNumber,
				ContributedBy: audits[len(audits)-i-2].ContributedBy,
				ContributedAt: audits[len(audits)-i-2].ContributedAt,
				Invalidation:  audits[len(audits)-i-2].Invalidation,
			},
		)
	}
}

func TestEpisodesAuditsCountBySeries(t *testing.T) {
	require := require.New(t)

	teardown := setup()
	t.Cleanup(teardown)

	r := repo.NewRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "email"}
	err := r.UserCreate(ctx, user)
	require.NoError(err)
	series := &models.Series{Title: "series"}
	err = r.SeriesCreate(ctx, user.ID, series)
	require.NoError(err)

	seasonNumber := 1
	episodeNumber := 1

	episode := &models.Film{Title: "episode"}
	err = r.EpisodePut(
		ctx,
		series.ID,
		seasonNumber,
		episodeNumber,
		user.ID,
		episode,
	)
	require.NoError(err)

	// first there's no audits

	auditsCount, err := r.EpisodesAuditsCountBySeries(ctx, series.ID)
	require.NoError(err)
	require.Equal(0, auditsCount)

	// audits for episode

	episodeNewVersions := []*models.Film{
		{
			Title:        "a1",
			DateReleased: testutils.Date(2000, 1, 1),
		},
		{
			Title:        "a2",
			DateReleased: testutils.Date(2001, 1, 1),
		},
		{
			Title:        "a3",
			DateReleased: testutils.Date(2002, 1, 1),
		},
		{
			Title:        "a4",
			DateReleased: testutils.Date(2003, 1, 1),
		},
		{
			Title:        "a5",
			DateReleased: testutils.Date(2004, 1, 1),
		},
	}

	for _, env := range episodeNewVersions {
		err := r.EpisodePut(
			ctx,
			series.ID,
			seasonNumber,
			episodeNumber,
			user.ID,
			env,
		)
		require.NoError(err)
	}

	// count audits

	auditsCount, err = r.EpisodesAuditsCountBySeries(ctx, series.ID)
	require.NoError(err)
	require.Equal(len(episodeNewVersions), auditsCount)
}
