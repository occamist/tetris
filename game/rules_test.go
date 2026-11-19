package game

import "testing"

func TestFramesPerRow(t *testing.T) {
	tests := map[int]int{
		-1: 48, 0: 48, 1: 43, 8: 8, 9: 6, 10: 5, 12: 5, 13: 4, 16: 3, 18: 3,
		19: 2, 28: 2, 29: 1, 99: 1,
	}
	for level, want := range tests {
		if got := FramesPerRow(level); got != want {
			t.Errorf("FramesPerRow(%d) = %d, want %d", level, got, want)
		}
	}
}

func TestLineScore(t *testing.T) {
	tests := []struct{ level, lines, want int }{
		{0, 0, 0},
		{0, 1, 40},
		{0, 2, 100},
		{0, 3, 300},
		{0, 4, 1200},
		{9, 4, 12000},
		{18, 1, 760},
		{0, 5, 0},
		{0, -1, 0},
	}
	for _, tt := range tests {
		if got := LineScore(tt.level, tt.lines); got != tt.want {
			t.Errorf("LineScore(%d, %d) = %d, want %d", tt.level, tt.lines, got, tt.want)
		}
	}
}

func TestLinesForNextLevel(t *testing.T) {
	tests := []struct{ start, level, want int }{
		{0, 0, 10},
		{0, 1, 20},
		{0, 5, 60},
		{5, 5, 60},
		{9, 9, 100},
		{10, 10, 100},
		{15, 15, 100},
		{16, 16, 110},
		{18, 18, 130},
		{18, 19, 140},
		{19, 19, 140},
	}
	for _, tt := range tests {
		if got := LinesForNextLevel(tt.start, tt.level); got != tt.want {
			t.Errorf("LinesForNextLevel(%d, %d) = %d, want %d", tt.start, tt.level, got, tt.want)
		}
	}
}

func TestEntryDelay(t *testing.T) {
	tests := map[int]int{
		19: 10, 18: 10, // bottom two rows
		17: 12, 14: 12,
		13: 14, 10: 14,
		9: 16, 6: 16,
		5: 18, 0: 18, -1: 18,
	}
	for row, want := range tests {
		if got := EntryDelay(row); got != want {
			t.Errorf("EntryDelay(%d) = %d, want %d", row, got, want)
		}
	}
}
