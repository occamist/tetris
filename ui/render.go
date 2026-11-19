package ui

import (
	"fmt"
	"slices"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"

	"github.com/occamist/tetris/v2/game"
)

// Layout, in terminal cells.
const (
	boardW  = game.Width*2 + 2 // two columns per cell plus the side borders
	boardH  = game.Height + 2
	panelX  = boardW + 2 // panel starts two columns right of the board
	panelW  = 18
	MinCols = panelX + panelW
	MinRows = boardH
)

// Options changes how the game is drawn.
type Options struct {
	// NoGrid draws solid blocks without grid lines, for terminals or fonts
	// that can't show the grid glyphs.
	NoGrid bool
}

// UI renders a game onto a tcell screen.
type UI struct {
	screen tcell.Screen
	look   look
	ghost  bool // draw the landing preview
	help   bool // show the key help popup instead of the board

	ox, oy int // top-left corner of the layout
}

// New returns a renderer for screen with the ghost piece turned on.
func New(screen tcell.Screen, opts Options) *UI {
	u := &UI{screen: screen, ghost: true}
	switch {
	case screen.Colors() < 8:
		u.look = lookMono
	case opts.NoGrid:
		u.look = lookPlain
	default:
		u.look = lookGrid
	}
	return u
}

// ToggleGhost turns the landing preview on or off.
func (u *UI) ToggleGhost() { u.ghost = !u.ghost }

// ToggleHelp opens or closes the key help popup.
func (u *UI) ToggleHelp() { u.help = !u.help }

// HelpOpen reports whether the help popup is showing.
func (u *UI) HelpOpen() bool { return u.help }

// Draw redraws the whole screen for the current game state.
func (u *UI) Draw(g *game.Game) {
	s := u.screen
	s.Clear()
	w, h := s.Size()
	if w < MinCols || h < MinRows {
		u.drawTooSmall(w, h)
		s.Show()
		return
	}
	u.ox, u.oy = (w-MinCols)/2, (h-MinRows)/2

	u.drawBorder()
	u.drawPanel(g)

	switch {
	case u.help:
		u.drawHelp()
	case g.Phase == game.PhaseStart:
		u.drawTitle(g)
	case g.IsPaused:
		// Hide the board while paused so pausing can't be used to plan ahead.
		u.centre(9, "PAUSED", styleTitle)
		u.centre(11, "p to resume", styleText)
	default:
		u.drawBoard(g)
		if g.Phase == game.PhaseOver {
			u.drawGameOver(g)
		}
	}
	s.Show()
}

func (u *UI) drawTooSmall(w, h int) {
	lines := []string{
		"Terminal too small",
		fmt.Sprintf("need %dx%d, have %dx%d", MinCols, MinRows, w, h),
	}
	for i, l := range lines {
		u.text(max(0, (w-runewidth.StringWidth(l))/2), h/2-1+i, l, styleText)
	}
}

func (u *UI) drawBorder() {
	u.drawBox(u.ox, u.oy, u.ox+boardW-1, u.oy+boardH-1)
}

// Box edges. Line-drawing characters like ─ and │ run through the middle of
// their character cell, which leaves half a cell of gap between the border
// and the blocks inside, so a piece on the floor looks like it's floating.
// The one-eighth block characters sit on the edge of their cell instead, so
// each side is drawn with the one that hugs the inside of the box.
var (
	flushEdges = boxEdges{top: '▁', bottom: '▔', left: '▕', right: '▏'}
	lineEdges  = boxEdges{top: '─', bottom: '─', left: '│', right: '│', tl: '┌', tr: '┐', bl: '└', br: '┘'}
)

type boxEdges struct {
	top, bottom, left, right rune
	tl, tr, bl, br           rune // corners; zero leaves them blank
}

func (u *UI) drawBox(x0, y0, x1, y1 int) {
	e := flushEdges
	if u.look == lookMono {
		e = lineEdges
	}
	s := u.screen
	for x := x0 + 1; x < x1; x++ {
		s.SetContent(x, y0, e.top, nil, styleBorder)
		s.SetContent(x, y1, e.bottom, nil, styleBorder)
	}
	for y := y0 + 1; y < y1; y++ {
		s.SetContent(x0, y, e.left, nil, styleBorder)
		s.SetContent(x1, y, e.right, nil, styleBorder)
	}
	// Flush edges meet at the inner corner of the corner cells, so those
	// stay blank.
	if e.tl != 0 {
		s.SetContent(x0, y0, e.tl, nil, styleBorder)
		s.SetContent(x1, y0, e.tr, nil, styleBorder)
		s.SetContent(x0, y1, e.bl, nil, styleBorder)
		s.SetContent(x1, y1, e.br, nil, styleBorder)
	}
}

// cell draws one board cell at board coordinates (row, col).
func (u *UI) cell(row, col int, glyph string, style tcell.Style) {
	if row < 0 || row >= game.Height || col < 0 || col >= game.Width {
		return
	}
	u.text(u.ox+1+col*2, u.oy+1+row, glyph, style)
}

func (u *UI) drawBoard(g *game.Game) {
	board := g.Board
	clearRows, pairs := g.Clearing()
	wiped := func(row, col int) bool {
		if slices.Contains(clearRows, row) {
			return col >= game.Width/2-pairs && col < game.Width/2+pairs
		}
		return false
	}

	emptyStyle, emptyGlyph, drawEmpty := u.emptyStyle()
	for r := range game.Height {
		for c := range game.Width {
			if k := board[r][c]; k != game.Empty && !wiped(r, c) {
				style, glyph := u.blockStyle(k)
				u.cell(r, c, glyph, style)
			} else if drawEmpty {
				u.cell(r, c, emptyGlyph, emptyStyle)
			}
		}
	}

	if u.ghost {
		if ghost, ok := g.Ghost(); ok {
			style, glyph := u.ghostStyle()
			for _, p := range ghost.Cells() {
				u.cell(p.Row, p.Col, glyph, style)
			}
		}
	}
	if piece, ok := g.Piece(); ok {
		style, glyph := u.blockStyle(piece.Kind)
		for _, p := range piece.Cells() {
			u.cell(p.Row, p.Col, glyph, style)
		}
	}
}

func (u *UI) drawPanel(g *game.Game) {
	x, y := u.ox+panelX, u.oy

	u.text(x, y, "NEXT", styleTitle)
	// A 4x2-cell box: wide enough for every piece in its spawn orientation.
	const boxW, boxH = 4*2 + 2, 2 + 2
	u.drawBox(x, y+1, x+boxW-1, y+boxH)
	if g.Phase != game.PhaseStart && !g.IsPaused {
		next := game.NewPiece(g.NextKind)
		style, glyph := u.blockStyle(next.Kind)
		for _, p := range next.Cells() {
			// Spawn cells sit in rows 0-1 and columns SpawnCol-2..SpawnCol+1.
			u.text(x+1+(p.Col-game.SpawnCol+2)*2, y+2+p.Row, glyph, style)
		}
	}

	u.text(x, y+6, "SCORE", styleTitle)
	u.text(x, y+7, fmt.Sprintf("%06d", g.Score), styleText)
	u.text(x, y+9, "LINES", styleTitle)
	u.text(x, y+10, fmt.Sprintf("%03d", g.Lines), styleText)
	u.text(x, y+12, "LEVEL", styleTitle)
	u.text(x, y+13, fmt.Sprintf("%02d", g.Level), styleText)

	u.text(x, y+15, "?  help", styleBorder)
}

func (u *UI) drawHelp() {
	var helpLines = []string{
		"← →  move",
		"↑ x  rotate ↻",
		"z    rotate ↺",
		"↓    soft drop",
		"spc  hard drop",
		"p    pause",
		"g    ghost",
		"q    quit",
	}

	u.centre(3, "HELP", styleTitle)
	// Centre the block as a whole so the key column stays aligned.
	width := 0
	for _, l := range helpLines {
		width = max(width, runewidth.StringWidth(l))
	}
	x := u.ox + 1 + (boardW-2-width)/2
	for i, l := range helpLines {
		u.text(x, u.oy+1+5+i, l, styleText)
	}
	u.centre(15, "any key to close", styleBorder)
}

func (u *UI) drawTitle(g *game.Game) {
	u.centre(5, "T E T R I S", styleTitle)
	u.centre(9, "START LEVEL", styleText)
	u.centre(10, fmt.Sprintf("◀  %02d  ▶", g.StartLevel), styleTitle)
	u.centre(13, "↑ ↓ to choose", styleBorder)
	u.centre(14, "enter to start", styleBorder)
}

func (u *UI) drawGameOver(g *game.Game) {
	lines := []struct {
		row   int
		text  string
		style tcell.Style
	}{
		{7, "", styleText},
		{8, "GAME OVER", styleTitle},
		{9, "", styleText},
		{10, "SCORE", styleText},
		{11, fmt.Sprintf("%06d", g.Score), styleTitle},
		{12, "", styleText},
		{13, "enter to continue", styleBorder},
		{14, "", styleText},
	}
	for _, l := range lines {
		u.blankRow(l.row)
		u.centre(l.row, l.text, l.style)
	}
}

// blankRow clears one row inside the board so overlay text stays readable.
func (u *UI) blankRow(row int) {
	for x := 1; x < boardW-1; x++ {
		u.screen.SetContent(u.ox+x, u.oy+1+row, ' ', nil, styleText)
	}
}

// centre writes text centred inside the board on the given board row.
func (u *UI) centre(row int, text string, style tcell.Style) {
	inner := boardW - 2
	x := u.ox + 1 + (inner-runewidth.StringWidth(text))/2
	u.text(x, u.oy+1+row, text, style)
}

// text writes a string starting at screen position (x, y).
func (u *UI) text(x, y int, s string, style tcell.Style) {
	for _, r := range s {
		u.screen.SetContent(x, y, r, nil, style)
		x += runewidth.RuneWidth(r)
	}
}
