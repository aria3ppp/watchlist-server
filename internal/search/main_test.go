package search_test

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

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
	if err := searchtestutils.DeleteIndex(
		esClient,
		config.Config.Elasticsearch.Index.Serieses,
		config.Config.Elasticsearch.Index.Movies,
	); err != nil {
		log.Fatalf("search_test.teardown: error deleting indices: %s", err)
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

	err := config.Load(filepath.Join("..", "..", "config.yml"))
	if err != nil {
		log.Fatalf("search_test.TestMain: failed loading configs: %s", err)
	}

	// setup client
	esClient, err = elasticsearch.NewClient(
		elasticsearch.Config{Addresses: []string{"http://localhost:9200"}},
	)
	if err != nil {
		log.Fatalf(
			"search_test.TestMain: elasticsearcch.NewClient error: %s",
			err,
		)
	}

	// run test cases
	code := m.Run()

	os.Exit(code)
}
