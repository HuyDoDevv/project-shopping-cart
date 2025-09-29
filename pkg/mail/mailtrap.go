package mail

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gin/user-management-api/internal/utils"
	"gin/user-management-api/pkg/loggers"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type MailTrapConfig struct {
	MailSender     string
	NameSender     string
	MailtrapUrl    string
	MailtrapApiKey string
}

type MailTrapProvider struct {
	client *http.Client
	config *MailTrapConfig
	logger *zerolog.Logger
}

func NewMailTrapProvider(config *MailConfig) (EmailProviderService, error) {
	mailtrapCfg, ok := config.ProviderConfig["mailtrap"].(map[string]any)
	if !ok {
		return nil, utils.NewError(utils.InternalServerError, "Invalid or missing Mailtrap configuaretion")
	}
	return &MailTrapProvider{
		client: &http.Client{Timeout: config.Timeout},
		config: &MailTrapConfig{
			MailSender:     mailtrapCfg["mail_sender"].(string),
			NameSender:     mailtrapCfg["name_sender"].(string),
			MailtrapUrl:    mailtrapCfg["mailtrap_url"].(string),
			MailtrapApiKey: mailtrapCfg["mailtrap_api_key"].(string),
		},
		logger: config.Logger,
	}, nil
}

func (p *MailTrapProvider) SendMail(ctx context.Context, email *Email) error {
	traceID := loggers.GetTraceID(ctx)
	start_time := time.Now()

	time.Sleep(5 * time.Second)

	email.From = Address{
		Email: p.config.MailSender,
		Name:  p.config.NameSender,
	}

	payload, err := json.Marshal(email)
	if err != nil {
		return utils.WrapError(utils.InternalServerError, "Failed to marshal email", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.config.MailtrapUrl, bytes.NewReader(payload))
	if err != nil {
		return utils.WrapError(utils.InternalServerError, "Failed to create request", err)
	}
	req.Header.Add("Authorization", "Bearer "+p.config.MailtrapApiKey)
	req.Header.Add("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		p.logger.Error().Str("trace_id", traceID).
			Dur("duration", time.Since(start_time)).
			Str("operation", "send_mail").
			Err(err).
			Msg("Failed to send request")
		return utils.WrapError(utils.InternalServerError, "Failed to send request", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		p.logger.Error().Str("trace_id", traceID).
			Dur("duration", time.Since(start_time)).
			Str("operation", "send_mail").
			Int("status_code", resp.StatusCode).
			Str("response_body", string(body)).
			Msg("Enexpected response from mailtrap")
		return utils.NewError(utils.InternalServerError, fmt.Sprintf("Unexpected response from mailtrap with code %d: %s", resp.StatusCode, string(body)))
	}

	return nil
}
