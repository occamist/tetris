package ui

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/occamist/tetris/v2/internal/game"
)

type sequence struct{ kinds []game.Kind }

func (s *sequence) NextKind() game.Kind {
	k := s.kinds[0]
	if len(s.kinds) > 1 {
		s.kinds = s.kinds[1:]
	}
	return k
}

func newScreen(t *testing.T, w, h int) tcell.SimulationScreen {
	t.Helper()
	s := tcell.NewSimulationScreen("UTF-8")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	s.SetSize(w, h)
	t.Cleanup(s.Fini)
	return s
}

// screenText returns the screen contents as one string per row.
func screenText(s tcell.SimulationScreen) []string {
	cells, w, h := s.GetContents()
	rows := make([]string, h)
	for y := range h {
		var b strings.Builder
		for x := range w {
			c := cells[y*w+x]
			if len(c.Runes) == 0 {
				b.WriteByte(' ')
			} else {
				b.WriteString(string(c.Runes))
			}
		}
		rows[y] = b.String()
	}
	return rows
}

func contains(rows []string, s string) bool {
	for _, r := range rows {
		if strings.Contains(r, s) {
			return true
		}
	}
	return false
}

// boardCell returns the two characters shown for board cell (row, col) when
// the layout is drawn at the top-left of the screen.
func boardCell(rows []string, row, col int) string {
	r := []rune(rows[1+row])
	return string(r[1+col*2 : 3+col*2])
}

func TestMapKey(t *testing.T) {
	tests := []struct {
		name   string
		ev     *tcell.EventKey
		cmd    command
		action game.Action
	}{
		{"left", tcell.NewEventKey(tcell.KeyLeft, 0, 0), cmdGame, game.ActionLeft},
		{"right", tcell.NewEventKey(tcell.KeyRight, 0, 0), cmdGame, game.ActionRight},
		{"up", tcell.NewEventKey(tcell.KeyUp, 0, 0), cmdGame, game.ActionUp},
		{"down", tcell.NewEventKey(tcell.KeyDown, 0, 0), cmdGame, game.ActionDown},
		{"enter", tcell.NewEventKey(tcell.KeyEnter, 0, 0), cmdGame, game.ActionStart},
		{"space", tcell.NewEventKey(tcell.KeyRune, ' ', 0), cmdGame, game.ActionHardDrop},
		{"x", tcell.NewEventKey(tcell.KeyRune, 'x', 0), cmdGame, game.ActionRotateCW},
		{"Z", tcell.NewEventKey(tcell.KeyRune, 'Z', 0), cmdGame, game.ActionRotateCCW},
		{"p", tcell.NewEventKey(tcell.KeyRune, 'p', 0), cmdGame, game.ActionPause},
		{"g", tcell.NewEventKey(tcell.KeyRune, 'g', 0), cmdGhost, 0},
		{"?", tcell.NewEventKey(tcell.KeyRune, '?', 0), cmdHelp, 0},
		{"q", tcell.NewEventKey(tcell.KeyRune, 'q', 0), cmdQuit, 0},
		{"esc", tcell.NewEventKey(tcell.KeyEscape, 0, 0), cmdQuit, 0},
		{"ctrl-c", tcell.NewEventKey(tcell.KeyCtrlC, 0, tcell.ModCtrl), cmdQuit, 0},
		{"unbound", tcell.NewEventKey(tcell.KeyRune, 'k', 0), cmdNone, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, action := mapKey(tt.ev)
			if cmd != tt.cmd || (cmd == cmdGame && action != tt.action) {
				t.Errorf("mapKey = (%v, %v), want (%v, %v)", cmd, action, tt.cmd, tt.action)
			}
		})
	}
}

func TestDrawTitle(t *testing.T) {
	s := newScreen(t, MinCols, MinRows)
	g := game.New(&sequence{kinds: []game.Kind{game.T}})
	g.SetStartLevel(7)
	New(s, Options{}).Draw(g)
	rows := screenText(s)
	for _, want := range []string{"T E T R I S", "START LEVEL", "07", "NEXT", "SCORE"} {
		if !contains(rows, want) {
			t.Errorf("title screen missing %q:\n%s", want, strings.Join(rows, "\n"))
		}
	}
}

// boardCellStyle returns the style of the left half of board cell (row, col).
func boardCellStyle(s tcell.SimulationScreen, row, col int) tcell.Style {
	cells, w, _ := s.GetContents()
	return cells[(1+row)*w+1+col*2].Style
}

func TestDrawGridMesh(t *testing.T) {
	s := newScreen(t, MinCols, MinRows)
	g := game.New(&sequence{kinds: []game.Kind{game.O, game.T}})
	g.Handle(game.ActionStart)
	New(s, Options{}).Draw(g)
	rows := screenText(s)

	check := func(name string, row, col int, wantFill, wantLine tcell.Color) {
		t.Helper()
		if got := boardCell(rows, row, col); got != gridGlyph {
			t.Errorf("%s cell glyph = %q, want %q", name, got, gridGlyph)
		}
		fg, bg, attrs := boardCellStyle(s, row, col).Decompose()
		if bg != wantFill || fg != wantLine {
			t.Errorf("%s cell fill/line = %v/%v, want %v/%v", name, bg, fg, wantFill, wantLine)
		}
		// Underlines sit above the cell's bottom edge and misalign the mesh.
		if attrs&tcell.AttrUnderline != 0 {
			t.Errorf("%s cell is underlined; grid lines must come from glyphs only", name)
		}
	}
	// The border hugs the board: the floor is drawn at the top of the row
	// below the board, and the walls on the sides facing the board.
	if got := []rune(rows[1+game.Height])[1]; got != '▔' {
		t.Errorf("floor glyph = %q, want '▔' so pieces sit directly on it", got)
	}
	if got := []rune(rows[1+game.Height])[0]; got != ' ' {
		t.Errorf("corner glyph = %q, want blank", got)
	}
	if got := []rune(rows[0])[1]; got != '▁' {
		t.Errorf("top edge glyph = %q, want '▁'", got)
	}
	if got := []rune(rows[5])[0]; got != '▕' {
		t.Errorf("left wall glyph = %q, want '▕'", got)
	}
	if got := []rune(rows[5])[boardW-1]; got != '▏' {
		t.Errorf("right wall glyph = %q, want '▏'", got)
	}

	check("piece", 0, 4, pieceColors[game.O], lineColors[game.O])
	check("ghost", 19, 4, ghostFill, gridColor)
	check("empty", 10, 0, tcell.ColorDefault, gridColor)
}

func TestDrawPlayShowsPieceGhostAndStats(t *testing.T) {
	s := newScreen(t, MinCols, MinRows)
	g := game.New(&sequence{kinds: []game.Kind{game.O, game.T}})
	g.Handle(game.ActionStart)
	u := New(s, Options{NoGrid: true})
	u.Draw(g)
	rows := screenText(s)

	// O spawns covering board cells (0,4), (0,5), (1,4), (1,5).
	for _, c := range [][2]int{{0, 4}, {0, 5}, {1, 4}, {1, 5}} {
		if got := boardCell(rows, c[0], c[1]); got != blockGlyph {
			t.Errorf("cell %v = %q, want block", c, got)
		}
	}
	// Ghost lands on the floor: rows 18-19.
	if got := boardCell(rows, 19, 4); got != ghostGlyph {
		t.Errorf("ghost cell = %q, want %q", got, ghostGlyph)
	}
	if !contains(rows, "000000") || !contains(rows, "LEVEL") {
		t.Error("stats panel missing")
	}

	u.ToggleGhost()
	u.Draw(g)
	if got := boardCell(screenText(s), 19, 4); got != "  " {
		t.Errorf("ghost still drawn after toggle: %q", got)
	}
}

func TestDrawPausedHidesBoard(t *testing.T) {
	s := newScreen(t, MinCols, MinRows)
	g := game.New(&sequence{kinds: []game.Kind{game.O}})
	g.Handle(game.ActionStart)
	g.Handle(game.ActionPause)
	New(s, Options{}).Draw(g)
	rows := screenText(s)
	if !contains(rows, "PAUSED") {
		t.Error("PAUSED not shown")
	}
	if got := boardCell(rows, 0, 4); got != "  " {
		t.Errorf("board visible while paused: %q", got)
	}
}

func TestHelpPopup(t *testing.T) {
	s := newScreen(t, MinCols, MinRows)
	g := game.New(&sequence{kinds: []game.Kind{game.O}})
	g.Handle(game.ActionStart)
	u := New(s, Options{})

	u.Draw(g)
	rows := screenText(s)
	if contains(rows, "hard drop") {
		t.Error("key help shown without opening the popup")
	}
	if !contains(rows, "?  help") {
		t.Error("panel hint for ? missing")
	}

	u.ToggleHelp()
	u.Draw(g)
	rows = screenText(s)
	for _, want := range []string{"HELP", "move", "rotate ↻", "rotate ↺", "soft drop", "hard drop", "pause", "ghost", "quit"} {
		if !contains(rows, want) {
			t.Errorf("help popup missing %q:\n%s", want, strings.Join(rows, "\n"))
		}
	}
	if got := boardCell(rows, 0, 4); got != "  " {
		t.Errorf("board visible behind help: %q", got)
	}
}

func TestRunHelpFreezesGame(t *testing.T) {
	s := newScreen(t, MinCols, MinRows)
	g := game.New(&sequence{kinds: []game.Kind{game.O}})
	g.Handle(game.ActionStart)
	done := make(chan error, 1)
	go func() { done <- Run(context.Background(), s, g, Options{}) }()

	s.InjectKey(tcell.KeyRune, '?', 0)
	time.Sleep(100 * time.Millisecond)
	// Escape closes the popup instead of quitting.
	s.InjectKey(tcell.KeyEscape, 0, 0)
	select {
	case <-done:
		t.Fatal("escape quit the game while help was open")
	case <-time.After(100 * time.Millisecond):
	}
	s.InjectKey(tcell.KeyRune, 'q', 0)
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after q")
	}
}

func TestDrawGameOver(t *testing.T) {
	s := newScreen(t, MinCols, MinRows)
	g := game.New(&sequence{kinds: []game.Kind{game.I}})
	g.Handle(game.ActionStart)
	// Stack vertical I pieces in the middle column until the game ends.
	for range 20 {
		if g.Phase() == game.PhaseOver {
			break
		}
		g.Handle(game.ActionUp)
		g.Handle(game.ActionHardDrop)
		for range 20 {
			g.Tick()
		}
	}
	if g.Phase() != game.PhaseOver {
		t.Fatal("could not reach game over")
	}
	New(s, Options{}).Draw(g)
	if !contains(screenText(s), "GAME OVER") {
		t.Error("GAME OVER not shown")
	}
}

func TestDrawTooSmall(t *testing.T) {
	s := newScreen(t, 30, 10)
	New(s, Options{}).Draw(game.New(&sequence{kinds: []game.Kind{game.T}}))
	rows := screenText(s)
	if !contains(rows, "Terminal too small") || !contains(rows, "have 30x10") {
		t.Errorf("too-small message missing:\n%s", strings.Join(rows, "\n"))
	}
}

func TestRunQuitsOnKey(t *testing.T) {
	s := newScreen(t, MinCols, MinRows)
	g := game.New(&sequence{kinds: []game.Kind{game.T}})
	done := make(chan error, 1)
	go func() { done <- Run(context.Background(), s, g, Options{}) }()

	time.Sleep(50 * time.Millisecond)
	s.InjectKey(tcell.KeyEnter, 0, 0)
	s.InjectKey(tcell.KeyRune, 'q', 0)

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after q")
	}
	if g.Phase() != game.PhasePlay {
		t.Errorf("phase = %v, want PhasePlay after enter", g.Phase())
	}
}

func TestRunStopsOnContextCancel(t *testing.T) {
	s := newScreen(t, MinCols, MinRows)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- Run(ctx, s, game.New(&sequence{kinds: []game.Kind{game.T}}), Options{}) }()
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after cancel")
	}
}
