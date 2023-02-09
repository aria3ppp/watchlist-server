package search

import (
	"net/http"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
)

const (
	seriesesIndexMappings = `{
		"mappings": {
			"properties": {
				"id": { "type": "keyword", "index": false },
				"title": { "type": "text" },
				"descriptions": { "type": "text" },
				"date_started": { "type": "date", "index": false },
				"date_ended": { "type": "date", "index": false },
				"contributed_by": { "type": "keyword", "index": false },
				"contributed_at": { "type": "date", "index": false },
				"invalidation": { "type": "keyword", "index": false }
			}
		}
	}`

	moviesIndexMappings = `{
		"mappings": {
			"properties": {
				"id": { "type": "keyword", "index": false },
				"title": { "type": "text" },
				"descriptions": { "type": "text" },
				"date_released": { "type": "date", "index": false },
				"duration": { "type": "short", "index": false },
				"contributed_by": { "type": "keyword", "index": false },
				"contributed_at": { "type": "date", "index": false },
				"invalidation": { "type": "keyword", "index": false }
			}
		}
	}`
)

func createIndexIfNotExists(
	client *elasticsearch.Client,
	index string,
	mappings string,
) error {
	// check index exists
	resp, err := client.Indices.Exists([]string{index})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// create index with mappings
	if resp.StatusCode == http.StatusNotFound {
		resp, err := client.Indices.Create(
			index,
			client.Indices.Create.WithBody(strings.NewReader(mappings)),
		)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.IsError() {
			return responseError(resp)
		}
	} else if resp.IsError() {
		return responseError(resp)
	}
	return nil
}
