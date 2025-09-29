package main

import (
	"context"
	"encoding/json"
	"gin/user-management-api/internal/config"
	"gin/user-management-api/internal/utils"
	"gin/user-management-api/pkg/loggers"
	"gin/user-management-api/pkg/mail"
	"gin/user-management-api/pkg/rabbitmq"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
)

type Worker struct {
	rabbitMQ    rabbitmq.RabbitMQSerivce
	mailService mail.EmailProviderService
	cfg         *config.Config
	logger      *zerolog.Logger
}

func newWorker(cfg *config.Config) *Worker {
	log := utils.NewLoggerWithPath("worker.log", "info")

	// Connect RabbitMQ
	rabbitMG, err := rabbitmq.NewRabbitMQService(
		utils.GetEnv("RABBITMQURL", "amqp://guest:guest@rabbitmq:5672/"), log,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to init RabbitMQ service")
	}

	// Init mail service

	mailLogger := utils.NewLoggerWithPath("mail.log", "info")
	factory, err := mail.NewProviderFactory(mail.ProviderMailtrap)
	if err != nil {
		mailLogger.Error().Err(err).Msg("Failed to create mail provider factory")
		return nil
	}

	mailService, err := mail.NewMailService(cfg, mailLogger, factory)
	if err != nil {
		mailLogger.Error().Err(err).Msg("Failed to initializa mail service")
		return nil
	}

	return &Worker{
		rabbitMQ:    rabbitMG,
		mailService: mailService,
		cfg:         cfg,
		logger:      log,
	}
}

func (wk *Worker) Start(ctx context.Context) error {
	const emailQueueName = "auth_email_queue"
	handler := func(body []byte) error {
		wk.logger.Debug().Msgf("Receiver message: %s", string(body))

		var email mail.Email
		if err := json.Unmarshal(body, &email); err != nil {
			wk.logger.Error().Err(err).Msg("Failed to unmarshal message")
			return err
		}

		if err := wk.mailService.SendMail(ctx, &email); err != nil {
			utils.NewError(utils.InternalServerError, "Failed to send password reset email")
		}

		wk.logger.Info().Msgf("Email sent successfully to %v", email.To)
		return nil
	}

	if err := wk.rabbitMQ.Consume(ctx, emailQueueName, handler); err != nil {
		wk.logger.Error().Err(err).Msg("Failed to start consumer")
		return err
	}

	wk.logger.Info().Msgf("Worker started, consuming from queue: %s", emailQueueName)
	<-ctx.Done()
	wk.logger.Info().Msgf("Worker stopped consuming due to context cancellation: %s", emailQueueName)
	return ctx.Err()
}

func (wk *Worker) Shutdown(ctx context.Context) error {
	wk.logger.Info().Msgf("Shutting down worker .....")
	if err := wk.rabbitMQ.Close(); err != nil {
		wk.logger.Error().Err(err).Msg("Failed to close rabbitMQ")
		return err
	}
	wk.logger.Info().Msgf("RabbitMQ connection closed successfully")

	select {
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			wk.logger.Warn().Msg("Shotdown timeout execeded")
			return ctx.Err()
		}
	default:
	}

	wk.logger.Info().Msg("Worker shotdown completed")
	return nil
}

func main() {
	rootDir := utils.MustGetWorkingDir()

	logFile := filepath.Join(rootDir, "internal/logs/app.log")

	loggers.InitLogger(loggers.LoggerConfig{
		Level:      "info",
		Filename:   logFile,
		MaxSize:    1,
		MaxBackups: 5,
		MaxAge:     5,
		Compress:   true,
		IsDev:      utils.GetEnv("APP_ENV", "development"),
	})

	if err := godotenv.Load(filepath.Join(rootDir, ".env")); err != nil {
		loggers.Log.Warn().Msg("No .env file found")
	} else {
		loggers.Log.Info().Msg("Load successfully .env in worker")
	}

	// Initialize the configuration
	config := config.NewConfig()

	worker := newWorker(config)
	if worker == nil {
		loggers.Log.Fatal().Msg("Failed to create worker")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	defer stop()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := worker.Start(ctx); err != nil && err != context.Canceled {
			loggers.Log.Error().Err(err).Msg("Worker failed to start")
		}
	}()
	<-ctx.Done()
	loggers.Log.Info().Msg("Received shutdown signal")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	if err := worker.Shutdown(shutdownCtx); err != nil {
		loggers.Log.Error().Err(err).Msg("Shutdown failed")
	}
	wg.Wait()
	loggers.Log.Info().Msg("Main process teminated")
}
