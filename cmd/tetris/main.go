package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/occamist/tetris/v2/internal/game"
	"github.com/occamist/tetris/v2/internal/ui"
)

func main() {
	seed := flag.Uint64("seed", 0, "random seed for the piece sequence (0 picks one from the clock)")
	level := flag.Int("level", 0, fmt.Sprintf("start level shown on the title screen (0-%d)", game.MaxStartLevel))
	noGrid := flag.Bool("nogrid", false, "draw solid blocks without grid lines (for terminals that don't show underlines)")
	flag.Parse()

	if *level < 0 || *level > game.MaxStartLevel {
		fmt.Fprintf(os.Stderr, "tetris: -level must be between 0 and %d\n", game.MaxStartLevel)
		os.Exit(2)
	}
	if *seed == 0 {
		*seed = uint64(time.Now().UnixNano())
	}

	if err := run(*seed, *level, ui.Options{NoGrid: *noGrid}); err != nil {
		fmt.Fprintln(os.Stderr, "tetris:", err)
		os.Exit(1)
	}
}

func run(seed uint64, level int, opts ui.Options) error {
	screen, err := tcell.NewScreen()
	if err != nil {
		return err
	}
	if err := screen.Init(); err != nil {
		return err
	}
	// Always give the terminal back, even if the game panics; otherwise the
	// shell is left in raw mode and the panic message is unreadable.
	defer func() {
		r := recover()
		var stack []byte
		if r != nil {
			stack = debug.Stack()
		}
		screen.Fini()
		if r != nil {
			fmt.Fprintf(os.Stderr, "panic: %v\n\n%s", r, stack)
			os.Exit(2)
		}
	}()
	screen.SetStyle(tcell.StyleDefault)
	screen.HideCursor()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	g := game.New(game.NewNESRandomizer(seed))
	g.SetStartLevel(level)
	return ui.Run(ctx, screen, g, opts)
}
