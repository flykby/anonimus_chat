package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const secretTokenHeader = "X-Telegram-Bot-Api-Secret-Token"

type Webhook struct {
	bot         *bot.Bot
	secretToken string
	logger      *slog.Logger
}

func New(b *bot.Bot, secretToken string, logger *slog.Logger) *Webhook {
	return &Webhook{
		bot:         b,
		secretToken: secretToken,
		logger:      logger,
	}
}

func (w *Webhook) Register(ctx context.Context, webhookURL, certPath string) error {
	params := &bot.SetWebhookParams{
		URL:            webhookURL,
		AllowedUpdates: []string{"message", "callback_query", "pre_checkout_query"},
	}

	if w.secretToken != "" {
		params.SecretToken = w.secretToken
	}

	if certPath != "" {
		certFile, err := os.Open(certPath)
		if err != nil {
			return fmt.Errorf("open certificate file: %w", err)
		}
		defer certFile.Close()
		params.Certificate = &models.InputFileUpload{
			Filename: "webhook.pem",
			Data:     certFile,
		}
	}

	ok, err := w.bot.SetWebhook(ctx, params)
	if err != nil {
		return fmt.Errorf("set webhook: %w", err)
	}
	if !ok {
		return fmt.Errorf("set webhook returned false")
	}

	return nil
}

func (w *Webhook) Handler(rw http.ResponseWriter, r *http.Request) {
	if w.secretToken != "" {
		token := r.Header.Get(secretTokenHeader)
		if token != w.secretToken {
			w.logger.Warn("invalid secret token", "remote_addr", r.RemoteAddr)
			http.Error(rw, "Forbidden", http.StatusForbidden)
			return
		}
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.logger.Error("read webhook body failed", "err", err)
		http.Error(rw, "Bad Request", http.StatusBadRequest)
		return
	}

	var update models.Update
	if err := json.Unmarshal(body, &update); err != nil {
		w.logger.Error("parse webhook update failed", "err", err)
		http.Error(rw, "Bad Request", http.StatusBadRequest)
		return
	}

	w.logger.Debug("webhook update received", "update_id", update.ID)

	w.bot.ProcessUpdate(r.Context(), &update)

	rw.WriteHeader(http.StatusOK)
}

func (w *Webhook) Delete(ctx context.Context) error {
	ok, err := w.bot.DeleteWebhook(ctx, &bot.DeleteWebhookParams{})
	if err != nil {
		return fmt.Errorf("delete webhook: %w", err)
	}
	if !ok {
		return fmt.Errorf("delete webhook returned false")
	}
	return nil
}
