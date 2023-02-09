package search_test

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/aria3ppp/watchlist-server/internal/config"
	"github.com/aria3ppp/watchlist-server/internal/search/searchtestutils"
	"github.com/elastic/go-elasticsearch/v8"
)

// To run this test suite set TEST_ES_INTEGRATION env
const ENV_TEST_ES_INTEGRATION = "TEST_ES_INTEGRATION"

var esClient *elasticsearch.Client

// run teardown on test's cleanup
func teardown() {
	// delete indices
	if err := searchtestutils.DeleteIndexWait(
		esClient,
		10*time.Second,
		time.Second,
		config.Config.Elasticsearch.Index.Serieses,
		config.Config.Elasticsearch.Index.Movies,
	); err != nil {
		log.Panicf("search_test.teardown: error deleting indices: %s", err)
	}
}

func TestMain(m *testing.M) {
	// run only when TEST_ELASTICSEARCH_INTEGRATION env is set
	if os.Getenv(ENV_TEST_ES_INTEGRATION) == "" {
		fmt.Printf(
			"search_test.TestMain: elasticsearch integration tests skipped: to enable, set %s env!\n",
			ENV_TEST_ES_INTEGRATION,
		)
		return
	}

	// setup client
	var err error
	esClient, err = elasticsearch.NewClient(
		elasticsearch.Config{Addresses: []string{"http://localhost:9200"}},
	)
	if err != nil {
		log.Panicf(
			"search_test.TestMain: elasticsearcch.NewClient error: %s",
			err,
		)
	}

	// run test cases
	code := m.Run()

	os.Exit(code)
}
