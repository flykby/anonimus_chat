package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type ProfileView struct {
	PublicUUID       string     `json:"public_uuid"`
	Age              int16      `json:"age"`
	Gender           string     `json:"gender"`
	Seeking          string     `json:"seeking"`
	Language         string     `json:"language"`
	PremiumActive    bool       `json:"premium_active"`
	PremiumExpiresAt *time.Time `json:"premium_expires_at,omitempty"`
}

func (c *Client) GetProfileView(ctx context.Context, telegramID int64) (ProfileView, error) {
	url := fmt.Sprintf("%s/users/by-telegram/%d/profile", c.BaseURL, telegramID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return ProfileView{}, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return ProfileView{}, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode == http.StatusNotFound {
		return ProfileView{}, ErrNotRegistered
	}
	if resp.StatusCode != http.StatusOK {
		return ProfileView{}, fmt.Errorf("api profile view: status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var view ProfileView
	if err := json.Unmarshal(body, &view); err != nil {
		return ProfileView{}, err
	}
	return view, nil
}

type UpdateProfileRequest struct {
	TelegramID int64   `json:"telegram_id"`
	Age        *int16  `json:"age,omitempty"`
	Gender     *string `json:"gender,omitempty"`
	Seeking    *string `json:"seeking,omitempty"`
}

func (c *Client) UpdateProfile(ctx context.Context, req UpdateProfileRequest) (ProfileView, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return ProfileView{}, err
	}

	url := c.BaseURL + "/users/me/profile"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, bytes.NewReader(body))
	if err != nil {
		return ProfileView{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return ProfileView{}, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode == http.StatusConflict {
		return ProfileView{}, ErrActiveDialog
	}
	if resp.StatusCode == http.StatusNotFound {
		return ProfileView{}, ErrNotRegistered
	}
	if resp.StatusCode != http.StatusOK {
		return ProfileView{}, fmt.Errorf("api update profile: status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var view ProfileView
	if err := json.Unmarshal(respBody, &view); err != nil {
		return ProfileView{}, err
	}
	return view, nil
}
