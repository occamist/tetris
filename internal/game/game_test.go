package game

import (
	"slices"
	"testing"
)

// sequence hands out the given kinds in order, then repeats the last one.
type sequence struct{ kinds []Kind }

func (s *sequence) Next() Kind {
	k := s.kinds[0]
	if len(s.kinds) > 1 {
		s.kinds = s.kinds[1:]
	}
	return k
}

func newGame(kinds ...Kind) *Game {
	return New(&sequence{kinds: kinds})
}

// started returns a game already in play at the given level.
func started(level int, kinds ...Kind) *Game {
	g := newGame(kinds...)
	g.SetStartLevel(level)
	g.Handle(ActionStart)
	return g
}

func ticks(g *Game, n int) {
	for range n {
		g.Tick()
	}
}

func mustPiece(t *testing.T, g *Game) Piece {
	t.Helper()
	p, ok := g.Piece()
	if !ok {
		t.Fatalf("no falling piece (phase %v)", g.Phase())
	}
	return p
}

func TestTitleScreenLevelSelect(t *testing.T) {
	g := newGame(T)
	if g.Phase() != PhaseStart {
		t.Fatalf("new game phase = %v, want PhaseStart", g.Phase())
	}
	g.Handle(ActionDown)
	if g.StartLevel() != 0 {
		t.Errorf("level went below 0: %d", g.StartLevel())
	}
	g.Handle(ActionUp)
	g.Handle(ActionRight)
	if g.StartLevel() != 2 {
		t.Errorf("start level = %d, want 2", g.StartLevel())
	}
	g.SetStartLevel(50)
	if g.StartLevel() != MaxStartLevel {
		t.Errorf("start level = %d, want clamp to %d", g.StartLevel(), MaxStartLevel)
	}
	g.Handle(ActionStart)
	if g.Phase() != PhasePlay || g.Level() != MaxStartLevel {
		t.Errorf("after start: phase %v level %d", g.Phase(), g.Level())
	}
	g.SetStartLevel(0)
	if g.StartLevel() != MaxStartLevel {
		t.Error("SetStartLevel changed the level during play")
	}
}

func TestStartSpawnsFirstPieceAndPreview(t *testing.T) {
	g := started(0, T, L, O)
	if p := mustPiece(t, g); p != NewPiece(T) {
		t.Errorf("first piece = %+v, want T at spawn", p)
	}
	if g.Next() != L {
		t.Errorf("next = %v, want L", g.Next())
	}
}

func TestGravityFollowsLevelSpeed(t *testing.T) {
	g := started(0, T)
	ticks(g, FramesPerRow(0)-1)
	if mustPiece(t, g).Row != SpawnRow {
		t.Fatal("piece fell too early")
	}
	g.Tick()
	if mustPiece(t, g).Row != SpawnRow+1 {
		t.Fatal("piece did not fall after 48 frames at level 0")
	}

	fast := started(19, T)
	ticks(fast, 2)
	if mustPiece(t, fast).Row != SpawnRow+1 {
		t.Fatal("piece did not fall after 2 frames at level 19")
	}
}

func TestMoveStopsAtWalls(t *testing.T) {
	g := started(0, O)
	for range 10 {
		g.Handle(ActionLeft)
	}
	for _, c := range mustPiece(t, g).Cells() {
		if c.Col < 0 {
			t.Fatalf("piece went through left wall: %v", c)
		}
	}
	if got := mustPiece(t, g).Col; got != 1 {
		t.Errorf("O pivot column at left wall = %d, want 1", got)
	}
	for range 10 {
		g.Handle(ActionRight)
	}
	if got := mustPiece(t, g).Col; got != Width-1 {
		t.Errorf("O pivot column at right wall = %d, want %d", got, Width-1)
	}
}

func TestRotationBlockedWithoutWallKick(t *testing.T) {
	g := started(0, I)
	g.Handle(ActionRotateCW) // vertical
	for range 10 {
		g.Handle(ActionLeft)
	}
	before := mustPiece(t, g)
	if before.Col != 0 {
		t.Fatalf("vertical I did not reach the wall, col %d", before.Col)
	}
	g.Handle(ActionRotateCCW) // horizontal would need cols -2..1
	if after := mustPiece(t, g); after != before {
		t.Errorf("rotation into the wall was allowed: %+v", after)
	}
}

func TestRotateBothDirections(t *testing.T) {
	g := started(0, T)
	g.Handle(ActionUp)
	if r := mustPiece(t, g).Rotation; r != 1 {
		t.Errorf("after up rotation = %d, want 1", r)
	}
	g.Handle(ActionRotateCCW)
	g.Handle(ActionRotateCCW)
	if r := mustPiece(t, g).Rotation; r != 3 {
		t.Errorf("after two CCW rotation = %d, want 3", r)
	}
}

func TestSoftDropResetsGravityWithoutScoring(t *testing.T) {
	g := started(0, T)
	ticks(g, 40)
	g.Handle(ActionDown)
	if mustPiece(t, g).Row != SpawnRow+1 || g.Score() != 0 {
		t.Fatalf("soft drop: row %d score %d", mustPiece(t, g).Row, g.Score())
	}
	ticks(g, FramesPerRow(0)-1)
	if mustPiece(t, g).Row != SpawnRow+1 {
		t.Error("gravity timer was not reset by soft drop")
	}
}

func TestHardDropLocksWithoutScoringAndWaitsEntryDelay(t *testing.T) {
	g := started(0, O, T)
	g.Handle(ActionHardDrop)
	// Only cleared lines score; dropping alone gives nothing.
	if g.Score() != 0 {
		t.Errorf("hard drop score = %d, want 0", g.Score())
	}
	b := g.Board()
	if b[19][4] != O || b[18][5] != O {
		t.Error("O was not locked at the bottom")
	}
	if _, ok := g.Piece(); ok {
		t.Fatal("next piece appeared with no entry delay")
	}
	ticks(g, EntryDelay(18)-1)
	if _, ok := g.Piece(); ok {
		t.Fatal("next piece appeared before the entry delay ended")
	}
	g.Tick()
	if p := mustPiece(t, g); p.Kind != T {
		t.Errorf("next piece = %v, want T", p.Kind)
	}
}

func TestLineClear(t *testing.T) {
	g := started(0, I, O)
	fillRow(&g.board, 19, J, 3, 4, 5, 6) // gap exactly where a flat I lands
	g.Handle(ActionHardDrop)

	if g.Phase() != PhaseLineClear {
		t.Fatalf("phase = %v, want PhaseLineClear", g.Phase())
	}
	rows, pairs := g.Clearing()
	if !slices.Equal(rows, []int{19}) || pairs != 1 {
		t.Errorf("Clearing() = %v, %d; want [19], 1", rows, pairs)
	}

	// Input is ignored during the animation.
	g.Handle(ActionLeft)

	ticks(g, LineClearFrames-1)
	if _, pairs = g.Clearing(); pairs != Width/2 {
		t.Errorf("pairs near end of animation = %d, want %d", pairs, Width/2)
	}
	g.Tick()

	if g.Phase() != PhasePlay {
		t.Fatalf("phase after animation = %v, want PhasePlay", g.Phase())
	}
	if g.Lines() != 1 {
		t.Errorf("lines = %d, want 1", g.Lines())
	}
	if want := LineScore(0, 1); g.Score() != want {
		t.Errorf("score = %d, want %d", g.Score(), want)
	}
	if g.Board() != (Board{}) {
		t.Error("board not empty after clearing the only row")
	}
	if p := mustPiece(t, g); p.Kind != O {
		t.Errorf("piece after clear = %v, want O", p.Kind)
	}
}

func TestFourLineClearAndLevelUp(t *testing.T) {
	g := started(0, I, O)
	g.lines = 9
	for r := 16; r < Height; r++ {
		fillRow(&g.board, r, L, 0)
	}
	g.Handle(ActionRotateCW) // vertical I, cells rows -2..1 at pivot col 5
	for range 5 {
		g.Handle(ActionLeft)
	}
	if c := mustPiece(t, g).Col; c != 0 {
		t.Fatalf("vertical I col = %d, want 0", c)
	}
	g.Handle(ActionHardDrop)
	ticks(g, LineClearFrames)

	if g.Lines() != 13 {
		t.Errorf("lines = %d, want 13", g.Lines())
	}
	if g.Level() != 1 {
		t.Errorf("level = %d, want 1 after passing 10 lines", g.Level())
	}
	if want := LineScore(0, 4); g.Score() != want {
		t.Errorf("score = %d, want %d (scored at the old level)", g.Score(), want)
	}
}

func TestGameOverWhenSpawnBlocked(t *testing.T) {
	g := started(0, O, T)
	g.board[0][6] = Z // clear of the falling O, but where T spawns
	g.Handle(ActionHardDrop)
	ticks(g, EntryDelay(18))
	if g.Phase() != PhaseOver {
		t.Fatalf("phase = %v, want PhaseOver", g.Phase())
	}
	if _, ok := g.Piece(); ok {
		t.Error("Piece() reports a falling piece after game over")
	}
	g.Handle(ActionStart)
	if g.Phase() != PhaseStart {
		t.Errorf("phase after confirm = %v, want PhaseStart", g.Phase())
	}
}

func TestGameOverWhenLockingAboveTop(t *testing.T) {
	g := started(0, I)
	g.board[2][5] = Z
	g.Handle(ActionRotateCW) // vertical I: rows -2..1, resting on the Z
	g.Handle(ActionHardDrop)
	if g.Phase() != PhaseOver {
		t.Fatalf("phase = %v, want PhaseOver", g.Phase())
	}
}

func TestPauseFreezesGame(t *testing.T) {
	g := started(0, T)
	g.Handle(ActionPause)
	if !g.Paused() {
		t.Fatal("not paused")
	}
	ticks(g, 500)
	g.Handle(ActionLeft)
	g.Handle(ActionHardDrop)
	if p := mustPiece(t, g); p != NewPiece(T) {
		t.Errorf("piece changed while paused: %+v", p)
	}
	g.Handle(ActionPause)
	ticks(g, FramesPerRow(0))
	if mustPiece(t, g).Row != SpawnRow+1 {
		t.Error("gravity did not resume after unpausing")
	}
}

func TestRestartResetsState(t *testing.T) {
	g := started(3, O)
	g.score, g.lines, g.level = 999, 40, 7
	g.board[19][0] = J
	g.phase = PhaseOver
	g.Handle(ActionStart) // to title
	g.Handle(ActionStart) // new game
	if g.Score() != 0 || g.Lines() != 0 || g.Level() != 3 || g.Board() != (Board{}) {
		t.Errorf("restart kept old state: score %d lines %d level %d", g.Score(), g.Lines(), g.Level())
	}
}

func TestGhost(t *testing.T) {
	g := started(0, O)
	g.board[10][4] = S
	ghost, ok := g.Ghost()
	if !ok {
		t.Fatal("no ghost")
	}
	if ghost.Row != 8 { // O cells rows 8-9, resting on row 10
		t.Errorf("ghost row = %d, want 8", ghost.Row)
	}
	if mustPiece(t, g).Row != SpawnRow {
		t.Error("Ghost moved the real piece")
	}
}
