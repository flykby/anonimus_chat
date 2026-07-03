package shared

import "testing"

func TestGenderValues(t *testing.T) {
	t.Parallel()

	if GenderMale != "male" || GenderFemale != "female" {
		t.Fatal("unexpected gender constants")
	}
}
