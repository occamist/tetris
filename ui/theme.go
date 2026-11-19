package ui

import (
	"github.com/gdamore/tcell/v2"

	"github.com/occamist/tetris/v2/game"
)

// pieceColors fill the blocks and lineColors draw the grid lines on them.
// Both come from the original C version's light and dark palettes. tcell
// maps them to the nearest colour on terminals without truecolor.
var (
	pieceColors = [game.NumKinds + 1]tcell.Color{
		game.I: tcell.NewHexColor(0x44E5E5),
		game.O: tcell.NewHexColor(0xE5E544),
		game.T: tcell.NewHexColor(0xE544E5),
		game.S: tcell.NewHexColor(0x44E57A),
		game.Z: tcell.NewHexColor(0xE54444),
		game.J: tcell.NewHexColor(0x4495E5),
		game.L: tcell.NewHexColor(0xE59544),
	}
	lineColors = [game.NumKinds + 1]tcell.Color{
		game.I: tcell.NewHexColor(0x1E6666),
		game.O: tcell.NewHexColor(0x66661E),
		game.T: tcell.NewHexColor(0x661E66),
		game.S: tcell.NewHexColor(0x1E6636),
		game.Z: tcell.NewHexColor(0x661E1E),
		game.J: tcell.NewHexColor(0x1E4266),
		game.L: tcell.NewHexColor(0x66421E),
	}
	gridColor = tcell.NewHexColor(0x3A3A3A) // lines on empty cells
	ghostFill = tcell.NewHexColor(0x303030) // background of the landing preview
)

// Each board cell is two terminal columns wide so blocks look square.
//
// In grid mode a cell is "lower one-eighth block" followed by "right and
// lower one-eighth block" (U+1FB7F): together a line along the cell's bottom
// edge and one down its right edge, so neighbouring cells form a mesh. Modern
// terminals draw these block characters themselves, filling the cell to the
// exact pixel, so the lines meet cleanly. (Underlines don't work for this:
// terminals draw them at the font's underline height, a few pixels above the
// cell's bottom edge, so they never meet the vertical lines.)
const (
	gridGlyph      = "▁\U0001FB7F"
	blockGlyph     = "██"
	ghostGlyph     = "░░"
	monoBlockGlyph = "[]"
	monoGhostGlyph = "::"
)

// look is how the board cells are drawn.
type look int

const (
	lookGrid  look = iota // coloured cells with thin grid lines
	lookPlain             // coloured solid blocks, empty cells left blank
	lookMono              // ASCII glyphs for terminals with few colours
)

var (
	styleText   = tcell.StyleDefault
	styleBorder = tcell.StyleDefault.Foreground(tcell.ColorGray)
	styleTitle  = tcell.StyleDefault.Bold(true)
	styleGhost  = tcell.StyleDefault.Foreground(tcell.ColorGray)
)

// lined returns a grid-mode style: fill as the background, line as the
// colour of the edge glyphs.
func lined(fill, line tcell.Color) tcell.Style {
	return tcell.StyleDefault.Background(fill).Foreground(line)
}

// blockStyle returns the style and glyph for a locked or falling cell.
func (u *UI) blockStyle(k game.Kind) (tcell.Style, string) {
	switch u.look {
	case lookGrid:
		return lined(pieceColors[k], lineColors[k]), gridGlyph
	case lookPlain:
		return tcell.StyleDefault.Foreground(pieceColors[k]), blockGlyph
	default:
		return styleText, monoBlockGlyph
	}
}

// ghostStyle returns the style and glyph for the landing preview.
func (u *UI) ghostStyle() (tcell.Style, string) {
	switch u.look {
	case lookGrid:
		return lined(ghostFill, gridColor), gridGlyph
	case lookPlain:
		return styleGhost, ghostGlyph
	default:
		return styleText, monoGhostGlyph
	}
}

// emptyStyle returns the style and glyph for an empty board cell, and false
// if empty cells are left blank in the current look.
func (u *UI) emptyStyle() (tcell.Style, string, bool) {
	if u.look != lookGrid {
		return styleText, "", false
	}
	return lined(tcell.ColorDefault, gridColor), gridGlyph, true
}
