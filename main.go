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
	"github.com/aria3ppp/watchlist-server/internal/token"
	"github.com/elastic/elastic-transport-go/v8/elastictransport"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func main() {
	if err := config.Load("config.yml"); err != nil {
		log.Fatalf("failed loading configs: %s", err)
	}

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
		log.Fatalf("search.NewElasticSearch error: %s", err)
	}

	application := app.NewApplication(
		repository,
		tokenService,
		searchService,
		hasher,
	)

	server := server.NewServer(application, echo.New(), tokenService, logger)
	server.Run(fmt.Sprintf(":%d", config.Config.Server.Port))
}
