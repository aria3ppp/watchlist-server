package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/aria3ppp/watchlist-server/internal/app"
	"github.com/aria3ppp/watchlist-server/internal/auth"
	"github.com/aria3ppp/watchlist-server/internal/config"
	"github.com/aria3ppp/watchlist-server/internal/hasher"
	"github.com/aria3ppp/watchlist-server/internal/repo"
	"github.com/aria3ppp/watchlist-server/internal/search"
	"github.com/aria3ppp/watchlist-server/internal/server"
	"github.com/aria3ppp/watchlist-server/internal/storage"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	var logFile *os.File
	if config.Config.Server.Production {
		var err error
		logFile, err = os.Create(config.Config.Server.Logfile)
		if err != nil {
			log.Panicf(
				"failed creating log file %q: %s",
				config.Config.Server.Logfile,
				err,
			)
		}
	}

	logger := newZapLogger(logFile)
	defer logger.Sync()

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		config.Config.Postgres.User,
		config.Config.Postgres.Password,
		config.Config.Postgres.Host,
		config.Config.Postgres.Port,
		config.Config.Postgres.DB,
	)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		logger.Panic("failed openning databse connection", zap.Error(err))
	}
	err = db.Ping()
	if err != nil {
		logger.Panic("failed to ping database connection", zap.Error(err))
	}

	repository := repo.NewRepository(db)
	hasher := hasher.NewBcrypt(bcrypt.DefaultCost)

	signingKey, err := auth.ECPrivateKeyFromBase64(
		[]byte(config.Config.Auth.ECDSASigningKeyBase64),
		logger,
	)
	if err != nil {
		logger.Panic("failed parsing signing key", zap.Error(err))
	}

	auth := auth.NewAuth(
		signingKey,
		config.Config.Auth.ExpireInSecs.Jwt,
		config.Config.Auth.ExpireInSecs.Refresh,
	)

	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{config.Config.Elasticsearch.Url},
		Logger:    &esCustomLogger{logger},
	})
	if err != nil {
		logger.Panic("failed creating elasticsearch client", zap.Error(err))
	}
	searchService, err := search.NewElasticSearch(esClient)
	if err != nil {
		logger.Panic("failed initializing search service", zap.Error(err))
	}

	minioClient, err := minio.New(config.Config.MinIO.Url, &minio.Options{
		Creds: credentials.NewStaticV4(
			config.Config.MinIO.RootUser,
			config.Config.MinIO.RootPassword,
			"",
		),
	})
	if err != nil {
		logger.Panic("failed creating minio client", zap.Error(err))
	}
	storageService, err := storage.NewMinIO(minioClient)
	if err != nil {
		logger.Panic("failed initializing storage service", zap.Error(err))
	}

	application := app.NewApplication(
		repository,
		auth,
		searchService,
		hasher,
		storageService,
	)

	server := server.NewServer(
		application,
		echo.New(),
		func(ctx echo.Context, s string) (any, error) { return auth.ParseJwtToken(s) },
		echo.MustSubFS(openapiFS, "openapi"),
		logger,
	)
	server.Run(fmt.Sprintf(":%d", config.Config.Server.Port))
}
