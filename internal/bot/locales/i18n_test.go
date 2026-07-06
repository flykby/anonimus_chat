package locales_test

import (
	"strings"
	"testing"

	"github.com/flykby/anonimus_chat/internal/bot/locales"
	"github.com/flykby/anonimus_chat/internal/shared"
)

func TestTRussianMenuTitle(t *testing.T) {
	t.Parallel()

	got := locales.T("menu.title", shared.LanguageRU, nil)
	if got != "Главное меню" {
		t.Fatalf("got %q", got)
	}
}

func TestTEnglishMenuTitle(t *testing.T) {
	t.Parallel()

	got := locales.T("menu.title", shared.LanguageEN, nil)
	if got != "Main menu" {
		t.Fatalf("got %q", got)
	}
}

func TestTParametrization(t *testing.T) {
	t.Parallel()

	got := locales.T("queue.waiting", shared.LanguageRU, map[string]string{
		"count":   "5",
		"seeking": "Ищем лучшего из них",
	})
	if !strings.Contains(got, "5") {
		t.Fatalf("expected count in %q", got)
	}
}

func TestFallbackToRussian(t *testing.T) {
	t.Parallel()

	got := locales.T("menu.title", shared.LanguageEN, nil)
	if got == "menu.title" {
		t.Fatal("expected resolved key")
	}
}

func TestMissingKeyReturnsKey(t *testing.T) {
	t.Parallel()

	got := locales.T("missing.key", shared.LanguageRU, nil)
	if got != "missing.key" {
		t.Fatalf("got %q", got)
	}
}
