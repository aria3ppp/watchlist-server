package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aria3ppp/watchlist-server/internal/app"
	"github.com/aria3ppp/watchlist-server/internal/config"
	"github.com/aria3ppp/watchlist-server/internal/hasher"
	"github.com/aria3ppp/watchlist-server/internal/repo"
	"github.com/aria3ppp/watchlist-server/internal/search"
	"github.com/aria3ppp/watchlist-server/internal/server"
	"github.com/aria3ppp/watchlist-server/internal/storage"
	"github.com/aria3ppp/watchlist-server/internal/token"
	"github.com/elastic/elastic-transport-go/v8/elastictransport"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

func main() {
	var logFile *os.File
	if config.Config.Server.Production {
		var err error
		logFile, err = os.Create(config.Config.Server.Logfile)
		if err != nil {
			log.Fatalf(
				"failed creating log file %q: %s",
				config.Config.Server.Logfile,
				err,
			)
		}
	}

	logger := newLogger(logFile)
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
		logger.Fatal("failed openning databse connection", zap.Error(err))
	}
	err = db.Ping()
	if err != nil {
		logger.Fatal("failed to ping database connection", zap.Error(err))
	}

	repository := repo.NewRepository(db)
	hasher := hasher.NewBcrypt()

	tokenService := token.NewJWT(
		token.JWTConfig{
			Key:           []byte(config.Config.Token.SigningKey),
			SigningMethod: jwt.SigningMethodHS512,
			AccessDuration: time.Minute * time.Duration(
				config.Config.Token.Access.Duration.InMinutes,
			),
			RefreshDuration: time.Minute * time.Duration(
				config.Config.Token.Refresh.Duration.InMinutes,
			),
		},
	)

	var esLogger elastictransport.Logger
	if config.Config.Server.Production {
		esLogger = &esCustomLogger{logger}
	} else {
		esLogger = &elastictransport.ColorLogger{
			Output:             os.Stdout,
			EnableRequestBody:  true,
			EnableResponseBody: true,
		}
	}
	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{config.Config.Elasticsearch.Url},
		Logger:    esLogger,
	})
	if err != nil {
		logger.Fatal("failed creating elasticsearch client", zap.Error(err))
	}
	searchService, err := search.NewElasticSearch(esClient)
	if err != nil {
		logger.Fatal("failed initializing search service", zap.Error(err))
	}

	minioClient, err := minio.New(config.Config.MinIO.Url, &minio.Options{
		Creds: credentials.NewStaticV4(
			config.Config.MinIO.RootUser,
			config.Config.MinIO.RootPassword,
			"",
		),
	})
	if err != nil {
		logger.Fatal("failed creating minio client", zap.Error(err))
	}
	storageService, err := storage.NewMinIO(minioClient)
	if err != nil {
		logger.Fatal("failed initializing storage service", zap.Error(err))
	}

	application := app.NewApplication(
		repository,
		tokenService,
		searchService,
		hasher,
		storageService,
	)

	server := server.NewServer(application, echo.New(), tokenService, logger)
	server.Run(fmt.Sprintf(":%d", config.Config.Server.Port))
}
