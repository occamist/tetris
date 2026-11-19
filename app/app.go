package app

import (
	"context"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/occamist/tetris/v2/game"
	"github.com/occamist/tetris/v2/ui"
)

const (
	// frame is the length of one game frame.
	frame = time.Second / game.FPS

	// maxCatchUp limits how many frames are replayed after a stall (for example
	// when the process was suspended), so the game doesn't jump ahead.
	maxCatchUp = 5 * frame
)

type Options struct {
	// NoGrid draws solid blocks without grid lines, for terminals or fonts
	// that can't show the grid glyphs.
	NoGrid bool
}

func Run(ctx context.Context, screen tcell.Screen, g *game.Game, opts Options) error {
	u := ui.New(screen, ui.Options{NoGrid: opts.NoGrid})
	if g == nil {
		seed := uint64(time.Now().UnixNano())
		g = game.New(game.NewNESRandomizer(seed))
		g.SetStartLevel(0)
	}

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
				if u.HelpOpen() {
					// Any key closes the popup; only q and ctrl-c still quit.
					if cmd == cmdQuit && ev.Key() != tcell.KeyEscape {
						return nil
					}
					u.ToggleHelp()
					break
				}
				switch cmd {
				case cmdHelp:
					u.ToggleHelp()
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
			if u.HelpOpen() {
				pending = 0
				break
			}
			for pending >= frame {
				g.Tick()
				pending -= frame
			}
			u.Draw(g)
		}
	}
}
