package searchtestutils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"time"
	_ "unsafe"

	"github.com/aria3ppp/watchlist-server/internal/testutils"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

//go:linkname responseError github.com/aria3ppp/watchlist-server/internal/search.responseError
func responseError(*esapi.Response) error

func CreateDocument(
	client *elasticsearch.Client,
	index string,
	body []byte,
	id string,
) error {
	resp, err := client.Create(index, id, bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return responseError(resp)
	}
	return nil
}

func CountIndex(client *elasticsearch.Client, index string) (int, error) {
	resp, err := client.Count(
		client.Count.WithIndex(index),
		client.Count.WithBody(
			strings.NewReader(`{"query" : {"match_all" : {}}}`),
		),
	)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return 0, responseError(resp)
	}
	var r struct {
		Count int `json:"count"`
	}
	err = json.NewDecoder(resp.Body).Decode(&r)
	return r.Count, err
}

func DeleteIndexWait(
	client *elasticsearch.Client,
	timeout, cooldown time.Duration,
	index ...string,
) error {
	for _, idx := range index {
		resp, err := client.Indices.Delete([]string{idx})
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.IsError() {
			return responseError(resp)
		}
	}

	err := testutils.WaitUntil(
		func() (done bool, err error) {
			for _, idx := range index {
				if exists, err := IndexExists(client, idx); err != nil {
					return false, err
				} else if exists {
					return false, nil
				}
			}
			return true, nil
		},
		timeout,
		cooldown,
	)

	return err
}

func IndexExists(
	client *elasticsearch.Client,
	index string,
) (bool, error) {
	resp, err := client.Indices.Exists([]string{index})
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.IsError() {
		if resp.StatusCode == http.StatusNotFound {
			return false, nil
		}
		return false, responseError(resp)
	}
	return true, nil
}

/*
func (e *ElasticSearch) deleteMe() {
	client := e.client

	// create a new index
	client.Indices.Create(
		"new_index",
		// options
		client.Indices.Create.WithErrorTrace(),
		client.Indices.Create.WithTimeout(time.Second),
	)

	// create/update a document in index
	client.Index(
		"index",
		strings.NewReader("document body"),
		// options
		client.Index.WithDocumentID("id"),
		client.Index.WithContext(context.Background()),
	)

	// create a document in index
	client.Create(
		"index",
		"id",
		strings.NewReader("document body"),
		// options
		client.Create.WithContext(context.Background()),
		client.Create.WithHuman(),
	)

	// update a document in index
	client.Update(
		"index",
		"id",
		strings.NewReader("document body"),
		// options
		client.Update.WithContext(context.Background()),
	)

	// search documents
	client.Search(
		client.Search.WithBody(nil),
		client.Search.WithFilterPath([]string{}...),
		client.Search.WithFrom(0),
		client.Search.WithSize(100),
		client.Search.WithTrackTotalHits(true),
	)

	client.Reindex(
		strings.NewReader("body"),
		client.Reindex.WithRequestsPerSecond(0),
		client.Reindex.WithWaitForCompletion(true),
	)
}
*/
