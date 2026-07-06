package menu

import (
	"testing"
	"time"

	"github.com/flykby/anonimus_chat/internal/shared"
)

func TestProfileViewTextRUWithoutPremium(t *testing.T) {
	t.Parallel()

	text := ProfileViewText(ProfileViewData{
		PublicUUID: "abc-123",
		Age:        25,
		Gender:     "male",
		Seeking:    "female",
		Language:   "ru",
	}, shared.LanguageRU)

	if text != "Пользователь abc-123\n\nPremium: отсутствует\n\nАнкета:\nПол: Парень\nИщу: Девушку\nВозраст: 25\nЯзык: RU" {
		t.Fatalf("unexpected text:\n%s", text)
	}
}

func TestProfileViewTextENWithPremium(t *testing.T) {
	t.Parallel()

	expires := time.Date(2026, 7, 15, 12, 30, 0, 0, time.UTC)
	text := ProfileViewText(ProfileViewData{
		PublicUUID:       "uuid-42",
		Age:              30,
		Gender:           "female",
		Seeking:          "male",
		Language:         "en",
		PremiumActive:    true,
		PremiumExpiresAt: &expires,
	}, shared.LanguageEN)

	wantPremium := "active until 15.07.2026 12:30 UTC+0"
	if text != "User uuid-42\n\nPremium: "+wantPremium+"\n\nProfile:\nGender: Girl\nLooking for: Guy\nAge: 30\nLanguage: EN" {
		t.Fatalf("unexpected text:\n%s", text)
	}
}

func TestProfileViewButtonsPremiumActive(t *testing.T) {
	t.Parallel()

	labels := LabelsFor(shared.LanguageRU)
	buttons := ProfileViewButtons(labels, true)
	if buttons[0][0].Text != labels.ProfileExtendPremium {
		t.Fatalf("premium button = %q", buttons[0][0].Text)
	}
}
