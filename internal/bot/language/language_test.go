package language_test

import (
	"testing"

	"github.com/flykby/anonimus_chat/internal/bot/language"
	"github.com/flykby/anonimus_chat/internal/shared"
)

func TestParseCallback(t *testing.T) {
	t.Parallel()

	lang, ok := language.ParseCallback(language.CBRU)
	if !ok || lang != shared.LanguageRU {
		t.Fatalf("ru = %q, ok=%v", lang, ok)
	}

	lang, ok = language.ParseCallback(language.CBEN)
	if !ok || lang != shared.LanguageEN {
		t.Fatalf("en = %q, ok=%v", lang, ok)
	}

	_, ok = language.ParseCallback("lang:xx")
	if ok {
		t.Fatal("expected unknown callback to fail")
	}
}

func TestPromptLocalized(t *testing.T) {
	t.Parallel()

	if language.Prompt(shared.LanguageRU) == "" || language.Prompt(shared.LanguageEN) == "" {
		t.Fatal("prompt must not be empty")
	}
}

func TestChangedUsesLanguageCode(t *testing.T) {
	t.Parallel()

	got := language.Changed(shared.LanguageEN)
	if got == "" {
		t.Fatal("changed message must not be empty")
	}
}
