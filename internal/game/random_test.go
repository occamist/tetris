package game

import "testing"

// scriptedIntN returns the given rolls in order.
func scriptedIntN(t *testing.T, rolls ...int) func(int) int {
	t.Helper()
	return func(n int) int {
		if len(rolls) == 0 {
			t.Fatal("randomizer asked for more rolls than scripted")
		}
		r := rolls[0]
		rolls = rolls[1:]
		if r >= n {
			t.Fatalf("scripted roll %d out of range [0,%d)", r, n)
		}
		return r
	}
}

func TestNESRandomizerRerolls(t *testing.T) {
	// Rolls: T, then T again (repeat, reroll to S), then the unused 8th
	// option (reroll to I), then O.
	r := newNESRandomizer(scriptedIntN(t,
		2,    // T
		2, 3, // T repeats -> reroll S
		7, 0, // 8th option -> reroll I
		1, // O
	))
	want := []Kind{T, S, I, O}
	for i, w := range want {
		if got := r.Next(); got != w {
			t.Fatalf("piece %d = %v, want %v", i, got, w)
		}
	}
}

func TestNESRandomizerRerollMayRepeat(t *testing.T) {
	// The reroll is accepted as-is, so a repeat can still happen.
	r := newNESRandomizer(scriptedIntN(t, 4, 4, 4))
	if got := r.Next(); got != Z {
		t.Fatalf("first = %v, want Z", got)
	}
	if got := r.Next(); got != Z {
		t.Fatalf("second = %v, want Z after reroll", got)
	}
}

func TestNESRandomizerSeeded(t *testing.T) {
	a, b := NewNESRandomizer(42), NewNESRandomizer(42)
	counts := map[Kind]int{}
	for i := range 7000 {
		x, y := a.Next(), b.Next()
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
