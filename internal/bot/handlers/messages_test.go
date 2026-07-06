package handlers

import (
	"testing"

	"github.com/flykby/anonimus_chat/internal/bot/registration"
	"github.com/flykby/anonimus_chat/internal/shared"
)

func TestWelcomeTextNotEmpty(t *testing.T) {
	t.Parallel()

	if registration.WelcomeText(shared.LanguageRU) == "" {
		t.Fatal("WelcomeText must not be empty")
	}
	if registration.WelcomeText(shared.LanguageEN) == "" {
		t.Fatal("WelcomeText EN must not be empty")
	}
}
