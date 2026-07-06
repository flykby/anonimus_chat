package keys

import "testing"

func TestKeyFormats(t *testing.T) {
	t.Parallel()

	if got := P2PQueue("male"); got != "anonimus:queue:p2p:male" {
		t.Fatalf("P2PQueue = %q", got)
	}
	if got := HeteroQueue("female"); got != "anonimus:queue:hetero:female" {
		t.Fatalf("HeteroQueue = %q", got)
	}
	if got := Session(42); got != "anonimus:session:42" {
		t.Fatalf("Session = %q", got)
	}
	if got := FSM(99); got != "anonimus:fsm:99" {
		t.Fatalf("FSM = %q", got)
	}
}
