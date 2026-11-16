package ui

import (
	"github.com/gdamore/tcell/v2"

	"github.com/occamist/tetris/v2/internal/game"
)

// command is what a key press asks the program to do.
type command int

const (
	cmdNone  command = iota // key is not bound
	cmdGame                 // pass the action to the game
	cmdGhost                // toggle the ghost piece
	cmdQuit                 // leave the program
)

// mapKey translates a key press. Terminals don't report key releases, so
// holding a key relies on the terminal's own key repeat.
func mapKey(ev *tcell.EventKey) (command, game.Action) {
	switch ev.Key() {
	case tcell.KeyLeft:
		return cmdGame, game.ActionLeft
	case tcell.KeyRight:
		return cmdGame, game.ActionRight
	case tcell.KeyUp:
		return cmdGame, game.ActionUp
	case tcell.KeyDown:
		return cmdGame, game.ActionDown
	case tcell.KeyEnter:
		return cmdGame, game.ActionStart
	case tcell.KeyEscape, tcell.KeyCtrlC:
		return cmdQuit, 0
	case tcell.KeyRune:
		switch ev.Rune() {
		case ' ':
			return cmdGame, game.ActionHardDrop
		case 'x', 'X':
			return cmdGame, game.ActionRotateCW
		case 'z', 'Z':
			return cmdGame, game.ActionRotateCCW
		case 'p', 'P':
			return cmdGame, game.ActionPause
		case 'g', 'G':
			return cmdGhost, 0
		case 'q', 'Q':
			return cmdQuit, 0
		}
	}
	return cmdNone, 0
}
