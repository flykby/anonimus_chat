package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type EndDialogResponse struct {
	Status            string  `json:"status"`
	DialogID          int64   `json:"dialog_id"`
	PartnerTelegramID *int64  `json:"partner_telegram_id,omitempty"`
	PartnerLanguage   *string `json:"partner_language,omitempty"`
}

func (c *Client) EndDialog(ctx context.Context, dialogID, telegramID int64, reason string) (EndDialogResponse, error) {
	if reason == "" {
		reason = "user_confirmed"
	}
	body, err := json.Marshal(map[string]any{
		"telegram_id": telegramID,
		"reason":      reason,
	})
	if err != nil {
		return EndDialogResponse{}, err
	}

	url := fmt.Sprintf("%s/dialogs/%d/end", c.BaseURL, dialogID)
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
		return EndDialogResponse{}, fmt.Errorf("api end dialog: status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var result EndDialogResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return EndDialogResponse{}, err
	}
	return result, nil
}
