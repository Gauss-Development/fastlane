package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"notification-service/pkg/logger"
)

const resendEndpoint = "https://api.resend.com/emails"

// Sender delivers transactional email through Resend's HTTP API. No SDK —
// mirrors the raw-HTTP style of the other external clients in this repo.
// With an empty API key it runs in degraded mode: emails are logged, not
// sent, so local/dev/CI work without secrets.
type Sender struct {
	apiKey     string
	from       string
	httpClient *http.Client
	logger     *logger.Logger
}

func NewSender(apiKey, from string, logger *logger.Logger) *Sender {
	return &Sender{
		apiKey:     apiKey,
		from:       from,
		httpClient: &http.Client{Timeout: 15 * time.Second},
		logger:     logger,
	}
}

func (s *Sender) Enabled() bool {
	return s.apiKey != ""
}

type sendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

func (s *Sender) Send(ctx context.Context, to, subject, html string) error {
	if to == "" {
		return fmt.Errorf("email recipient is empty")
	}

	if !s.Enabled() {
		s.logger.Info(fmt.Sprintf("email (degraded mode, not sent) to=%s subject=%q", to, subject))
		return nil
	}

	body, err := json.Marshal(sendRequest{
		From:    s.from,
		To:      []string{to},
		Subject: subject,
		HTML:    html,
	})
	if err != nil {
		return fmt.Errorf("marshal resend request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, resendEndpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build resend request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send email via resend: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("resend returned %d: %s", resp.StatusCode, string(detail))
	}

	s.logger.Info(fmt.Sprintf("email sent to=%s subject=%q", to, subject))
	return nil
}
