package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var (
	ErrRateLimited  = errors.New("rate limited")
	ErrPhotoLimit   = errors.New("photo limit")
	ErrInvalidRelay = errors.New("invalid relay")
)

type RelayResponse struct {
	Status            string `json:"status"`
	PartnerTelegramID int64  `json:"partner_telegram_id"`
	PartnerLanguage   string `json:"partner_language"`
	Kind              string `json:"kind"`
	Text              string `json:"text,omitempty"`
	TelegramFileID    string `json:"telegram_file_id,omitempty"`
}

type ReportResponse struct {
	Status   string `json:"status"`
	DialogID int64  `json:"dialog_id"`
}

func (c *Client) RelayDialog(ctx context.Context, dialogID, telegramID int64, kind, text, fileID string) (RelayResponse, error) {
	body, err := json.Marshal(map[string]any{
		"telegram_id":      telegramID,
		"kind":             kind,
		"text":             text,
		"telegram_file_id": fileID,
	})
	if err != nil {
		return RelayResponse{}, err
	}

	url := fmt.Sprintf("%s/dialogs/%d/relay", c.BaseURL, dialogID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return RelayResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return RelayResponse{}, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	switch resp.StatusCode {
	case http.StatusOK:
		var result RelayResponse
		if err := json.Unmarshal(respBody, &result); err != nil {
			return RelayResponse{}, err
		}
		return result, nil
	case http.StatusTooManyRequests:
		return RelayResponse{}, ErrRateLimited
	case http.StatusConflict:
		if strings.Contains(string(respBody), "photo_limit") {
			return RelayResponse{}, ErrPhotoLimit
		}
		return RelayResponse{}, fmt.Errorf("api relay: status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	case http.StatusBadRequest:
		if strings.Contains(string(respBody), "invalid_relay") {
			return RelayResponse{}, ErrInvalidRelay
		}
		return RelayResponse{}, fmt.Errorf("api relay: status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	default:
		return RelayResponse{}, fmt.Errorf("api relay: status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
}

func (c *Client) ReportDialog(ctx context.Context, dialogID, telegramID int64, reason string) (ReportResponse, error) {
	body, err := json.Marshal(map[string]any{
		"telegram_id": telegramID,
		"reason":      reason,
	})
	if err != nil {
		return ReportResponse{}, err
	}

	url := fmt.Sprintf("%s/dialogs/%d/report", c.BaseURL, dialogID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return ReportResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return ReportResponse{}, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode != http.StatusOK {
		return ReportResponse{}, fmt.Errorf("api report: status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var result ReportResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return ReportResponse{}, err
	}
	return result, nil
}

func (c *Client) BlockDialog(ctx context.Context, dialogID, telegramID int64) (EndDialogResponse, error) {
	body, err := json.Marshal(map[string]int64{"telegram_id": telegramID})
	if err != nil {
		return EndDialogResponse{}, err
	}

	url := fmt.Sprintf("%s/dialogs/%d/block", c.BaseURL, dialogID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return EndDialogResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return EndDialogResponse{}, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode != http.StatusOK {
		return EndDialogResponse{}, fmt.Errorf("api block: status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var result EndDialogResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return EndDialogResponse{}, err
	}
	return result, nil
}
