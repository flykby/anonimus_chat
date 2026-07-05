package registration

import (
	"strings"
	"testing"

	"github.com/flykby/anonimus_chat/internal/redis/regdraft"
	"github.com/flykby/anonimus_chat/internal/shared"
)

func TestParseAgeValid(t *testing.T) {
	t.Parallel()

	age, err := ParseAge("25")
	if err != nil {
		t.Fatalf("ParseAge() err = %v", err)
	}
	if age != 25 {
		t.Fatalf("age = %d", age)
	}
}

func TestParseAgeInvalid(t *testing.T) {
	t.Parallel()

	if _, err := ParseAge("abc"); err == nil {
		t.Fatal("expected error for letters")
	}
	if !IsTooYoung("17") {
		t.Fatal("expected too young")
	}
	if _, err := ParseAge("17"); err == nil {
		t.Fatal("expected error for under 18")
	}
}

func TestConfirmationTextContainsFields(t *testing.T) {
	t.Parallel()

	text := ConfirmationText(regdraft.Draft{
		Age:      25,
		Gender:   shared.GenderMale,
		Seeking:  shared.GenderFemale,
		Language: shared.LanguageRU,
	})
	for _, part := range []string{"25", "Парень", "Девушку", "Русский"} {
		if !strings.Contains(text, part) {
			t.Fatalf("confirmation missing %q: %s", part, text)
		}
	}
}
