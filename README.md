# Tetris

Tetris for the terminal written in Go and Tcell Go library. The rules follow
the NES-era classic: same gravity speeds, scoring, level progression, piece
randomizer and rotation, with a ghost piece and hard drop added on top.

For original one that I have written manually by hand in C, do visit [tetris v1](https://github.com/occamist/tetris/tree/ae25cc05923464328a811e2291f46035dcf7b19c)

> [!NOTE]
> Grid lines use the block characters `▁` (U+2581) and `🭿` (U+1FB7F).
> Terminals that draw block characters themselves line them up exactly with
> the cell edges: GNOME Console/Terminal (VTE), kitty, WezTerm, foot, iTerm2.
> Elsewhere they come from your font. If you see boxes or broken lines, run with `-nogrid`.

## Installation

### Grab Binaries

You can find the binaries through GitHub [releases](https://github.com/occamist/tetris/releases/).

### Install Via Go

```shell
  go install github.com/occamist/tetris/v2/cmd/tetris@latest
```

## Rules

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
▕        ████        ▏  ?  help
▕        ██          ▏
▕      ██████        ▏
▕      ██  ██        ▏
▕        ████        ▏
▕        ██          ▏
 ▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔
```

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
