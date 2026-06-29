package handlers

import "testing"

func TestEchoText(t *testing.T) {
	t.Parallel()

	input := "hello world"
	if got := EchoText(input); got != input {
		t.Fatalf("EchoText(%q) = %q, want %q", input, got, input)
	}
}

func TestStartMessageNotEmpty(t *testing.T) {
	t.Parallel()

	if StartMessage == "" {
		t.Fatal("StartMessage must not be empty")
	}
}
