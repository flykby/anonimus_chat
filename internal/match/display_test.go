package match

import "testing"

func TestSyntheticDisplayCountRange(t *testing.T) {
	t.Parallel()

	for i := 0; i < 50; i++ {
		n := SyntheticDisplayCount()
		if n < syntheticCountMin || n > syntheticCountMax {
			t.Fatalf("count = %d, want %d-%d", n, syntheticCountMin, syntheticCountMax)
		}
	}
}
