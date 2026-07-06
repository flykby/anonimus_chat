package rules_test

import (
	"strings"
	"testing"

	"github.com/flykby/anonimus_chat/internal/bot/rules"
	"github.com/flykby/anonimus_chat/internal/shared"
)

func TestMessagesRU(t *testing.T) {
	t.Parallel()

	msgs, err := rules.Messages(shared.LanguageRU)
	if err != nil {
		t.Fatalf("Messages(): %v", err)
	}
	if len(msgs) == 0 {
		t.Fatal("expected at least one message")
	}
	for i, msg := range msgs {
		if len(msg) > 4096 {
			t.Fatalf("chunk %d length = %d", i, len(msg))
		}
	}
	if !strings.Contains(msgs[0], "Правила использования") {
		t.Fatalf("unexpected start: %q", msgs[0][:40])
	}
}

func TestMessagesEN(t *testing.T) {
	t.Parallel()

	msgs, err := rules.Messages(shared.LanguageEN)
	if err != nil {
		t.Fatalf("Messages(): %v", err)
	}
	if !strings.Contains(msgs[0], "Anonimus Chat Rules") {
		t.Fatalf("unexpected start: %q", msgs[0][:40])
	}
}

func TestRulesVersion(t *testing.T) {
	t.Parallel()

	if rules.RulesVersion == "" {
		t.Fatal("RulesVersion must not be empty")
	}
}
