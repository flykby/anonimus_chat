package env

import "testing"

func TestBool(t *testing.T) {
	t.Setenv("TEST_BOOL", "true")
	if !Bool("TEST_BOOL") {
		t.Fatal("expected true")
	}
}

func TestSet(t *testing.T) {
	t.Setenv("TEST_SET", "value")
	if !Set("TEST_SET") {
		t.Fatal("expected set")
	}
	if Set("TEST_MISSING") {
		t.Fatal("expected unset")
	}
}
