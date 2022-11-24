package search

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"unsafe"

	"github.com/aria3ppp/watchlist-server/internal/config"
	"github.com/aria3ppp/watchlist-server/internal/models"
	"github.com/elastic/go-elasticsearch/v8"
)

//go:generate mockgen -destination mock_search/mock_service.go . Service

type Service interface {
	SearchSerieses(
		ctx context.Context,
		query string,
		from, size int,
	) (hits []*models.Series, totalHits int, err error)
	SearchMovies(
		ctx context.Context,
		query string,
		from, size int,
	) (hits []*models.Film, totalHits int, err error)
}

type ElasticSearch struct {
	client *elasticsearch.Client
}

var _ Service = &ElasticSearch{}

func NewElasticSearch(client *elasticsearch.Client) (*ElasticSearch, error) {
	// check cluster is up and running: this does not check if cluster is healthy
	if resp, err := client.Ping(); err != nil {
		return nil, err
	} else if resp.IsError() {
		return nil, responseError(resp)
	}
	// create serieses index
	if err := createIndexIfNotExists(
		client,
		config.Config.Elasticsearch.Index.Serieses,
		seriesesIndexMappings,
	); err != nil {
		return nil, err
	}
	// create movies index
	if err := createIndexIfNotExists(
		client,
		config.Config.Elasticsearch.Index.Movies,
		moviesIndexMappings,
	); err != nil {
		return nil, err
	}
	return &ElasticSearch{client: client}, nil
}

// note: caller must close 'responseBody.Close' manually
func (e *ElasticSearch) search(
	ctx context.Context,
	index string,
	query string,
	from, size int,
) (responseBody io.ReadCloser, err error) {
	resp, err := e.client.Search(
		e.client.Search.WithContext(ctx),
		e.client.Search.WithIndex(index),
		e.client.Search.WithBody(strings.NewReader(query)),
		e.client.Search.WithTrackTotalHits(true),
		e.client.Search.WithFrom(from),
		e.client.Search.WithSize(size),
	)
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, responseError(resp)
	}
	return resp.Body, nil
}

func (e *ElasticSearch) SearchSerieses(
	ctx context.Context,
	query string,
	from, size int,
) (hits []*models.Series, totalHits int, err error) {
	// prepare search query
	searchQuery := fmt.Sprintf(
		`{"query": {"multi_match": {"query": "%s", "fields": ["title", "descriptions"], "fuzziness": "AUTO"}}}`,
		query,
	)

	// search query
	responseBody, err := e.search(
		ctx,
		config.Config.Elasticsearch.Index.Serieses,
		searchQuery,
		from,
		size,
	)
	if err != nil {
		return nil, 0, err
	}
	defer responseBody.Close()

	// decode response body
	type R struct {
		Hits struct {
			Total struct {
				Value int
			}
			Hits []*seriesHit
		}
	}
	var r R
	if err := json.NewDecoder(responseBody).Decode(&r); err != nil {
		return nil, 0, err
	}

	// SAFETY: seriesHit is a subtype of models.Series
	hits = *(*[]*models.Series)(unsafe.Pointer(&r.Hits.Hits))
	totalHits = r.Hits.Total.Value

	return hits, totalHits, nil
}

func (e *ElasticSearch) SearchMovies(
	ctx context.Context,
	query string,
	from, size int,
) (hits []*models.Film, totalHits int, err error) {
	// prepare search query
	searchQuery := fmt.Sprintf(
		`{"query": {"multi_match": {"query": "%s", "fields": ["title", "descriptions"], "fuzziness": "AUTO"}}}`,
		query,
	)

	// search query
	responseBody, err := e.search(
		ctx,
		config.Config.Elasticsearch.Index.Movies,
		searchQuery,
		from,
		size,
	)
	if err != nil {
		return nil, 0, err
	}
	defer responseBody.Close()

	// decode response body
	type R struct {
		Hits struct {
			Total struct {
				Value int
			}
			Hits []*movieHit
		}
	}
	var r R
	if err := json.NewDecoder(responseBody).Decode(&r); err != nil {
		return nil, 0, err
	}

	// SAFETY: movieHit is a subtype of models.Film
	hits = *(*[]*models.Film)(unsafe.Pointer(&r.Hits.Hits))
	totalHits = r.Hits.Total.Value

	return hits, totalHits, nil
}
