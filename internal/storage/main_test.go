package storage_test

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/aria3ppp/watchlist-server/internal/config"
	"github.com/aria3ppp/watchlist-server/internal/storage/storagetestutils"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// To run this test suite set TEST_MINIO_INTEGRATION env
const ENV_TEST_MINIO_INTEGRATION = "TEST_MINIO_INTEGRATION"

var client *minio.Client

// run teardown on test's cleanup
func teardown() {
	// delete bucket
	if err := storagetestutils.DeleteBucketWait(
		client,
		10*time.Second,
		time.Second,
		config.Config.MinIO.Bucket.Image.Name,
	); err != nil {
		log.Panicf("storage_test.teardown: error deleting buckets: %s", err)
	}
}

func TestMain(m *testing.M) {
	// run only when ENV_TEST_MINIO_INTEGRATION env is set
	if os.Getenv(ENV_TEST_MINIO_INTEGRATION) == "" {
		fmt.Printf(
			"storage_test.TestMain: minio integration tests skipped: to enable, set %s env!\n",
			ENV_TEST_MINIO_INTEGRATION,
		)
		return
	}

	// setup client
	var err error
	client, err = minio.New("localhost:9000", &minio.Options{
		Creds: credentials.NewStaticV4(
			os.Getenv("MINIO_ROOT_USER"),
			os.Getenv("MINIO_ROOT_PASSWORD"),
			"",
		),
	})
	if err != nil {
		log.Panicf("storage_test.TestMain: minio.New error: %s", err)
	}

	// run test cases
	code := m.Run()

	os.Exit(code)
}
