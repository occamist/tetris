package game

import (
	"slices"
	"testing"
)

func allKinds() []Kind { return []Kind{I, O, T, S, Z, J, L} }

func sorted(ps [4]Point) []Point {
	out := ps[:]
	slices.SortFunc(out, func(a, b Point) int {
		if a.Row != b.Row {
			return a.Row - b.Row
		}
		return a.Col - b.Col
	})
	return out
}

func TestOrientationsHaveFourDistinctCells(t *testing.T) {
	for _, k := range allKinds() {
		for rot, cells := range orientations[k] {
			seen := map[Point]bool{}
			for _, c := range cells {
				seen[c] = true
			}
			if len(seen) != 4 {
				t.Errorf("%v rotation %d has %d distinct cells, want 4", k, rot, len(seen))
			}
		}
	}
}

func TestRotationStateCounts(t *testing.T) {
	want := map[Kind]int{I: 2, O: 1, T: 4, S: 2, Z: 2, J: 4, L: 4}
	for k, n := range want {
		if got := len(orientations[k]); got != n {
			t.Errorf("%v has %d states, want %d", k, got, n)
		}
	}
}

// For T, J and L every next state must be the previous one turned clockwise
// about the pivot. On screen coordinates (rows down) that maps (r, c) to (c, -r).
func TestFourStatePiecesTurnClockwise(t *testing.T) {
	for _, k := range []Kind{T, J, L} {
		states := orientations[k]
		for i, cur := range states {
			var turned [4]Point
			for j, p := range cur {
				turned[j] = Point{p.Col, -p.Row}
			}
			next := states[(i+1)%len(states)]
			if !slices.Equal(sorted(turned), sorted(next)) {
				t.Errorf("%v state %d turned clockwise = %v, want state %d = %v", k, i, sorted(turned), (i+1)%len(states), sorted(next))
			}
		}
	}
}

func TestSpawnInsideBoard(t *testing.T) {
	var b Board
	for _, k := range allKinds() {
		p := NewPiece(k)
		if !b.Fits(p.Cells()) {
			t.Errorf("%v does not fit at spawn", k)
		}
		for _, c := range p.Cells() {
			if c.Row < 0 || c.Row > 1 {
				t.Errorf("%v spawns with a cell on row %d, want rows 0-1", k, c.Row)
			}
		}
	}
}

func TestRotatedWrapsBothWays(t *testing.T) {
	p := NewPiece(T)
	if got := p.Rotated(false).Rotation; got != 3 {
		t.Errorf("T counter-clockwise from spawn = %d, want 3", got)
	}
	q := p
	for range 4 {
		q = q.Rotated(true)
	}
	if q != p {
		t.Errorf("four clockwise turns = %+v, want %+v", q, p)
	}
	if got := NewPiece(O).Rotated(true); got != NewPiece(O) {
		t.Errorf("O rotation changed the piece: %+v", got)
	}
}

func TestKindString(t *testing.T) {
	if got := T.String(); got != "T" {
		t.Errorf("T.String() = %q", got)
	}
	if got := Kind(99).String(); got != "?" {
		t.Errorf("Kind(99).String() = %q", got)
	}
}
