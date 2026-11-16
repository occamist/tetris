package ui

import (
	"context"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/occamist/tetris/v2/internal/game"
)

// frame is the length of one game frame.
const frame = time.Second / game.FPS

// maxCatchUp limits how many frames are replayed after a stall (for example
// when the process was suspended), so the game doesn't jump ahead.
const maxCatchUp = 5 * frame

// Run drives the game until the player quits or ctx is cancelled. The caller
// owns the screen and must have initialised it.
func Run(ctx context.Context, screen tcell.Screen, g *game.Game, opts Options) error {
	u := New(screen, opts)

	events := make(chan tcell.Event, 64)
	stop := make(chan struct{})
	go screen.ChannelEvents(events, stop)
	defer close(stop)

	ticker := time.NewTicker(frame)
	defer ticker.Stop()

	last := time.Now()
	var pending time.Duration
	u.Draw(g)

	for {
		select {
		case <-ctx.Done():
			return nil

		case ev, ok := <-events:
			if !ok {
				return nil
			}
			switch ev := ev.(type) {
			case *tcell.EventResize:
				screen.Sync()
			case *tcell.EventKey:
				cmd, action := mapKey(ev)
				switch cmd {
				case cmdQuit:
					return nil
				case cmdGhost:
					u.ToggleGhost()
				case cmdGame:
					g.Handle(action)
				}
			}
			u.Draw(g)

		case now := <-ticker.C:
			pending = min(pending+now.Sub(last), maxCatchUp)
			last = now
			for pending >= frame {
				g.Tick()
				pending -= frame
			}
			u.Draw(g)
		}
	}
}
