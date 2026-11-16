# Tetris

A falling-block puzzle game for the terminal, written in Go. The rules follow
the NES-era classic: same gravity speeds, scoring, level progression, piece
randomizer and rotation, with a ghost piece and hard drop added on top.

The board and pieces are drawn as a mesh of thin grid lines, in colour. The
snapshot below is plain text, so it shows the `-nogrid` look:

```
 ▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁   NEXT
▕                    ▏   ▁▁▁▁▁▁▁▁
▕                    ▏  ▕  ██████▏
▕        ██████      ▏  ▕  ██    ▏
▕            ██      ▏   ▔▔▔▔▔▔▔▔
▕                    ▏
▕                    ▏  SCORE
▕        ░░░░░░      ▏  000000
▕    ████████░░      ▏
▕        ██          ▏  LINES
▕        ████        ▏  000
▕          ██        ▏
▕      ██████        ▏  LEVEL
▕          ██        ▏  00
▕          ██        ▏
▕        ████        ▏  ← →  move
▕        ██          ▏  ↑ x  rotate ↻
▕      ██████        ▏  z    rotate ↺
▕      ██  ██        ▏  ↓    soft drop
▕        ████        ▏  spc  hard drop
▕        ██          ▏  p g  pause/ghost
 ▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔   q    quit
```

## Requirements

- Go
- Grid lines use the block characters `▁` (U+2581) and `🭿` (U+1FB7F).
  Terminals that draw block characters themselves line them up exactly with
  the cell edges: GNOME Console/Terminal (VTE), kitty, WezTerm, foot, iTerm2.
  Elsewhere they come from your font. If you see boxes or broken lines, run
  with `-nogrid`.

## Run

```sh
go run ./cmd/tetris
```

or install it onto your `PATH`:

```sh
go install github.com/occamist/tetris/v2/cmd/tetris@latest
tetris
```

### Flags

| Flag | Default | Meaning |
|---|---|---|
| `-level N` | `0` | Start level preselected on the title screen (0–19) |
| `-seed N` | clock | Seed for the piece sequence. Use the same seed to replay the same pieces |
| `-nogrid` | off | Draw solid blocks without grid lines |

## Controls

| Key | Title screen | In game |
|---|---|---|
| `←` `→` | Change start level | Move |
| `↑` | Raise start level | Rotate clockwise |
| `↓` | Lower start level | Soft drop one row |
| `x` / `z` | | Rotate clockwise / counter-clockwise |
| `space` | Start | Hard drop |
| `enter` | Start | |
| `p` | | Pause (the board is hidden while paused) |
| `g` | | Toggle the ghost piece |
| `q` / `esc` / `ctrl-c` | Quit | Quit |

Terminals only report key presses, not releases, so holding a key repeats at
your terminal's key-repeat rate.

## Rules

- **Board:** 10 × 20. A new piece appears at the top centre. If it overlaps
  the stack, or a piece locks partly above the top, the game is over.
- **Gravity:** the NES frames-per-row table at 60 frames per second: 48 frames
  at level 0, down to 6 at level 9, 2 from level 19 and 1 from level 29.
- **Rotation:** the NES system. No wall kicks; I, S and Z have two states.
- **Randomizer:** the NES rule. Roll one of eight; if that is the unused
  eighth value or repeats the previous piece, roll once more among the seven.
- **Entry delay:** 10–18 frames after a piece locks, longer the higher it locked.
- **Line clear:** full rows are wiped from the centre outwards over 20 frames.
- **Scoring:** 40 / 100 / 300 / 1200 × (level + 1) for 1 / 2 / 3 / 4 lines.
  Only cleared lines score; dropping pieces gives no points.
- **Levels:** the first level-up comes at `min(start×10+10, max(100, start×10−50))`
  lines, then every 10 lines.
