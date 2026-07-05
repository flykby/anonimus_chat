package menu

import "testing"

func TestActionForTextRU(t *testing.T) {
	t.Parallel()

	action, lang := ActionForText("Начать разговор")
	if action != ActionStartChat || lang != ParseLanguage("ru") {
		t.Fatalf("got action=%v lang=%v", action, lang)
	}
}

func TestActionForTextEN(t *testing.T) {
	t.Parallel()

	action, _ := ActionForText("Profile")
	if action != ActionProfile {
		t.Fatalf("got action=%v", action)
	}
}

func TestMainKeyboardRows(t *testing.T) {
	t.Parallel()

	kb := MainKeyboard(LabelsFor(ParseLanguage("ru")))
	if len(kb.Keyboard) != 2 {
		t.Fatalf("rows = %d", len(kb.Keyboard))
	}
	if !kb.ResizeKeyboard {
		t.Fatal("expected resize keyboard")
	}
}

func TestDialogKeyboardSingleButton(t *testing.T) {
	t.Parallel()

	kb := DialogKeyboard(LabelsFor(ParseLanguage("ru")))
	if len(kb.Keyboard) != 1 || kb.Keyboard[0][0].Text != "Завершить диалог" {
		t.Fatalf("unexpected keyboard: %+v", kb.Keyboard)
	}
}
