package events

import (
	"encoding/json"
	"fmt"
)

type UserRegisteredMeta struct {
	TelegramID int64  `json:"telegram_id"`
	Age        int16  `json:"age"`
	Gender     string `json:"gender"`
	Seeking    string `json:"seeking"`
	Language   string `json:"language"`
}

type UserProfileUpdatedMeta struct {
	Fields []string `json:"fields"`
}

type UserDeletedMeta struct {
	Reason string `json:"reason,omitempty"`
}

type DialogStartedMeta struct {
	Type       string `json:"type"`
	PersonaID  *int64 `json:"persona_id,omitempty"`
	MatchRoute string `json:"match_route"`
}

type DialogEndedMeta struct {
	Reason       string `json:"reason"`
	DurationSec  int    `json:"duration_sec"`
	MessageCount int    `json:"message_count"`
}

type MessageSentMeta struct {
	ContentLength int `json:"content_length"`
}

type MessageReceivedMeta struct {
	Source        string `json:"source"`
	ContentLength int    `json:"content_length"`
}

type PhotoRequestedMeta struct {
	Intent string `json:"intent,omitempty"`
}

type PhotoSentMeta struct {
	PhotoID    int64  `json:"photo_id"`
	NsfwLevel  string `json:"nsfw_level"`
	WasBlurred bool   `json:"was_blurred"`
}

type PhotoUnlockedMeta struct {
	PhotoID   int64 `json:"photo_id"`
	StarsPaid int   `json:"stars_paid"`
}

type PremiumPurchasedMeta struct {
	ExpiresAt string `json:"expires_at"`
	StarsPaid int    `json:"stars_paid"`
}

type PremiumExpiredMeta struct {
	SubscriptionID int64 `json:"subscription_id,omitempty"`
}

type QueueEnteredMeta struct {
	Route   string `json:"route"`
	Gender  string `json:"gender"`
	Seeking string `json:"seeking"`
}

type QueueMatchedMeta struct {
	Route   string `json:"route"`
	WaitSec int    `json:"wait_sec,omitempty"`
}

func marshalMetadata(eventType Type, metadata any) ([]byte, error) {
	if metadata == nil {
		metadata = map[string]any{}
	}
	if err := validateMetadata(eventType, metadata); err != nil {
		return nil, err
	}
	raw, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("marshal metadata: %w", err)
	}
	return raw, nil
}

func validateMetadata(eventType Type, metadata any) error {
	switch eventType {
	case TypeUserRegistered:
		return decodeValidate[UserRegisteredMeta](metadata, validateUserRegistered)
	case TypeUserProfileUpdated:
		return decodeValidate[UserProfileUpdatedMeta](metadata, validateUserProfileUpdated)
	case TypeUserDeleted:
		return decodeValidate[UserDeletedMeta](metadata, validateUserDeleted)
	case TypeDialogStarted:
		return decodeValidate[DialogStartedMeta](metadata, validateDialogStarted)
	case TypeDialogEnded:
		return decodeValidate[DialogEndedMeta](metadata, validateDialogEnded)
	case TypeMessageSent:
		return decodeValidate[MessageSentMeta](metadata, validateMessageSent)
	case TypeMessageReceived:
		return decodeValidate[MessageReceivedMeta](metadata, validateMessageReceived)
	case TypePhotoRequested:
		return decodeValidate[PhotoRequestedMeta](metadata, validatePhotoRequested)
	case TypePhotoSent:
		return decodeValidate[PhotoSentMeta](metadata, validatePhotoSent)
	case TypePhotoUnlocked:
		return decodeValidate[PhotoUnlockedMeta](metadata, validatePhotoUnlocked)
	case TypePremiumPurchased:
		return decodeValidate[PremiumPurchasedMeta](metadata, validatePremiumPurchased)
	case TypePremiumExpired:
		return decodeValidate[PremiumExpiredMeta](metadata, validatePremiumExpired)
	case TypeQueueEntered:
		return decodeValidate[QueueEnteredMeta](metadata, validateQueueEntered)
	case TypeQueueMatched:
		return decodeValidate[QueueMatchedMeta](metadata, validateQueueMatched)
	default:
		return fmt.Errorf("unknown event type %q", eventType)
	}
}

func decodeValidate[T any](metadata any, validate func(T) error) error {
	raw, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata for validation: %w", err)
	}
	var v T
	if err := json.Unmarshal(raw, &v); err != nil {
		return fmt.Errorf("invalid metadata shape: %w", err)
	}
	return validate(v)
}

func validateUserRegistered(m UserRegisteredMeta) error {
	if m.TelegramID <= 0 {
		return fmt.Errorf("telegram_id required")
	}
	if m.Age < 18 || m.Age > 99 {
		return fmt.Errorf("age must be 18-99")
	}
	if m.Gender == "" || m.Seeking == "" || m.Language == "" {
		return fmt.Errorf("gender, seeking, and language required")
	}
	return nil
}

func validateUserProfileUpdated(m UserProfileUpdatedMeta) error {
	if len(m.Fields) == 0 {
		return fmt.Errorf("fields required")
	}
	return nil
}

func validateUserDeleted(_ UserDeletedMeta) error {
	return nil
}

func validateDialogStarted(m DialogStartedMeta) error {
	if m.Type != "ai" && m.Type != "p2p" {
		return fmt.Errorf("type must be ai or p2p")
	}
	if m.MatchRoute == "" {
		return fmt.Errorf("match_route required")
	}
	return nil
}

func validateDialogEnded(m DialogEndedMeta) error {
	if m.Reason == "" {
		return fmt.Errorf("reason required")
	}
	if m.DurationSec < 0 || m.MessageCount < 0 {
		return fmt.Errorf("duration_sec and message_count must be >= 0")
	}
	return nil
}

func validateMessageSent(m MessageSentMeta) error {
	if m.ContentLength <= 0 {
		return fmt.Errorf("content_length must be > 0")
	}
	return nil
}

func validateMessageReceived(m MessageReceivedMeta) error {
	if m.Source != "ai" && m.Source != "partner" {
		return fmt.Errorf("source must be ai or partner")
	}
	if m.ContentLength <= 0 {
		return fmt.Errorf("content_length must be > 0")
	}
	return nil
}

func validatePhotoRequested(_ PhotoRequestedMeta) error {
	return nil
}

func validatePhotoSent(m PhotoSentMeta) error {
	if m.PhotoID <= 0 {
		return fmt.Errorf("photo_id required")
	}
	if m.NsfwLevel != "safe" && m.NsfwLevel != "adult" {
		return fmt.Errorf("nsfw_level must be safe or adult")
	}
	return nil
}

func validatePhotoUnlocked(m PhotoUnlockedMeta) error {
	if m.PhotoID <= 0 {
		return fmt.Errorf("photo_id required")
	}
	if m.StarsPaid < 0 {
		return fmt.Errorf("stars_paid must be >= 0")
	}
	return nil
}

func validatePremiumPurchased(m PremiumPurchasedMeta) error {
	if m.ExpiresAt == "" {
		return fmt.Errorf("expires_at required")
	}
	if m.StarsPaid < 0 {
		return fmt.Errorf("stars_paid must be >= 0")
	}
	return nil
}

func validatePremiumExpired(_ PremiumExpiredMeta) error {
	return nil
}

func validateQueueEntered(m QueueEnteredMeta) error {
	if m.Route != "ai" && m.Route != "p2p" {
		return fmt.Errorf("route must be ai or p2p")
	}
	if m.Gender == "" || m.Seeking == "" {
		return fmt.Errorf("gender and seeking required")
	}
	return nil
}

func validateQueueMatched(m QueueMatchedMeta) error {
	if m.Route != "ai" && m.Route != "p2p" {
		return fmt.Errorf("route must be ai or p2p")
	}
	if m.WaitSec < 0 {
		return fmt.Errorf("wait_sec must be >= 0")
	}
	return nil
}
