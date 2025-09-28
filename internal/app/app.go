package app

import (
	"context"
	"gin/user-management-api/internal/config"
	"gin/user-management-api/internal/db"
	"gin/user-management-api/internal/db/sqlc"
	"gin/user-management-api/internal/routes"
	"gin/user-management-api/internal/validation"
	"gin/user-management-api/pkg/auth"
	"gin/user-management-api/pkg/cache"
	"gin/user-management-api/pkg/loggers"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Module interface {
	Routes() routes.Route
}

type Application struct {
	config  *config.Config
	routes  *gin.Engine
	modules []Module
}

type MouldeContext struct {
	DB    sqlc.Querier
	Redis *redis.Client
}

func NewApplication(cfg *config.Config) *Application {
	if err := validation.InitValidator(); err != nil {
		loggers.Log.Fatal().Err(err).Msg("Failed to initialize validator")
	}

	r := gin.Default()

	if err := db.InitDB(); err != nil {
		loggers.Log.Fatal().Err(err).Msg("Database init failed")
	}

	redisClinet := config.NewRedisClient()
	cacheRedisService := cache.NewRedisCacheService(redisClinet)
	tokenService := auth.NewJWTService(cacheRedisService)

	ctx := &MouldeContext{
		DB:    db.DB,
		Redis: redisClinet,
	}

	models := []Module{
		NewUserModule(ctx),
		NewAuthModule(ctx, tokenService, cacheRedisService),
	}

	routes.RegisterRoutes(r, tokenService, cacheRedisService, getModlRoutes(models)...)

	return &Application{
		config:  cfg,
		routes:  r,
		modules: models,
	}
}

func (app *Application) Run() error {
	srv := &http.Server{
		Addr:    app.config.ServerAddress,
		Handler: app.routes,
	}

	quit := make(chan os.Signal, 1)

	// syscall.SIGINT -> ctrl + c
	// syscall.SIGTERM -> kill service
	// syscall.SIGHUP -> reload service
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		loggers.Log.Info().Msgf("Server is running at %s \n", app.config.ServerAddress)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			loggers.Log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	<-quit
	loggers.Log.Warn().Msg("Shutdown signal received.....")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		loggers.Log.Fatal().Err(err).Msg("Server corced to shutdown")
	}
	loggers.Log.Info().Msg("Server exited gracefully")
	return nil
}

func getModlRoutes(modules []Module) []routes.Route {
	routeList := make([]routes.Route, len(modules))
	for i, module := range modules {
		routeList[i] = module.Routes()
	}
	return routeList
}
