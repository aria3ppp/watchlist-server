package server

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/aria3ppp/watchlist-server/internal/app"
	"github.com/aria3ppp/watchlist-server/internal/config"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

type Server struct {
	app            app.Service
	router         *echo.Echo
	parseTokenFunc func(echo.Context, string) (any, error)
	openapiSubFS   fs.FS
	logger         *zap.Logger
}

func NewServer(
	app app.Service,
	router *echo.Echo,
	parseTokenFunc func(echo.Context, string) (any, error),
	openapiSubFS fs.FS,
	logger *zap.Logger,
) *Server {
	if !config.Config.Server.Production {
		router.Debug = true
	}
	server := &Server{
		app:            app,
		router:         router,
		parseTokenFunc: parseTokenFunc,
		openapiSubFS:   openapiSubFS,
		logger:         logger,
	}
	server.setHandlers()
	return server
}

func (s *Server) setHandlers() {
	// log requests
	s.router.Use(
		middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
			LogURI:      true,
			LogStatus:   true,
			LogRemoteIP: true,
			LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
				s.logger.Info("request",
					zap.String("ip", v.RemoteIP),
					zap.String("URI", v.URI),
					zap.Int("status", v.Status),
				)
				return nil
			},
		}),
	)

	// set timeout middleware for all paths
	// if the request timeouts it sends "503 - Service Unavailable"
	s.router.Use(
		middleware.TimeoutWithConfig(
			middleware.TimeoutConfig{
				Timeout: time.Second * time.Duration(
					config.Config.Server.HandlerTimeoutInSeconds,
				),
			},
		),
	)

	// set request body max size for all paths
	// if the size exceeds it sends "413 - Request Entity Too Large" response
	s.router.Use(
		middleware.BodyLimit(fmt.Sprintf(
			"%dKB",
			config.Config.Validation.Request.Body.MaxLengthInKB,
		)),
	)

	// api v1
	{
		v1 := s.router.Group("/v1")

		// serve openapi spec documentation
		v1.StaticFS("/openapi", s.openapiSubFS)

		// user basic operations
		{
			user := v1.Group("/user")
			user.POST("", s.HandleUserCreate)
			user.POST("/login", s.HandleUserLogin)
			{
				userID := user.Group("/:id")
				userID.POST("/logout", s.HandleUserLogout)
				userID.POST("/refresh", s.HandleUserRefreshToken)
			}
		}

		// set jwt middleware for authorized path
		{
			authorized := v1.Group(
				"/authorized",
				echojwt.WithConfig(echojwt.Config{
					ContextKey:     contextKey,
					ParseTokenFunc: s.parseTokenFunc,
				}),
			)

			// user
			{
				authorizedUser := authorized.Group("/user")
				authorizedUser.GET("/:id", s.HandleUserGet)
				authorizedUser.PATCH("", s.HandleUserUpdate)
				authorizedUser.PUT("/email", s.HandleUserEmailUpdate)
				authorizedUser.PUT("/password", s.HandleUserPasswordUpdate)
				authorizedUser.DELETE("", s.HandleUserDelete)
				authorizedUser.PUT("/avatar", s.HandleUserPutAvatar)
			}

			// movie
			{
				movies := authorized.Group("/movie")
				movies.GET("", s.HandleMoviesGetAll)
				movies.POST("", s.HandleMovieCreate)
				movies.GET("/search", s.HandleMoviesSearch)

				{
					movie := movies.Group("/:id")
					movie.GET("", s.HandleMovieGet)
					movie.PATCH("", s.HandleMovieUpdate)
					movie.POST("/invalidate", s.HandleMovieInvalidate)
					movie.GET("/audits", s.HandleMovieAuditsGetAll)
					movie.PUT("/poster", s.HandleMoviePutPoster)
				}
			}

			// series
			{
				serieses := authorized.Group("/series")
				serieses.GET("", s.HandleSeriesesGetAll)
				serieses.POST("", s.HandleSeriesCreate)
				serieses.GET("/search", s.HandleSeriesesSearch)

				{
					series := serieses.Group("/:id")
					series.GET("", s.HandleSeriesGet)
					series.PATCH("", s.HandleSeriesUpdate)
					series.POST("/invalidate", s.HandleSeriesInvalidate)
					series.GET("/audits", s.HandleSeriesAuditsGetAll)
					series.PUT("/poster", s.HandleSeriesPutPoster)

					// episode
					series.GET("/episode", s.HandleEpisodesGetAllBySeries)
					{
						episodes := series.Group(
							"/season/:season_number/episode",
						)
						episodes.GET("", s.HandleEpisodesGetAllBySeason)
						episodes.PUT("", s.HandleEpisodesPutAllBySeason)
						episodes.POST(
							"/invalidate",
							s.HandleEpisodesInvalidateAllBySeason,
						)

						{
							episode := episodes.Group("/:episode_number")
							episode.GET("", s.HandleEpisodeGet)
							episode.PUT("", s.HandleEpisodePut)
							episode.PATCH("", s.HandleEpisodeUpdate)
							episode.POST(
								"/invalidate",
								s.HandleEpisodeInvalidate,
							)
							episode.GET("/audits", s.HandleEpisodeAuditsGetAll)
						}
					}
				}
			}

			// watchlist
			{
				watchlist := authorized.Group("/watchlist")
				watchlist.GET("", s.HandleWatchlistGet)
				watchlist.POST("/add", s.HandleWatchlistAdd)
				watchlist.DELETE("/:id", s.HandleWatchlistDelete)
				watchlist.PATCH("/:id", s.HandleWatchlistSetWatched)
			}
		}
	}
}

func (s *Server) GetHandler() http.Handler {
	return s.router.Server.Handler
}

func (s *Server) Run(addr string) {
	waitForShutdown := make(chan struct{})
	// handle shutdown signal
	go func() {
		shutdownSignal := make(chan os.Signal, 1)
		// listen to ctrl-C signal
		signal.Notify(shutdownSignal, os.Interrupt)
		<-shutdownSignal
		ctx, cancel := context.WithTimeout(
			context.Background(),
			time.Second*time.Duration(
				config.Config.Server.ShutdownTimeoutInSeconds,
			),
		)
		defer cancel()
		// shutdown the server gracefully
		if err := s.router.Shutdown(ctx); err != nil {
			s.logger.Error("server shutdown", zap.Error(err))
		}
		close(waitForShutdown)
	}()
	// start server
	if err := s.router.Start(addr); err != nil && err != http.ErrServerClosed {
		s.logger.Error("server closed unexpectedly", zap.Error(err))
	} else {
		// wait for current connections to complete with timeout config.Config.Server.ShutdownTimeoutInSeconds
		s.logger.Info(
			"waiting for shutdown!",
			zap.Duration(
				"timeout",
				time.Second*time.Duration(
					config.Config.Server.ShutdownTimeoutInSeconds,
				),
			),
		)
		<-waitForShutdown
		s.logger.Info("server closed!")
	}
}
