package shared

import "time"

type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
)

type Language string

const (
	LanguageRU Language = "ru"
	LanguageEN Language = "en"
)

type NsfwLevel string

const (
	NsfwLevelSafe  NsfwLevel = "safe"
	NsfwLevelAdult NsfwLevel = "adult"
)

type DialogType string

const (
	DialogTypeAI  DialogType = "ai"
	DialogTypeP2P DialogType = "p2p"
)

type MessageRole string

const (
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
	MessageRoleSystem    MessageRole = "system"
)

type User struct {
	ID         int64
	TelegramID int64
	PublicUUID string
	CreatedAt  time.Time
	DeletedAt  *time.Time
}

type Profile struct {
	UserID   int64
	Gender   Gender
	Seeking  Gender
	Age      int16
	Language Language
}

type Persona struct {
	ID            int64
	Name          string
	Gender        Gender
	PromptVersion string
	SystemPrompt  string
	Active        bool
	CreatedAt     time.Time
}

type Photo struct {
	ID               int64
	PersonaID        int64
	Tags             []string
	NsfwLevel        NsfwLevel
	TelegramFileID   string
	UnlockPriceStars int32
	CreatedAt        time.Time
}

type Dialog struct {
	ID            int64
	UserID        int64
	Type          DialogType
	PersonaID     *int64
	PartnerUserID *int64
	StartedAt     time.Time
	EndedAt       *time.Time
	EndReason     *string
}

type DialogMessage struct {
	ID        int64
	DialogID  int64
	Role      MessageRole
	Content   string
	CreatedAt time.Time
}

type Event struct {
	ID        int64
	UserID    *int64
	DialogID  *int64
	EventType string
	Metadata  map[string]any
	CreatedAt time.Time
}
