package repo_test

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// To run this test suite set TEST_DB_INTEGRATION env
const ENV_TEST_DB_INTEGRATION = "TEST_DB_INTEGRATION"

var db *sql.DB

// setup test cases
func setup() (teardown func()) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Panicf("repo_test.setup: postgres.WithInstance error: %s", err)
	}
	migrator, err := migrate.NewWithDatabaseInstance(
		"file://../../migrations",
		"postgres", driver,
	)
	if err != nil {
		log.Panicf(
			"repo_test.setup: migrate.NewWithDatabaseInstance error: %s",
			err,
		)
	}
	// run migrations
	err = migrator.Up()
	if err != nil {
		log.Panicf("repo_test.setup: migrator.Up error: %s", err)
	}
	// create teardown
	return func() {
		// drop migrations
		err = migrator.Drop()
		if err != nil {
			log.Panicf("repo_test.teardown: migrator.Drop error: %s", err)
		}
	}
}

func TestMain(m *testing.M) {
	// run only when TEST_DB_INTEGRATION env is set
	if os.Getenv(ENV_TEST_DB_INTEGRATION) == "" {
		fmt.Printf(
			"repo_test.TestMain: database integration tests skipped: to enable, set %s env!\n",
			ENV_TEST_DB_INTEGRATION,
		)
		return
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@localhost:%s/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DB"),
	)
	var err error
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Panicf(
			"repo_test.TestMain: could not connect to database %q: %s",
			dsn,
			err,
		)
	}
	err = db.Ping()
	if err != nil {
		log.Panicf(
			"repo_test.TestMain: could not ping database %q: %s",
			dsn,
			err,
		)
	}

	// Run tests
	code := m.Run()

	// close db
	err = db.Close()
	if err != nil {
		log.Panicf("repo_test.TestMain: db.Close error: %s", err)
	}

	os.Exit(code)
}
