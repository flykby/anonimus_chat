package deleteprofile_test

import (
	"testing"

	"github.com/flykby/anonimus_chat/internal/bot/deleteprofile"
	"github.com/flykby/anonimus_chat/internal/shared"
)

func TestConfirmMessagesLocalized(t *testing.T) {
	t.Parallel()

	if deleteprofile.Confirm1(shared.LanguageRU) == "" || deleteprofile.Confirm1(shared.LanguageEN) == "" {
		t.Fatal("confirm1 must not be empty")
	}
	if deleteprofile.Done(shared.LanguageEN) == "" {
		t.Fatal("done message must not be empty")
	}
}

func TestConfirmButtons(t *testing.T) {
	t.Parallel()

	buttons := deleteprofile.Confirm1Buttons(shared.LanguageRU)
	if len(buttons) != 1 || len(buttons[0]) != 2 {
		t.Fatalf("unexpected buttons: %+v", buttons)
	}
	if buttons[0][1].CallbackData != deleteprofile.CBCancel {
		t.Fatalf("cancel callback = %q", buttons[0][1].CallbackData)
	}
}
