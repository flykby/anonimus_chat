package shared

import "testing"

func TestEnumValues(t *testing.T) {
	t.Parallel()

	if GenderMale != "male" || DialogTypeAI != "ai" || MessageRoleUser != "user" {
		t.Fatal("unexpected enum constants")
	}
}

func TestUserSoftDeleteField(t *testing.T) {
	t.Parallel()

	u := User{TelegramID: 123}
	if u.DeletedAt != nil {
		t.Fatal("expected nil deleted_at by default")
	}
}
