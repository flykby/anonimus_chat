package menu

import (
	"strings"
	"testing"

	"github.com/flykby/anonimus_chat/internal/shared"
)

func TestQueueWaitingTextRUGendered(t *testing.T) {
	t.Parallel()

	maleText := QueueWaitingText(12, shared.GenderMale, shared.LanguageRU)
	if !strings.Contains(maleText, "лучшего") {
		t.Fatalf("male text = %q", maleText)
	}

	femaleText := QueueWaitingText(12, shared.GenderFemale, shared.LanguageRU)
	if !strings.Contains(femaleText, "лучшую") {
		t.Fatalf("female text = %q", femaleText)
	}
}

func TestQueueWaitingTextUsesCount(t *testing.T) {
	t.Parallel()

	text := QueueWaitingText(42, shared.GenderMale, shared.LanguageEN)
	if !strings.Contains(text, "42") {
		t.Fatalf("text = %q", text)
	}
}
