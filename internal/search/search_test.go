package search_test

import (
	"context"
	"encoding/json"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/aria3ppp/watchlist-server/internal/config"
	"github.com/aria3ppp/watchlist-server/internal/models"
	"github.com/aria3ppp/watchlist-server/internal/query"
	"github.com/aria3ppp/watchlist-server/internal/search"
	"github.com/aria3ppp/watchlist-server/internal/search/searchtestutils"
	"github.com/aria3ppp/watchlist-server/internal/testutils"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/null/v8"
)

func TestSearchSerieses(t *testing.T) {
	require := require.New(t)

	t.Cleanup(teardown)

	s, err := search.NewElasticSearch(esClient)
	require.NoError(err)

	ctx := context.Background()

	// no serieses

	gotSerieses, total, err := s.SearchSerieses(
		ctx,
		query.SearchOptions{
			Query: "query",
			From:  0,
			Size:  config.Config.Validation.Pagination.PageSize.MaxValue,
		},
	)
	require.NoError(err)
	require.Equal(0, total)
	require.Empty(gotSerieses)

	// index serieses

	querySerieses := []*models.Series{
		{Title: "query"},
		{Title: "title before query"},
		{Title: "query before title"},
		{Title: "title then query then title"},

		{Title: "QUERY"},
		{Title: "title before QUERY"},
		{Title: "QUERY before title"},
		{Title: "title then QUERY then title"},

		{Title: "uery"},
		{Title: "title before uery"},
		{Title: "uery before title"},
		{Title: "title then uery then title"},

		{Title: "qury"},
		{Title: "title before qury"},
		{Title: "qury before title"},
		{Title: "title then qury then title"},

		{Title: "quer"},
		{Title: "title before quer"},
		{Title: "quer before title"},
		{Title: "title then quer then title"},

		{Title: "Xuery"},
		{Title: "title before Xuery"},
		{Title: "Xuery before title"},
		{Title: "title then Xuery then title"},

		{Title: "quXry"},
		{Title: "title before quXry"},
		{Title: "quXry before title"},
		{Title: "title then quXry then title"},

		{Title: "querX"},
		{Title: "title before querX"},
		{Title: "querX before title"},
		{Title: "title then Xuery then title"},

		{Descriptions: null.StringFrom("query")},
		{Descriptions: null.StringFrom("descriptions before query")},
		{Descriptions: null.StringFrom("query before descriptions")},
		{Descriptions: null.StringFrom(
			"descriptions then query then descriptions",
		)},

		{Descriptions: null.StringFrom("QUERY")},
		{Descriptions: null.StringFrom("descriptions before QUERY")},
		{Descriptions: null.StringFrom("QUERY before descriptions")},
		{Descriptions: null.StringFrom(
			"descriptions then QUERY then descriptions",
		)},

		{Descriptions: null.StringFrom("uery")},
		{Descriptions: null.StringFrom("descriptions before uery")},
		{Descriptions: null.StringFrom("uery before descriptions")},
		{Descriptions: null.StringFrom(
			"descriptions then uery then descriptions",
		)},

		{Descriptions: null.StringFrom("qury")},
		{Descriptions: null.StringFrom("descriptions before qury")},
		{Descriptions: null.StringFrom("qury before descriptions")},
		{Descriptions: null.StringFrom(
			"descriptions then qury then descriptions",
		)},

		{Descriptions: null.StringFrom("quer")},
		{Descriptions: null.StringFrom("descriptions before quer")},
		{Descriptions: null.StringFrom("quer before descriptions")},
		{Descriptions: null.StringFrom(
			"descriptions then quer then descriptions",
		)},

		{Descriptions: null.StringFrom("Xuery")},
		{Descriptions: null.StringFrom("descriptions before Xuery")},
		{Descriptions: null.StringFrom("Xuery before descriptions")},
		{Descriptions: null.StringFrom(
			"descriptions then Xuery then descriptions",
		)},

		{Descriptions: null.StringFrom("quXry")},
		{Descriptions: null.StringFrom("descriptions before quXry")},
		{Descriptions: null.StringFrom("quXry before descriptions")},
		{Descriptions: null.StringFrom(
			"descriptions then quXry then descriptions",
		)},

		{Descriptions: null.StringFrom("querX")},
		{Descriptions: null.StringFrom("descriptions before querX")},
		{Descriptions: null.StringFrom("querX before descriptions")},
		{Descriptions: null.StringFrom(
			"descriptions then Xuery then descriptions",
		)},
	}

	// interleave querySerieses
	serieses := make([]*models.Series, len(querySerieses)*2)
	for i := range serieses {
		var series *models.Series
		if i%2 == 0 {
			qi := 0
			if i != 0 {
				qi = i / 2
			}
			series = querySerieses[qi]
		} else {
			series = &models.Series{
				Title:        "title",
				Descriptions: null.StringFrom("descriptions"),
			}
		}
		serieses[i] = series
	}

	// index documents
	for i, s := range serieses {
		id := i + 1
		s.ID = id
		jsonBody, err := json.Marshal(s)
		require.NoError(err)

		err = searchtestutils.CreateDocument(
			esClient,
			config.Config.Elasticsearch.Index.Serieses,
			jsonBody,
			strconv.Itoa(id),
		)
		require.NoError(err)
	}

	// wait until all documents are indexed
	err = testutils.WaitUntil(
		func() (bool, error) {
			c, err := searchtestutils.CountIndex(
				esClient,
				config.Config.Elasticsearch.Index.Serieses,
			)
			if err != nil {
				return false, err
			}
			return c == len(serieses), nil
		},
		10*time.Second,
		200*time.Millisecond,
	)
	require.NoError(err)

	// search query
	gotSerieses, total, err = s.SearchSerieses(
		ctx,
		query.SearchOptions{
			Query: "query",
			From:  0,
			Size:  config.Config.Validation.Pagination.PageSize.MaxValue,
		},
	)
	require.NoError(err)
	require.Equal(len(querySerieses), total)
	require.NotEmpty(gotSerieses)

	// sort results
	sort.Slice(
		gotSerieses,
		func(i, j int) bool { return gotSerieses[i].ID < gotSerieses[j].ID },
	)

	// check matched serieses
	for i, gs := range gotSerieses {
		require.Equal(querySerieses[i], gs)
	}
}

func TestSearchMovies(t *testing.T) {
	require := require.New(t)

	t.Cleanup(teardown)

	s, err := search.NewElasticSearch(esClient)
	require.NoError(err)

	ctx := context.Background()

	// no movies

	gotMovies, total, err := s.SearchMovies(
		ctx,
		query.SearchOptions{
			Query: "query",
			From:  0,
			Size:  config.Config.Validation.Pagination.PageSize.MaxValue,
		},
	)
	require.NoError(err)
	require.Equal(0, total)
	require.Empty(gotMovies)

	// index movies

	queryMovies := []*models.Film{
		{Title: "query"},
		{Title: "title before query"},
		{Title: "query before title"},
		{Title: "title then query then title"},

		{Title: "QUERY"},
		{Title: "title before QUERY"},
		{Title: "QUERY before title"},
		{Title: "title then QUERY then title"},

		{Title: "uery"},
		{Title: "title before uery"},
		{Title: "uery before title"},
		{Title: "title then uery then title"},

		{Title: "qury"},
		{Title: "title before qury"},
		{Title: "qury before title"},
		{Title: "title then qury then title"},

		{Title: "quer"},
		{Title: "title before quer"},
		{Title: "quer before title"},
		{Title: "title then quer then title"},

		{Title: "Xuery"},
		{Title: "title before Xuery"},
		{Title: "Xuery before title"},
		{Title: "title then Xuery then title"},

		{Title: "quXry"},
		{Title: "title before quXry"},
		{Title: "quXry before title"},
		{Title: "title then quXry then title"},

		{Title: "querX"},
		{Title: "title before querX"},
		{Title: "querX before title"},
		{Title: "title then Xuery then title"},

		{Descriptions: null.StringFrom("query")},
		{Descriptions: null.StringFrom("descriptions before query")},
		{Descriptions: null.StringFrom("query before descriptions")},
		{Descriptions: null.StringFrom(
			"descriptions then query then descriptions",
		)},

		{Descriptions: null.StringFrom("QUERY")},
		{Descriptions: null.StringFrom("descriptions before QUERY")},
		{Descriptions: null.StringFrom("QUERY before descriptions")},
		{Descriptions: null.StringFrom(
			"descriptions then QUERY then descriptions",
		)},

		{Descriptions: null.StringFrom("uery")},
		{Descriptions: null.StringFrom("descriptions before uery")},
		{Descriptions: null.StringFrom("uery before descriptions")},
		{Descriptions: null.StringFrom(
			"descriptions then uery then descriptions",
		)},

		{Descriptions: null.StringFrom("qury")},
		{Descriptions: null.StringFrom("descriptions before qury")},
		{Descriptions: null.StringFrom("qury before descriptions")},
		{Descriptions: null.StringFrom(
			"descriptions then qury then descriptions",
		)},

		{Descriptions: null.StringFrom("quer")},
		{Descriptions: null.StringFrom("descriptions before quer")},
		{Descriptions: null.StringFrom("quer before descriptions")},
		{Descriptions: null.StringFrom(
			"descriptions then quer then descriptions",
		)},

		{Descriptions: null.StringFrom("Xuery")},
		{Descriptions: null.StringFrom("descriptions before Xuery")},
		{Descriptions: null.StringFrom("Xuery before descriptions")},
		{Descriptions: null.StringFrom(
			"descriptions then Xuery then descriptions",
		)},

		{Descriptions: null.StringFrom("quXry")},
		{Descriptions: null.StringFrom("descriptions before quXry")},
		{Descriptions: null.StringFrom("quXry before descriptions")},
		{Descriptions: null.StringFrom(
			"descriptions then quXry then descriptions",
		)},

		{Descriptions: null.StringFrom("querX")},
		{Descriptions: null.StringFrom("descriptions before querX")},
		{Descriptions: null.StringFrom("querX before descriptions")},
		{Descriptions: null.StringFrom(
			"descriptions then Xuery then descriptions",
		)},
	}

	// interleave queryMovies
	movies := make([]*models.Film, len(queryMovies)*2)
	for i := range movies {
		var movie *models.Film
		if i%2 == 0 {
			qi := 0
			if i != 0 {
				qi = i / 2
			}
			movie = queryMovies[qi]
		} else {
			movie = &models.Film{
				Title:        "title",
				Descriptions: null.StringFrom("descriptions"),
			}
		}
		movies[i] = movie
	}

	// index documents
	for i, m := range movies {
		id := i + 1
		m.ID = id
		jsonBody, err := json.Marshal(m)
		require.NoError(err)

		err = searchtestutils.CreateDocument(
			esClient,
			config.Config.Elasticsearch.Index.Movies,
			jsonBody,
			strconv.Itoa(id),
		)
		require.NoError(err)
	}

	// wait until all documents are indexed
	err = testutils.WaitUntil(
		func() (bool, error) {
			c, err := searchtestutils.CountIndex(
				esClient,
				config.Config.Elasticsearch.Index.Movies,
			)
			if err != nil {
				return false, err
			}
			return c == len(movies), nil
		},
		10*time.Second,
		200*time.Millisecond,
	)
	require.NoError(err)

	// search query
	gotMovies, total, err = s.SearchMovies(
		ctx,
		query.SearchOptions{
			Query: "query",
			From:  0,
			Size:  config.Config.Validation.Pagination.PageSize.MaxValue,
		},
	)
	require.NoError(err)
	require.Equal(len(queryMovies), total)
	require.NotEmpty(gotMovies)

	// sort results
	sort.Slice(
		gotMovies,
		func(i, j int) bool { return gotMovies[i].ID < gotMovies[j].ID },
	)

	// check matched movies
	for i, gm := range gotMovies {
		require.Equal(queryMovies[i], gm)
	}
}
