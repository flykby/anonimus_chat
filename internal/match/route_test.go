package match

import (
	"testing"

	"github.com/flykby/anonimus_chat/internal/shared"
)

func TestResolveRoutes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		gender, seeking shared.Gender
		wantRoute       Route
		wantMatchRoute  string
	}{
		{shared.GenderMale, shared.GenderFemale, RouteAI, "m_seeks_f"},
		{shared.GenderMale, shared.GenderMale, RouteP2P, "m_seeks_m"},
		{shared.GenderFemale, shared.GenderFemale, RouteAI, "f_seeks_f"},
		{shared.GenderFemale, shared.GenderMale, RouteP2P, "f_seeks_m"},
	}

	for _, tt := range tests {
		got := Resolve(tt.gender, tt.seeking)
		if got.Route != tt.wantRoute || got.MatchRoute != tt.wantMatchRoute {
			t.Fatalf("Resolve(%q, %q) = %+v, want route=%q match_route=%q",
				tt.gender, tt.seeking, got, tt.wantRoute, tt.wantMatchRoute)
		}
	}
}
