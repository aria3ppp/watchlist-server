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

func TestSeriesGet(t *testing.T) {
	require := require.New(t)

	teardown := setup()
	t.Cleanup(teardown)

	r := repo.NewRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "email"}
	err := r.UserCreate(ctx, user)
	require.NoError(err)
	series := &models.Series{
		Title:       "serie",
		DateStarted: testutils.Date(2000, 1, 1),
	}

	// first there's no series

	fetchedSeries, err := r.SeriesGet(ctx, 1)
	require.Equal(repo.ErrNoRecord, err)
	require.Nil(fetchedSeries)

	// insert series

	err = r.SeriesCreate(ctx, user.ID, series)
	require.NoError(err)

	// fetch the series

	fetchedSeries, err = r.SeriesGet(ctx, series.ID)
	require.NoError(err)

	testutils.SetTimeLocation(
		&series.DateStarted,
		fetchedSeries.DateStarted.Location(),
	)
	testutils.SetTimeLocation(
		&series.DateEnded.Time,
		fetchedSeries.DateEnded.Time.Location(),
	)

	require.Equal(series, fetchedSeries)
}

func TestSeriesesGetAll(t *testing.T) {
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
		SortField: models.SeriesColumns.ID,
		SortOrder: "asc",
	}

	// first there's no serieses

	fetchedSerieses, err := r.SeriesesGetAll(ctx, queryOptions)
	require.NoError(err)
	require.Equal(0, len(fetchedSerieses))

	// insert serieses

	serieses := []*models.Series{
		{
			Title:       "s1",
			DateStarted: testutils.Date(2000, 1, 1),
		},
		{
			Title:       "s2",
			DateStarted: testutils.Date(2001, 1, 1),
		},
		{
			Title:       "s3",
			DateStarted: testutils.Date(2002, 1, 1),
		},
		{
			Title:       "s4",
			DateStarted: testutils.Date(2003, 1, 1),
		},
		{
			Title:       "s5",
			DateStarted: testutils.Date(2004, 1, 1),
		},
	}

	for _, s := range serieses {
		err := r.SeriesCreate(ctx, user.ID, s)
		require.NoError(err)
	}

	// fetch serieses

	fetchedSerieses, err = r.SeriesesGetAll(ctx, queryOptions)
	require.NoError(err)
	require.Equal(len(serieses), len(fetchedSerieses))

	// SeriesGetAll records are order by id by default
	for i, fs := range fetchedSerieses {
		testutils.SetTimeLocation(
			&serieses[i].DateStarted,
			fs.DateStarted.Location(),
		)
		testutils.SetTimeLocation(
			&serieses[i].DateEnded.Time,
			fs.DateEnded.Time.Location(),
		)
		require.Equal(serieses[i], fs)
	}
}

func TestSeriesesCount(t *testing.T) {
	require := require.New(t)

	teardown := setup()
	t.Cleanup(teardown)

	r := repo.NewRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "email"}
	err := r.UserCreate(ctx, user)
	require.NoError(err)

	// first there's no serieses

	nSerieses, err := r.SeriesesCount(ctx)
	require.NoError(err)
	require.Equal(0, nSerieses)

	// insert serieses

	serieses := []*models.Series{
		{
			Title:       "s1",
			DateStarted: testutils.Date(2000, 1, 1),
		},
		{
			Title:       "s2",
			DateStarted: testutils.Date(2001, 1, 1),
		},
		{
			Title:       "s3",
			DateStarted: testutils.Date(2002, 1, 1),
		},
		{
			Title:       "s4",
			DateStarted: testutils.Date(2003, 1, 1),
		},
		{
			Title:       "s5",
			DateStarted: testutils.Date(2004, 1, 1),
		},
	}

	for _, s := range serieses {
		err := r.SeriesCreate(ctx, user.ID, s)
		require.NoError(err)
	}

	// count serieses

	nSerieses, err = r.SeriesesCount(ctx)
	require.NoError(err)
	require.Equal(len(serieses), nSerieses)
}

func TestSeriesCreate(t *testing.T) {
	require := require.New(t)

	teardown := setup()
	t.Cleanup(teardown)

	r := repo.NewRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "email"}
	err := r.UserCreate(ctx, user)
	require.NoError(err)
	series := &models.Series{
		Title:       "series",
		DateStarted: testutils.Date(1990, 1, 1),
	}

	// first there's no series

	nSerieses, err := r.SeriesesCount(ctx)
	require.NoError(err)
	require.Equal(0, nSerieses)

	// create series

	err = r.SeriesCreate(ctx, user.ID, series)
	require.NoError(err)

	// fetch the series

	fetchedSeries, err := r.SeriesGet(ctx, series.ID)
	require.NoError(err)

	testutils.SetTimeLocation(
		&series.DateStarted,
		fetchedSeries.DateStarted.Location(),
	)
	testutils.SetTimeLocation(
		&series.DateEnded.Time,
		fetchedSeries.DateEnded.Time.Location(),
	)

	require.Equal(series, fetchedSeries)
}

func TestSeriesUpdate(t *testing.T) {
	require := require.New(t)

	teardown := setup()
	t.Cleanup(teardown)

	r := repo.NewRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "email"}
	err := r.UserCreate(ctx, user)
	require.NoError(err)
	series := &models.Series{
		Title:       "serie",
		DateStarted: testutils.Date(2000, 1, 1),
	}

	newTitle := "new serie title"
	newDescriptions := "new serie descriptions"
	newDateStarted := testutils.Date(1990, 1, 1)
	newDateEnded := testutils.Date(2017, 1, 1)

	updateColumns := map[string]any{
		models.SeriesColumns.Title:        newTitle,
		models.SeriesColumns.Descriptions: newDescriptions,
		models.SeriesColumns.DateStarted:  newDateStarted,
		models.SeriesColumns.DateEnded:    newDateEnded,
	}

	// first there's no series

	err = r.SeriesUpdate(ctx, series.ID, user.ID, updateColumns)
	require.Equal(repo.ErrNoRecord, err)

	// add series

	err = r.SeriesCreate(ctx, user.ID, series)
	require.NoError(err)

	// update series

	outdatedSeries := series

	err = r.SeriesUpdate(ctx, series.ID, user.ID, updateColumns)
	require.NoError(err)

	// fetch the updated series

	fetchedUpdatedSeries, err := r.SeriesGet(ctx, series.ID)
	require.NoError(err)

	testutils.SetTimeLocation(
		&newDateStarted,
		fetchedUpdatedSeries.DateStarted.Location(),
	)
	testutils.SetTimeLocation(
		&newDateEnded,
		fetchedUpdatedSeries.DateEnded.Time.Location(),
	)

	require.Equal(
		&models.Series{
			ID:            outdatedSeries.ID,
			Title:         newTitle,
			Descriptions:  null.StringFrom(newDescriptions),
			DateStarted:   newDateStarted,
			DateEnded:     null.TimeFrom(newDateEnded),
			ContributedBy: user.ID,
			// let's imagine we don't care about ContributedAt field for a moment :|
			ContributedAt: fetchedUpdatedSeries.ContributedAt,
			Invalidation:  outdatedSeries.Invalidation,
		},
		fetchedUpdatedSeries,
	)

	// check outdated series audited

	audits, err := r.SeriesAuditsGetAll(
		ctx,
		series.ID,
		query.SortOrderOptions{
			Offset:    0,
			Limit:     math.MaxInt,
			SortOrder: "desc",
		},
	)
	require.NoError(err)
	require.Equal(1, len(audits))

	testutils.SetTimeLocation(
		&outdatedSeries.DateStarted,
		audits[0].DateStarted.Location(),
	)
	testutils.SetTimeLocation(
		&outdatedSeries.DateEnded.Time,
		audits[0].DateEnded.Time.Location(),
	)

	require.Equal(
		outdatedSeries,
		&models.Series{
			ID:            audits[0].ID,
			Title:         audits[0].Title,
			Descriptions:  audits[0].Descriptions,
			DateStarted:   audits[0].DateStarted,
			DateEnded:     audits[0].DateEnded,
			ContributedBy: audits[0].ContributedBy,
			ContributedAt: audits[0].ContributedAt,
			Invalidation:  audits[0].Invalidation,
		},
	)
}

////////////////////////////////////////////////////////////////////////////////

func TestSeriesAuditsGetAll(t *testing.T) {
	require := require.New(t)

	teardown := setup()
	t.Cleanup(teardown)

	r := repo.NewRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "email"}
	err := r.UserCreate(ctx, user)
	require.NoError(err)

	series := &models.Series{
		Title:       "series",
		DateStarted: testutils.Date(2000, 1, 1),
	}
	err = r.SeriesCreate(ctx, user.ID, series)
	require.NoError(err)

	queryOptions := query.SortOrderOptions{
		Offset:    0,
		Limit:     math.MaxInt,
		SortOrder: "desc",
	}

	// first there's no audits

	audits, err := r.SeriesAuditsGetAll(ctx, series.ID, queryOptions)
	require.NoError(err)
	require.Equal(0, len(audits))

	// audits for episode

	seriesNewUpdates := []map[string]any{
		{models.SeriesColumns.Title: "nt1"},
		{models.SeriesColumns.Descriptions: "nd2"},
		{
			models.SeriesColumns.DateStarted: testutils.Date(2002, 11, 13),
		},
		{
			models.SeriesColumns.DateEnded: testutils.Date(2012, 5, 7),
		},

		{
			models.SeriesColumns.Title:        "nt5",
			models.SeriesColumns.Descriptions: "nd5",
			models.SeriesColumns.DateStarted:  testutils.Date(2004, 3, 6),
			models.SeriesColumns.DateEnded:    testutils.Date(2008, 2, 24),
		},
		{models.SeriesColumns.Title: "nt6"},
	}

	for _, snu := range seriesNewUpdates {
		err := r.SeriesUpdate(
			ctx,
			series.ID,
			user.ID,
			snu,
		)
		require.NoError(err)
	}

	// fetch audits

	audits, err = r.SeriesAuditsGetAll(ctx, series.ID, queryOptions)
	require.NoError(err)
	require.Equal(len(seriesNewUpdates), len(audits))

	testutils.SetTimeLocation(
		&series.DateStarted,
		audits[len(audits)-1].DateStarted.Location(),
	)
	testutils.SetTimeLocation(
		&series.DateEnded.Time,
		audits[len(audits)-1].DateEnded.Time.Location(),
	)

	require.Equal(
		series,
		&models.Series{
			ID:            audits[len(audits)-1].ID,
			Title:         audits[len(audits)-1].Title,
			Descriptions:  audits[len(audits)-1].Descriptions,
			DateStarted:   audits[len(audits)-1].DateStarted,
			DateEnded:     audits[len(audits)-1].DateEnded,
			ContributedBy: audits[len(audits)-1].ContributedBy,
			ContributedAt: audits[len(audits)-1].ContributedAt,
			Invalidation:  audits[len(audits)-1].Invalidation,
		},
	)

	for i, a := range audits[:len(audits)-1] {
		auditedSeries := &models.SeriesesAudit{}
		auditedSeries.ID = series.ID

		snu := seriesNewUpdates[len(seriesNewUpdates)-i-2]

		if titleString, exists := snu[models.SeriesColumns.Title]; exists {
			title, isString := titleString.(string)
			require.True(isString)
			auditedSeries.Title = title
		} else {
			auditedSeries.Title = audits[i+1].Title
		}
		if descriptionsString, exists := snu[models.SeriesColumns.Descriptions]; exists {
			descriptions, isString := descriptionsString.(string)
			require.True(isString)
			auditedSeries.Descriptions = null.StringFrom(descriptions)
		} else {
			auditedSeries.Descriptions = audits[i+1].Descriptions
		}
		if dateStartedTime, exists := snu[models.SeriesColumns.DateStarted]; exists {
			dateStarted, isTime := dateStartedTime.(time.Time)
			require.True(isTime)
			auditedSeries.DateStarted = dateStarted
		} else {
			auditedSeries.DateStarted = audits[i+1].DateStarted
		}
		if dateEndedTime, exists := snu[models.SeriesColumns.DateEnded]; exists {
			dateEnded, isTime := dateEndedTime.(time.Time)
			require.True(isTime)
			auditedSeries.DateEnded = null.TimeFrom(dateEnded)
		} else {
			auditedSeries.DateEnded = audits[i+1].DateEnded
		}
		auditedSeries.Invalidation = null.String{}
		auditedSeries.ContributedBy = user.ID
		auditedSeries.ContributedAt = a.ContributedAt

		testutils.SetTimeLocation(
			&auditedSeries.DateStarted,
			a.DateStarted.Location(),
		)
		testutils.SetTimeLocation(
			&auditedSeries.DateEnded.Time,
			a.DateEnded.Time.Location(),
		)

		require.Equal(auditedSeries, a)

	}
}

func TestSeriesAuditsCount(t *testing.T) {
	require := require.New(t)

	teardown := setup()
	t.Cleanup(teardown)

	r := repo.NewRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "email"}
	err := r.UserCreate(ctx, user)
	require.NoError(err)

	series := &models.Series{Title: "same series title"}
	err = r.SeriesCreate(ctx, user.ID, series)
	require.NoError(err)

	// first there's no audits

	auditsCount, err := r.SeriesAuditsCount(ctx, series.ID)
	require.NoError(err)
	require.Equal(0, auditsCount)

	// audits for episode

	seriesNewVersions := []map[string]any{
		{models.SeriesColumns.Title: "nt1"},
		{models.SeriesColumns.Title: "nt2"},
		{models.SeriesColumns.Title: "nt3"},
		{models.SeriesColumns.Title: "nt4"},
		{models.SeriesColumns.Title: "nt5"},
	}

	for _, snv := range seriesNewVersions {
		err := r.SeriesUpdate(
			ctx,
			series.ID,
			user.ID,
			snv,
		)
		require.NoError(err)
	}

	// count audits

	auditsCount, err = r.SeriesAuditsCount(ctx, series.ID)
	require.NoError(err)
	require.Equal(len(seriesNewVersions), auditsCount)
}
