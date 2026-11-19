package app

import (
	"context"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/occamist/tetris/v2/game"
	"github.com/occamist/tetris/v2/ui"
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

func TestRunHelpFreezesGame(t *testing.T) {
	s := newScreen(t, ui.MinCols, ui.MinRows)
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

func TestRunQuitsOnKey(t *testing.T) {
	s := newScreen(t, ui.MinCols, ui.MinRows)
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
	if g.Phase != game.PhasePlay {
		t.Errorf("phase = %v, want PhasePlay after enter", g.Phase)
	}
}

func TestRunStopsOnContextCancel(t *testing.T) {
	s := newScreen(t, ui.MinCols, ui.MinRows)
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
