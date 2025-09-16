package app

import (
	"context"
	"gin/user-management-api/internal/config"
	"gin/user-management-api/internal/db"
	"gin/user-management-api/internal/db/sqlc"
	"gin/user-management-api/internal/routes"
	"gin/user-management-api/internal/validation"
	"gin/user-management-api/pkg/auth"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

type Module interface {
	Routes() routes.Route
}

type Application struct {
	config *config.Config
	routes *gin.Engine
	modules []Module
}

type MouldeContext struct {
	DB sqlc.Querier
	Redis *redis.Client
}

func NewApplication(cfg *config.Config) *Application {
	LoadEnv()
	if err := validation.InitValidator(); err != nil {
		log.Fatalf("Failed to initialize validator: %v", err)
	}

	r := gin.Default()

	if err := db.InitDB(); err != nil {
		log.Fatalf("Database init failed: %v", err)
	}

	redisClinet := config.NewRedisClient()
	tokenService := auth.NewJWTService()

	ctx := &MouldeContext {
		DB: db.DB,
		Redis: redisClinet,
	}

	models := []Module{
		NewUserModule(ctx),
		NewAuthModule(ctx, tokenService),
	}

	routes.RegisterRoutes(r, tokenService, getModlRoutes(models)...)

	return &Application{
		config: cfg,
		routes: r,
		modules: models,
	}
}

func (app *Application) Run() error {
	srv := &http.Server{
		Addr: app.config.ServerAddress,
		Handler: app.routes,
	}

	quit := make(chan os.Signal,1)

	// syscall.SIGINT -> ctrl + c
	// syscall.SIGTERM -> kill service
	// syscall.SIGHUP -> reload service
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		log.Printf("Server is running at %s \n", app.config.ServerAddress)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	<- quit
	log.Println("Shutdown signal received.....")
	ctx, cancel := context.WithTimeout(context.Background(), 15 * time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server corced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
	return nil
}

func getModlRoutes(modules []Module) []routes.Route {
	routeList := make([]routes.Route, len(modules))
	for i, module := range modules {
		routeList[i] = module.Routes()
	}
	return routeList
}

func LoadEnv() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Println("No .env file found")
	}
}
