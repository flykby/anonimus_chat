package events

import "testing"

func TestAllEventTypesValid(t *testing.T) {
	t.Parallel()

	for _, eventType := range allTypes {
		if !eventType.Valid() {
			t.Fatalf("expected %q to be valid", eventType)
		}
	}
	if Type("unknown.event").Valid() {
		t.Fatal("expected unknown type to be invalid")
	}
}

func TestValidateUserRegisteredMeta(t *testing.T) {
	t.Parallel()

	if err := validateMetadata(TypeUserRegistered, UserRegisteredMeta{
		TelegramID: 1,
		Age:        25,
		Gender:     "male",
		Seeking:    "female",
		Language:   "ru",
	}); err != nil {
		t.Fatalf("valid metadata: %v", err)
	}

	if err := validateMetadata(TypeUserRegistered, UserRegisteredMeta{
		TelegramID: 0,
		Age:        25,
		Gender:     "male",
		Seeking:    "female",
		Language:   "ru",
	}); err == nil {
		t.Fatal("expected error for missing telegram_id")
	}
}

func TestValidateUserProfileUpdatedMeta(t *testing.T) {
	t.Parallel()

	if err := validateMetadata(TypeUserProfileUpdated, UserProfileUpdatedMeta{
		Changes: []ProfileFieldChange{
			{Field: "age", Old: "25", New: "26"},
		},
	}); err != nil {
		t.Fatalf("valid metadata: %v", err)
	}

	if err := validateMetadata(TypeUserProfileUpdated, UserProfileUpdatedMeta{}); err == nil {
		t.Fatal("expected error for empty changes")
	}
}

func TestValidateUserDeletedMeta(t *testing.T) {
	t.Parallel()

	if err := validateMetadata(TypeUserDeleted, UserDeletedMeta{Reason: "user_requested"}); err != nil {
		t.Fatalf("valid metadata: %v", err)
	}
	if err := validateMetadata(TypeUserDeleted, UserDeletedMeta{}); err != nil {
		t.Fatalf("empty reason allowed: %v", err)
	}
}

func TestValidateDialogStartedMeta(t *testing.T) {
	t.Parallel()

	if err := validateMetadata(TypeDialogStarted, DialogStartedMeta{
		Type:       "ai",
		MatchRoute: "m_seeks_f",
	}); err != nil {
		t.Fatalf("valid metadata: %v", err)
	}

	if err := validateMetadata(TypeDialogStarted, DialogStartedMeta{
		Type:       "invalid",
		MatchRoute: "m_seeks_f",
	}); err == nil {
		t.Fatal("expected error for invalid dialog type")
	}
}

func TestValidateDialogEndedMeta(t *testing.T) {
	t.Parallel()

	if err := validateMetadata(TypeDialogEnded, DialogEndedMeta{
		Reason:       "user_confirmed",
		DurationSec:  120,
		MessageCount: 10,
	}); err != nil {
		t.Fatalf("valid metadata: %v", err)
	}
}

func TestValidateMessageSentMeta(t *testing.T) {
	t.Parallel()

	if err := validateMetadata(TypeMessageSent, MessageSentMeta{ContentLength: 42}); err != nil {
		t.Fatalf("valid metadata: %v", err)
	}
	if err := validateMetadata(TypeMessageSent, MessageSentMeta{ContentLength: 0}); err == nil {
		t.Fatal("expected error for zero content length")
	}
}

func TestEmitRejectsUnknownType(t *testing.T) {
	t.Parallel()

	emitter := NewEmitter(nil)
	err := emitter.Emit(t.Context(), nil, Input{Type: Type("bad.type"), Metadata: map[string]any{}})
	if err == nil {
		t.Fatal("expected error for unknown event type")
	}
}
