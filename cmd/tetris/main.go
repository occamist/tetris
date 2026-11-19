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
	noGrid := flag.Bool("nogrid", false, "draw solid blocks without grid lines (for terminals that don't show underlines)")
	flag.Parse()

	if err := run(ui.Options{NoGrid: *noGrid}); err != nil {
		fmt.Fprintln(os.Stderr, "tetris:", err)
		os.Exit(1)
	}
}

func run(opts ui.Options) error {
	screen, err := tcell.NewScreen()
	if err != nil {
		return err
	}
	if err := screen.Init(); err != nil {
		return err
	}
	defer func() {
		// Always give the terminal back, even if the game panics
		// otherwise the shell is left in raw mode and the panic message is unreadable.
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

	seed := uint64(time.Now().UnixNano())
	g := game.New(game.NewNESRandomizer(seed))
	g.SetStartLevel(0)
	return ui.Run(ctx, screen, g, opts)
}
