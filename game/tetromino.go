package game

// Kind identifies a tetromino. The zero value is an empty board cell.
type Kind uint8

const (
	Empty Kind = iota
	I
	O
	T
	S
	Z
	J
	L
)

// NumKinds is the number of distinct tetrominoes.
const NumKinds = 7

// String returns the single-letter name of the kind.
func (k Kind) String() string {
	if k > NumKinds {
		return "?"
	}
	return string(" IOTSZJL"[k])
}

// Point is a (row, col) offset. Rows grow downwards.
type Point struct {
	Row, Col int
}

// orientations lists every rotation state of every kind as cell offsets from
// the piece's pivot. Index 0 is the spawn orientation and each next index is
// one clockwise turn. The states follow the NES rotation system: no wall
// kicks, and I, S and Z only have two states, so they don't drift as they turn.
var orientations = [NumKinds + 1][][4]Point{
	I: {
		{{0, -2}, {0, -1}, {0, 0}, {0, 1}},
		{{-2, 0}, {-1, 0}, {0, 0}, {1, 0}},
	},
	O: {
		{{0, -1}, {0, 0}, {1, -1}, {1, 0}},
	},
	T: {
		{{0, -1}, {0, 0}, {0, 1}, {1, 0}},
		{{-1, 0}, {0, -1}, {0, 0}, {1, 0}},
		{{0, -1}, {0, 0}, {0, 1}, {-1, 0}},
		{{-1, 0}, {0, 0}, {0, 1}, {1, 0}},
	},
	S: {
		{{0, 0}, {0, 1}, {1, -1}, {1, 0}},
		{{-1, 0}, {0, 0}, {0, 1}, {1, 1}},
	},
	Z: {
		{{0, -1}, {0, 0}, {1, 0}, {1, 1}},
		{{-1, 1}, {0, 0}, {0, 1}, {1, 0}},
	},
	J: {
		{{0, -1}, {0, 0}, {0, 1}, {1, 1}},
		{{-1, 0}, {0, 0}, {1, -1}, {1, 0}},
		{{-1, -1}, {0, -1}, {0, 0}, {0, 1}},
		{{-1, 0}, {-1, 1}, {0, 0}, {1, 0}},
	},
	L: {
		{{0, -1}, {0, 0}, {0, 1}, {1, -1}},
		{{-1, -1}, {-1, 0}, {0, 0}, {1, 0}},
		{{-1, 1}, {0, -1}, {0, 0}, {0, 1}},
		{{-1, 0}, {0, 0}, {1, 0}, {1, 1}},
	},
}

// SpawnRow and SpawnCol are where a new piece's pivot appears.
const (
	SpawnRow = 0
	SpawnCol = 5
)

// Piece is a tetromino placed on the board.
type Piece struct {
	Kind     Kind
	Rotation int // index into the kind's orientations
	Row, Col int // pivot position on the board
}

// NewPiece returns a piece of the given kind in its spawn position.
func NewPiece(k Kind) Piece {
	return Piece{Kind: k, Row: SpawnRow, Col: SpawnCol}
}

// Cells returns the board coordinates the piece covers.
func (p Piece) Cells() [4]Point {
	var out [4]Point
	for i, off := range orientations[p.Kind][p.Rotation] {
		out[i] = Point{p.Row + off.Row, p.Col + off.Col}
	}
	return out
}

// Rotated returns the piece turned by one step: clockwise when cw is true,
// counter-clockwise otherwise.
func (p Piece) Rotated(cw bool) Piece {
	n := len(orientations[p.Kind])
	if cw {
		p.Rotation = (p.Rotation + 1) % n
	} else {
		p.Rotation = (p.Rotation + n - 1) % n
	}
	return p
}

// Moved returns the piece shifted by the given number of rows and columns.
func (p Piece) Moved(dRow, dCol int) Piece {
	p.Row += dRow
	p.Col += dCol
	return p
}
