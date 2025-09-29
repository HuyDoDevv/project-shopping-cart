package mail

import (
	"context"
	"gin/user-management-api/internal/config"
	"gin/user-management-api/internal/utils"
	"gin/user-management-api/pkg/loggers"
	"time"

	"github.com/rs/zerolog"
)

type Email struct {
	From     Address   `json:"from"`
	To       []Address `json:"to"`
	Subject  string    `json:"subject"`
	Text     string    `json:"text"`
	Category string    `json:"category"`
}

type Address struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type MailConfig struct {
	ProviderConfig map[string]any
	ProviderType   ProviderType
	MaxRetries     int
	Timeout        time.Duration
	Logger         *zerolog.Logger
}
type MaiService struct {
	config   *MailConfig
	provider EmailProviderService
	logger   *zerolog.Logger
}

func NewMailService(cfg *config.Config, logger *zerolog.Logger, providerFactory ProviderFactory) (EmailProviderService, error) {
	config := &MailConfig{
		ProviderConfig: cfg.MailProviderConfig,
		ProviderType:   ProviderType(cfg.MailProviderType),
		MaxRetries:     3,
		Timeout:        10 * time.Second,
		Logger:         logger,
	}

	provider, err := providerFactory.CreateProvider(config)
	if err != nil {
		return nil, utils.WrapError(utils.InternalServerError, "Failed to create provider", err)
	}

	return &MaiService{
		config:   config,
		provider: provider,
		logger:   logger,
	}, nil
}

func (ms *MaiService) SendMail(ctx context.Context, email *Email) error {
	traceID := loggers.GetTraceID(ctx)
	start_time := time.Now()
	var lastErr error
	for attempt := 1; attempt <= ms.config.MaxRetries; attempt++ {
		startAttempt := time.Now()
		err := ms.provider.SendMail(ctx, email)
		if err == nil {
			ms.logger.Info().Str("trace_id", traceID).
				Dur("duration", time.Since(startAttempt)).
				Str("operation", "send_mail").
				Interface("to", email.To).
				Str("subject", email.Subject).
				Str("category", email.Category).
				Msg("Email send successfully")
			return nil
		}

		lastErr = err
		ms.logger.Warn().Str("trace_id", traceID).
			Dur("duration", time.Since(startAttempt)).
			Str("operation", "send_mail").
			Int("attempt", attempt).
			Err(err).
			Msg("Failed to send email, retrying")
		time.Sleep(time.Duration(attempt) * time.Second)
	}

	ms.logger.Error().Str("trace_id", traceID).
		Dur("duration", time.Since(start_time)).
		Str("operation", "send_mail").
		Int("attempt", ms.config.MaxRetries).
		Err(lastErr).
		Msg("Failed to send email after all retryies")

	return utils.WrapError(utils.InternalServerError, "Failed to send email after all retryies", lastErr)
}
