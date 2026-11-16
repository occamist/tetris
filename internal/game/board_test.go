package game

import (
	"slices"
	"testing"
)

// fillRow fills a row with kind k except for the listed columns.
func fillRow(b *Board, row int, k Kind, gaps ...int) {
	for c := range Width {
		if !slices.Contains(gaps, c) {
			b[row][c] = k
		}
	}
}

func TestFits(t *testing.T) {
	var b Board
	b[19][4] = J

	tests := []struct {
		name  string
		cells [4]Point
		want  bool
	}{
		{"open space", [4]Point{{5, 0}, {5, 1}, {5, 2}, {5, 3}}, true},
		{"left wall", [4]Point{{5, -1}, {5, 0}, {5, 1}, {5, 2}}, false},
		{"right wall", [4]Point{{5, 7}, {5, 8}, {5, 9}, {5, 10}}, false},
		{"floor", [4]Point{{17, 0}, {18, 0}, {19, 0}, {20, 0}}, false},
		{"above top is free", [4]Point{{-2, 0}, {-1, 0}, {0, 0}, {1, 0}}, true},
		{"overlaps stack", [4]Point{{18, 4}, {19, 4}, {18, 5}, {19, 5}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := b.Fits(tt.cells); got != tt.want {
				t.Errorf("Fits(%v) = %v, want %v", tt.cells, got, tt.want)
			}
		})
	}
}

func TestPlace(t *testing.T) {
	var b Board
	p := NewPiece(O).Moved(18, 0) // cells at rows 18-19, cols 4-5
	if !b.Place(p) {
		t.Fatal("Place inside board reported top-out")
	}
	for _, c := range p.Cells() {
		if b[c.Row][c.Col] != O {
			t.Errorf("cell %v = %v, want O", c, b[c.Row][c.Col])
		}
	}

	var b2 Board
	high := NewPiece(I).Rotated(true) // vertical I at spawn reaches row -2
	if b2.Place(high) {
		t.Error("Place with cells above the top should report top-out")
	}
}

func TestFullRows(t *testing.T) {
	var b Board
	fillRow(&b, 19, S)
	fillRow(&b, 18, S, 3)
	fillRow(&b, 17, Z)
	if got, want := b.FullRows(), []int{17, 19}; !slices.Equal(got, want) {
		t.Errorf("FullRows() = %v, want %v", got, want)
	}
	var empty Board
	if got := empty.FullRows(); len(got) != 0 {
		t.Errorf("empty FullRows() = %v", got)
	}
}

func TestClearRowsDropsRowsAbove(t *testing.T) {
	var b Board
	fillRow(&b, 19, L)    // full: cleared
	fillRow(&b, 18, J, 0) // stays, should end on row 19
	fillRow(&b, 17, T)    // full: cleared
	b[16][2] = I          // stays, should end on row 18

	b.ClearRows([]int{17, 19})

	var want Board
	fillRow(&want, 19, J, 0)
	want[18][2] = I
	if b != want {
		t.Errorf("after ClearRows:\n%v\nwant:\n%v", b, want)
	}
}
