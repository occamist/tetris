package game

// Board size. The NES playfield has no hidden rows above the visible 20.
const (
	Width  = 10
	Height = 20
)

// Board is the grid of locked cells, indexed [row][col] with row 0 at the top.
type Board [Height][Width]Kind

// Fits reports whether every cell is inside the walls and floor and not
// already occupied. Cells above the top edge count as free, so a piece may
// rotate partly out of the top of the playfield.
func (b *Board) Fits(cells [4]Point) bool {
	for _, c := range cells {
		if c.Col < 0 || c.Col >= Width || c.Row >= Height {
			return false
		}
		if c.Row >= 0 && b[c.Row][c.Col] != Empty {
			return false
		}
	}
	return true
}

// Place writes the piece's cells into the board. It reports false if any cell
// was above the top edge, which means the stack has topped out.
func (b *Board) Place(p Piece) bool {
	ok := true
	for _, c := range p.Cells() {
		if c.Row < 0 {
			ok = false
			continue
		}
		b[c.Row][c.Col] = p.Kind
	}
	return ok
}

// FullRows returns the indexes of all completely filled rows, top to bottom.
func (b *Board) FullRows() []int {
	var rows []int
	for r := range Height {
		full := true
		for c := range Width {
			if b[r][c] == Empty {
				full = false
				break
			}
		}
		if full {
			rows = append(rows, r)
		}
	}
	return rows
}

// ClearRows removes the given rows and drops everything above them down.
func (b *Board) ClearRows(rows []int) {
	remove := make(map[int]bool, len(rows))
	for _, r := range rows {
		remove[r] = true
	}
	var next Board
	dst := Height - 1
	for src := Height - 1; src >= 0; src-- {
		if remove[src] {
			continue
		}
		next[dst] = b[src]
		dst--
	}
	*b = next
}
