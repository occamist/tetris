package game

import "testing"

func TestNESRandomizerSeeded(t *testing.T) {
	a, b := NewNESRandomizer(42), NewNESRandomizer(42)
	counts := map[Kind]int{}
	for i := range 7000 {
		x, y := a.NextKind(), b.NextKind()
		if x != y {
			t.Fatalf("same seed diverged at %d: %v vs %v", i, x, y)
		}
		if x < I || x > L {
			t.Fatalf("got invalid kind %d", x)
		}
		counts[x]++
	}
	for _, k := range allKinds() {
		if counts[k] < 700 {
			t.Errorf("%v appeared %d times in 7000, suspiciously rare", k, counts[k])
		}
	}
}
