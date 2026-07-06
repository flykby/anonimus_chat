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
	"time"
)

var (
	ErrNotRegistered           = errors.New("user not registered")
	ErrActiveDialog            = errors.New("active dialog exists")
	ErrPaymentAlreadyProcessed = errors.New("payment already processed")
)

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

type Profile struct {
	TelegramID       int64   `json:"telegram_id"`
	Age              int16   `json:"age"`
	Gender           string  `json:"gender"`
	Seeking          string  `json:"seeking"`
	Language         string  `json:"language"`
	ActiveDialog     bool    `json:"active_dialog"`
	ActiveDialogID   *int64  `json:"active_dialog_id,omitempty"`
	ActiveDialogType *string `json:"active_dialog_type,omitempty"`
}

type RegisterRequest struct {
	TelegramID int64  `json:"telegram_id"`
	Age        int16  `json:"age"`
	Gender     string `json:"gender"`
	Seeking    string `json:"seeking"`
	Language   string `json:"language"`
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) GetByTelegramID(ctx context.Context, telegramID int64) (Profile, bool, error) {
	url := fmt.Sprintf("%s/users/by-telegram/%d", c.BaseURL, telegramID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Profile{}, false, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return Profile{}, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return Profile{}, false, nil
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return Profile{}, false, fmt.Errorf("api get user: status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var profile Profile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return Profile{}, false, err
	}
	return profile, true, nil
}

func (c *Client) Register(ctx context.Context, req RegisterRequest) (Profile, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return Profile{}, err
	}

	url := c.BaseURL + "/users/register"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return Profile{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return Profile{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return Profile{}, fmt.Errorf("api register: status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var profile Profile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return Profile{}, err
	}
	return profile, nil
}

type DeleteProfileResponse struct {
	Status            string  `json:"status"`
	PartnerTelegramID *int64  `json:"partner_telegram_id,omitempty"`
	PartnerLanguage   *string `json:"partner_language,omitempty"`
}

func (c *Client) DeleteProfile(ctx context.Context, telegramID int64) (DeleteProfileResponse, error) {
	body, err := json.Marshal(map[string]int64{"telegram_id": telegramID})
	if err != nil {
		return DeleteProfileResponse{}, err
	}

	url := c.BaseURL + "/users/me"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, bytes.NewReader(body))
	if err != nil {
		return DeleteProfileResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return DeleteProfileResponse{}, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode == http.StatusNotFound {
		return DeleteProfileResponse{}, ErrNotRegistered
	}
	if resp.StatusCode != http.StatusOK {
		return DeleteProfileResponse{}, fmt.Errorf("api delete profile: status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var result DeleteProfileResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return DeleteProfileResponse{}, err
	}
	return result, nil
}
