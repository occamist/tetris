package game

// Phase is the screen the game is on.
type Phase int

const (
	PhaseStart     Phase = iota // title screen, choosing a start level
	PhasePlay                   // a piece is falling, or the next one is about to appear
	PhaseLineClear              // full rows are being wiped
	PhaseOver                   // the stack topped out
)

// Action is a player command. What it does depends on the phase.
type Action int

const (
	ActionLeft      Action = iota // move left; lower start level on the title screen
	ActionRight                   // move right; raise start level on the title screen
	ActionUp                      // rotate clockwise; raise start level on the title screen
	ActionDown                    // soft drop one row; lower start level on the title screen
	ActionRotateCW                // rotate clockwise
	ActionRotateCCW               // rotate counter-clockwise
	ActionHardDrop                // drop to the bottom and lock; also confirms on title and game-over screens
	ActionStart                   // start a game, or leave the game-over screen
	ActionPause                   // pause or resume
)

// Game is a single game session, from title screen to game over and back.
type Game struct {
	rng Randomizer

	phase  Phase
	paused bool
	board  Board

	piece    Piece
	hasPiece bool // false while waiting out the entry delay
	next     Kind

	startLevel int
	level      int
	lines      int
	score      int

	gravityTimer int // frames since the piece last fell
	entryTimer   int // frames left before the next piece appears
	clearTimer   int // frames into the line-clear animation
	clearing     []int
}

// New returns a game on the title screen that draws pieces from rng.
func New(rng Randomizer) *Game {
	return &Game{rng: rng, phase: PhaseStart}
}

// SetStartLevel sets the level shown on the title screen, clamped to the
// allowed range. It has no effect once a game is running.
func (g *Game) SetStartLevel(level int) {
	if g.phase == PhaseStart {
		g.startLevel = max(0, min(MaxStartLevel, level))
	}
}

// Tick advances the game by one frame.
func (g *Game) Tick() {
	if g.paused {
		return
	}
	switch g.phase {
	case PhasePlay:
		if !g.hasPiece {
			g.entryTimer--
			if g.entryTimer <= 0 {
				g.spawn()
			}
			return
		}
		g.gravityTimer++
		if g.gravityTimer >= FramesPerRow(g.level) {
			g.gravityTimer = 0
			g.fall()
		}
	case PhaseLineClear:
		g.clearTimer++
		if g.clearTimer >= LineClearFrames {
			g.finishClear()
		}
	}
}

// Handle applies one player action.
func (g *Game) Handle(a Action) {
	switch g.phase {
	case PhaseStart:
		switch a {
		case ActionUp, ActionRight:
			g.SetStartLevel(g.startLevel + 1)
		case ActionDown, ActionLeft:
			g.SetStartLevel(g.startLevel - 1)
		case ActionStart, ActionHardDrop:
			g.begin()
		}
	case PhaseOver:
		if a == ActionStart || a == ActionHardDrop {
			g.phase = PhaseStart
		}
	case PhasePlay, PhaseLineClear:
		if a == ActionPause {
			g.paused = !g.paused
			return
		}
		if g.paused || g.phase != PhasePlay || !g.hasPiece {
			return
		}
		switch a {
		case ActionLeft:
			g.try(g.piece.Moved(0, -1))
		case ActionRight:
			g.try(g.piece.Moved(0, 1))
		case ActionUp, ActionRotateCW:
			g.try(g.piece.Rotated(true))
		case ActionRotateCCW:
			g.try(g.piece.Rotated(false))
		case ActionDown:
			g.fall()
		case ActionHardDrop:
			for g.try(g.piece.Moved(1, 0)) {
			}
			g.lock()
		}
	}
}

// begin starts a new game at the chosen start level.
func (g *Game) begin() {
	g.board = Board{}
	g.level = g.startLevel
	g.lines = 0
	g.score = 0
	g.paused = false
	g.clearing = nil
	g.phase = PhasePlay
	g.next = g.rng.Next()
	g.spawn()
}

// spawn brings in the next piece. If it overlaps the stack, the game is over.
func (g *Game) spawn() {
	g.piece = NewPiece(g.next)
	g.next = g.rng.Next()
	g.hasPiece = true
	g.gravityTimer = 0
	if !g.board.Fits(g.piece.Cells()) {
		g.phase = PhaseOver
	}
}

// try moves the piece to p if it fits there.
func (g *Game) try(p Piece) bool {
	if !g.board.Fits(p.Cells()) {
		return false
	}
	g.piece = p
	return true
}

// fall moves the piece down one row, locking it if it can't move. It reports
// whether the piece moved.
func (g *Game) fall() bool {
	g.gravityTimer = 0
	if g.try(g.piece.Moved(1, 0)) {
		return true
	}
	g.lock()
	return false
}

// lock fixes the piece into the board and moves on to clearing lines or
// waiting for the next piece.
func (g *Game) lock() {
	g.hasPiece = false
	if !g.board.Place(g.piece) {
		g.phase = PhaseOver
		return
	}
	if rows := g.board.FullRows(); len(rows) > 0 {
		g.clearing = rows
		g.clearTimer = 0
		g.phase = PhaseLineClear
		return
	}
	g.entryTimer = EntryDelay(g.piece.Row)
}

// finishClear removes the cleared rows, scores them and spawns the next piece.
func (g *Game) finishClear() {
	n := len(g.clearing)
	g.board.ClearRows(g.clearing)
	g.clearing = nil
	g.score += LineScore(g.level, n)
	g.lines += n
	if g.lines >= LinesForNextLevel(g.startLevel, g.level) {
		g.level++
	}
	g.phase = PhasePlay
	g.spawn()
}

// Phase returns the current phase.
func (g *Game) Phase() Phase { return g.phase }

// Paused reports whether the game is paused.
func (g *Game) Paused() bool { return g.paused }

// Board returns a copy of the locked cells.
func (g *Game) Board() Board { return g.board }

// Piece returns the falling piece, and false if there is none right now.
func (g *Game) Piece() (Piece, bool) {
	return g.piece, g.hasPiece && g.phase == PhasePlay
}

// Ghost returns where the falling piece would land, and false if there is no
// falling piece.
func (g *Game) Ghost() (Piece, bool) {
	p, ok := g.Piece()
	if !ok {
		return Piece{}, false
	}
	for g.board.Fits(p.Moved(1, 0).Cells()) {
		p = p.Moved(1, 0)
	}
	return p, true
}

// Next returns the kind of the upcoming piece.
func (g *Game) Next() Kind { return g.next }

// StartLevel returns the level chosen on the title screen.
func (g *Game) StartLevel() int { return g.startLevel }

// Level returns the current level.
func (g *Game) Level() int { return g.level }

// Lines returns the number of lines cleared this game.
func (g *Game) Lines() int { return g.lines }

// Score returns the score this game.
func (g *Game) Score() int { return g.score }

// Clearing returns the rows being wiped and how many column pairs have been
// wiped so far, counting outwards from the centre (1 to Width/2).
func (g *Game) Clearing() (rows []int, pairs int) {
	if g.phase != PhaseLineClear {
		return nil, 0
	}
	step := LineClearFrames / (Width / 2)
	return g.clearing, min(Width/2, g.clearTimer/step+1)
}
