package edit_test

import (
	"testing"

	"github.com/flykby/anonimus_chat/internal/bot/edit"
	"github.com/flykby/anonimus_chat/internal/shared"
)

func TestParseGenderCallback(t *testing.T) {
	t.Parallel()

	g, ok := edit.ParseGenderCallback(edit.CBGenderMale)
	if !ok || g != shared.GenderMale {
		t.Fatalf("male = %q, ok=%v", g, ok)
	}

	g, ok = edit.ParseGenderCallback(edit.CBSeekingFemale)
	if !ok || g != shared.GenderFemale {
		t.Fatalf("female = %q, ok=%v", g, ok)
	}

	_, ok = edit.ParseGenderCallback("edit:unknown")
	if ok {
		t.Fatal("expected unknown callback to fail")
	}
}

func TestIsEditState(t *testing.T) {
	t.Parallel()

	if !edit.IsEditState(edit.StateAge) {
		t.Fatal("expected age to be edit state")
	}
	if edit.IsEditState("reg:age") {
		t.Fatal("registration state must not be edit state")
	}
}

func TestMenuTitleLocalized(t *testing.T) {
	t.Parallel()

	if edit.MenuTitle(shared.LanguageRU) == "" || edit.MenuTitle(shared.LanguageEN) == "" {
		t.Fatal("menu title must not be empty")
	}
}
