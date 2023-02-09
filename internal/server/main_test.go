package server_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/aria3ppp/watchlist-server/internal/app"
	"github.com/aria3ppp/watchlist-server/internal/auth"
	"github.com/aria3ppp/watchlist-server/internal/config"
	"github.com/aria3ppp/watchlist-server/internal/dto"
	"github.com/aria3ppp/watchlist-server/internal/hasher"
	"github.com/aria3ppp/watchlist-server/internal/repo"
	"github.com/aria3ppp/watchlist-server/internal/search"
	"github.com/aria3ppp/watchlist-server/internal/search/searchtestutils"
	appServer "github.com/aria3ppp/watchlist-server/internal/server"
	"github.com/aria3ppp/watchlist-server/internal/storage"
	"github.com/aria3ppp/watchlist-server/internal/storage/storagetestutils"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/labstack/echo/v4"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// To run this test suite set TEST_E2E env
const ENV_TEST_E2E = "TEST_E2E"

type SetupOpt int

const (
	OptEnableLogger SetupOpt = 1 << iota
	OptEnableDefaultUser
	OptEnableDefaultSeries
)

type Defaults struct {
	user   *DefaultUser
	series *DefaultSeries
}
type DefaultUser struct {
	id           int
	email        string
	password     string
	auth         string
	refreshToken string
	reqObject    *dto.UserCreateRequest
}
type DefaultSeries struct {
	id        int
	reqObject *dto.SeriesCreateRequest
}

// setup test cases
func setup(
	options ...SetupOpt,
) (testServer *httptest.Server, appInstance *app.Application, defaults *Defaults, teardown func()) {
	var opts SetupOpt
	if len(options) > 0 {
		opts = options[0]
	}

	logger := zap.NewNop()
	if OptEnableLogger&opts != 0 {
		var err error
		logger, err = zap.NewDevelopment()
		if err != nil {
			log.Panicf("server_test.setup: zap.NewDevelopment error: %s", err)
		}
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Panicf("server_test.setup: postgres.WithInstance error: %s", err)
	}
	migrator, err := migrate.NewWithDatabaseInstance(
		"file://../../migrations",
		"postgres", driver,
	)
	if err != nil {
		log.Panicf(
			"server_test.setup: migrate.NewWithDatabaseInstance error: %s",
			err,
		)
	}
	// run migrations
	err = migrator.Up()
	if err != nil {
		log.Panicf("server_test.setup: migrator.Up error: %s", err)
	}

	// set server to run in production mode
	config.Config.Server.Production = true

	signingKey, err := auth.ECPrivateKeyFromBase64(
		[]byte(config.Config.Auth.ECDSASigningKeyBase64),
		logger,
	)
	if err != nil {
		log.Panicf(
			"server_test.setup: auth.ECPrivateKeyFromBase64 error: %s|%s",
			err, config.Config.Auth.ECDSASigningKeyBase64,
		)
	}

	// initialize server
	repo := repo.NewRepository(db)
	hasher := hasher.NewBcrypt(bcrypt.DefaultCost)
	auth := auth.NewAuth(
		signingKey,
		config.Config.Auth.ExpireInSecs.Jwt,
		config.Config.Auth.ExpireInSecs.Refresh,
	)
	searchService, err := search.NewElasticSearch(esClient)
	if err != nil {
		log.Panicf("server_test.setup: search.NewElasticSearch error: %s", err)
	}
	storageService, err := storage.NewMinIO(minioClient)
	if err != nil {
		log.Panicf("server_test.setup: storage.NewMinIO error: %s", err)
	}
	appInstance = app.NewApplication(
		repo,
		auth,
		searchService,
		hasher,
		storageService,
	)
	router := echo.New()
	server := appServer.NewServer(
		appInstance,
		router,
		func(_ echo.Context, ts string) (any, error) { return auth.ParseJwtToken(ts) },
		// TODO: should we test whether openapi documentaion path exists?
		nil,
		logger,
	)
	testServer = httptest.NewServer(server.GetHandler())

	var defaultUser *DefaultUser
	if OptEnableDefaultUser&opts != 0 || OptEnableDefaultSeries&opts != 0 {
		var (
			email    = "frank@prog.net"
			password = "pa$$W0RD1"
		)
		req := &dto.UserCreateRequest{Email: email, Password: password}
		id, err := appInstance.UserCreate(
			context.Background(),
			req,
		)
		if err != nil {
			log.Panicf(
				"server_test.setup: appInstance.UserCreate error: %s",
				err,
			)
		}
		loginTokens, err := appInstance.UserLogin(
			context.Background(),
			&dto.UserLoginRequest{
				Email:    email,
				Password: password,
			},
		)
		if err != nil {
			log.Panicf(
				"server_test.setup: appInstance.UserLogin error: %s",
				err,
			)
		}
		defaultUser = &DefaultUser{
			id:           id,
			email:        email,
			password:     password,
			auth:         "Bearer " + loginTokens.JwtToken,
			refreshToken: loginTokens.RefreshToken,
			reqObject:    req,
		}
	}

	var defaultSeries *DefaultSeries
	if OptEnableDefaultSeries&opts != 0 {
		req := &dto.SeriesCreateRequest{Title: "default series"}
		id, err := appInstance.SeriesCreate(
			context.Background(),
			defaultUser.id,
			req,
		)
		if err != nil {
			log.Panicf(
				"server_test.setup: appInstance.SeriesCreate error: %s",
				err,
			)
		}
		defaultSeries = &DefaultSeries{id: id, reqObject: req}
	}

	defaults = &Defaults{
		user:   defaultUser,
		series: defaultSeries,
	}

	// prepare teardown
	teardown = func() {
		// delete bucket
		if err := storagetestutils.DeleteBucketWait(
			minioClient,
			10*time.Second,
			time.Second,
			config.Config.MinIO.Bucket.Image.Name,
		); err != nil {
			log.Panicf("storage_test.teardown: error deleting buckets: %s", err)
		}
		// delete elasticsearch indices
		err = searchtestutils.DeleteIndexWait(
			esClient,
			10*time.Second,
			time.Second,
			config.Config.Elasticsearch.Index.Serieses,
			config.Config.Elasticsearch.Index.Movies,
		)
		if err != nil {
			log.Panicf(
				"server_test.teardown: searchtestutils.DeleteIndex error: %s",
				err,
			)
		}
		// drop migrations
		err = migrator.Drop()
		if err != nil {
			log.Panicf("server_test.teardown: migrator.Drop error: %s", err)
		}
		// close server
		testServer.Close()
	}

	return testServer, appInstance, defaults, teardown
}

var (
	db          *sql.DB
	esClient    *elasticsearch.Client
	minioClient *minio.Client
)

func TestMain(m *testing.M) {
	// run only when TEST_E2E env is set
	if os.Getenv(ENV_TEST_E2E) == "" {
		fmt.Printf(
			"server_test.TestMain: end-2-end tests skipped: to enable, set %s env!\n",
			ENV_TEST_E2E,
		)
		return
	}

	// setup db
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
			"server_test.TestMain: could not connect to database %q: %s",
			dsn,
			err,
		)
	}
	err = db.Ping()
	if err != nil {
		log.Panicf(
			"server_test.TestMain: could not ping database %q: %s",
			dsn,
			err,
		)
	}

	// setup es
	esClient, err = elasticsearch.NewClient(
		elasticsearch.Config{Addresses: []string{"http://localhost:9200"}},
	)
	if err != nil {
		log.Panicf(
			"server_test.TestMain: elasticsearcch.NewClient error: %s",
			err,
		)
	}

	// setup minio
	minioClient, err = minio.New("localhost:9000", &minio.Options{
		Creds: credentials.NewStaticV4(
			os.Getenv("MINIO_ROOT_USER"),
			os.Getenv("MINIO_ROOT_PASSWORD"),
			"",
		),
	})
	if err != nil {
		log.Panicf("server_test.TestMain: minio.New error: %s", err)
	}

	// Run tests
	code := m.Run()

	// close db
	err = db.Close()
	if err != nil {
		log.Panicf("server_test.TestMain: db.Close error: %s", err)
	}

	os.Exit(code)
}
