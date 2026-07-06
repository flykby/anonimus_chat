package match

import (
	"testing"

	"github.com/flykby/anonimus_chat/internal/shared"
)

func TestP2PQueueFor(t *testing.T) {
	t.Parallel()

	cases := []struct {
		gender  shared.Gender
		seeking shared.Gender
		kind    p2pQueueKind
		genderQ shared.Gender
	}{
		{shared.GenderMale, shared.GenderMale, p2pQueueSameGender, shared.GenderMale},
		{shared.GenderFemale, shared.GenderMale, p2pQueueHeteroFemale, shared.GenderFemale},
	}

	for _, tc := range cases {
		target := p2pQueueFor(tc.gender, tc.seeking)
		if target.kind != tc.kind || target.gender != tc.genderQ {
			t.Fatalf("p2pQueueFor(%q,%q) = %+v, want kind=%d gender=%q", tc.gender, tc.seeking, target, tc.kind, tc.genderQ)
		}
	}
}
