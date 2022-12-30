package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aria3ppp/watchlist-server/internal/app"
	"github.com/aria3ppp/watchlist-server/internal/config"
	"github.com/aria3ppp/watchlist-server/internal/token"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

type Server struct {
	app          app.Service
	router       *echo.Echo
	tokenService token.Service
	logger       *zap.Logger
}

func NewServer(
	app app.Service,
	router *echo.Echo,
	tokenService token.Service,
	logger *zap.Logger,
) *Server {
	if !config.Config.Server.Production {
		router.Debug = true
	}
	server := &Server{
		app:          app,
		router:       router,
		tokenService: tokenService,
		logger:       logger,
	}
	server.setHandlers()
	return server
}

func (s *Server) setHandlers() {
	// add trailing slash
	// so all pathes must end with a slash
	s.router.Pre(middleware.AddTrailingSlash())

	// set timeout middleware for all paths
	// this throws '503 Service Unavailable' on request timeout
	s.router.Use(
		middleware.TimeoutWithConfig(
			middleware.TimeoutConfig{
				Timeout: time.Second * time.Duration(
					config.Config.Server.HandlerTimeoutInSeconds,
				),
			},
		),
	)

	v1 := s.router.Group("/v1")

	user := v1.Group("/user")
	user.POST("/", s.HandleUserCreate)
	user.POST("/login/", s.HandleUserLogin)
	user.GET("/refresh/", s.HandleUserRefreshToken)

	// set jwt middleware for authorized paths
	authorized := v1.Group("/authorized", s.AuthMiddleware)

	authorizedUser := authorized.Group("/user")
	authorizedUser.GET("/:id/", s.HandleUserGet)
	authorizedUser.PATCH("/", s.HandleUserUpdate)
	authorizedUser.PUT("/email/", s.HandleUserEmailUpdate)
	authorizedUser.PUT("/password/", s.HandleUserPasswordUpdate)
	authorizedUser.DELETE("/", s.HandleUserDelete)
	authorizedUser.PUT("/avatar/", s.HandleUserPutAvatar)

	// TODO: Implement access-based modifications:
	// Only allowed users could change or delete a specific resource ==> admin? list of permited users per resource?
	// table serie_permission:
	//
	// id | user_id | serie_id
	// -----------------------
	//  1 |    23   |   244
	//------------------------
	//  2 |    23   |   245
	//------------------------
	//  2 |    23   |   246
	//------------------------
	//  2 |    24   |   270
	//
	// By default first contributor to this resouce is the owner, and they are the only one (beside possible admins)
	//  who can grant modification access to another user.

	movies := authorized.Group("/movie")
	movies.GET("/", s.HandleMoviesGetAll)
	movies.POST("/", s.HandleMovieCreate)
	movies.GET("/search/", s.HandleMoviesSearch)

	movie := movies.Group("/:id")
	movie.GET("/", s.HandleMovieGet)
	movie.PATCH("/", s.HandleMovieUpdate)
	movie.DELETE("/", s.HandleMovieInvalidate)
	movie.GET("/audits/", s.HandleMovieAuditsGetAll)
	movie.PUT("/poster/", s.HandleMoviePutPoster)

	serieses := authorized.Group("/series")
	serieses.GET("/", s.HandleSeriesesGetAll)
	serieses.POST("/", s.HandleSeriesCreate)
	serieses.GET("/search/", s.HandleSeriesesSearch)

	series := serieses.Group("/:id")
	series.GET("/", s.HandleSeriesGet)
	series.PATCH("/", s.HandleSeriesUpdate)
	series.DELETE("/", s.HandleSeriesInvalidate)
	series.GET("/audits/", s.HandleSeriesAuditsGetAll)
	series.PUT("/poster/", s.HandleSeriesPutPoster)

	series.GET("/episode/", s.HandleEpisodesGetAllBySeries)

	episodes := series.Group("/season/:season_number/episode")
	episodes.GET("/", s.HandleEpisodesGetAllBySeason)
	episodes.PUT("/", s.HandleEpisodesPutAllBySeason)
	episodes.DELETE("/", s.HandleEpisodesInvalidateAllBySeason)

	episode := episodes.Group("/:episode_number")
	episode.GET("/", s.HandleEpisodeGet)
	episode.PUT("/", s.HandleEpisodePut)
	episode.PATCH("/", s.HandleEpisodeUpdate)
	episode.DELETE("/", s.HandleEpisodeInvalidate)
	episode.GET("/audits/", s.HandleEpisodeAuditsGetAll)

	watchlist := authorized.Group("/watchlist")
	watchlist.GET("/", s.HandleWatchlistGet)
	watchlist.POST("/add/", s.HandleWatchlistAdd)
	watchlist.DELETE("/:id/", s.HandleWatchlistDelete)
	watchlist.PATCH("/:id/", s.HandleWatchlistSetWatched)
}

func (s *Server) GetHandler() http.Handler {
	return s.router.Server.Handler
}

func (s *Server) Run(addr string) {
	waitForShutdown := make(chan struct{})
	// handle shutdown signal
	go func() {
		shutdownSignal := make(chan os.Signal, 1)
		signal.Notify(shutdownSignal, os.Interrupt)
		<-shutdownSignal
		ctx, cancel := context.WithTimeout(
			context.Background(),
			time.Second*time.Duration(
				config.Config.Server.ShutdownTimeoutInSeconds,
			),
		)
		defer cancel()
		if err := s.router.Shutdown(ctx); err != nil {
			s.logger.Error("server shutdown", zap.Error(err))
		}
		close(waitForShutdown)
	}()
	// start server
	if err := s.router.Start(addr); err != nil {
		if err != http.ErrServerClosed {
			s.logger.Error("server closed unexpectedly", zap.Error(err))
			// as there's no user interuption let server do gracefull shutdown
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		}
	}
	// wait for current connections to complete with timeout config.Config.Server.ShutdownTimeoutInSeconds
	<-waitForShutdown
	s.logger.Info("server closed!")
}
