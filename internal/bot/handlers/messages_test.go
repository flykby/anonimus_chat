package handlers

import (
	"testing"

	"github.com/flykby/anonimus_chat/internal/bot/registration"
)

func TestWelcomeTextNotEmpty(t *testing.T) {
	t.Parallel()

	if registration.WelcomeText == "" {
		t.Fatal("WelcomeText must not be empty")
	}
}
