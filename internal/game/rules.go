package game

// FPS is the frame rate the rules are written for. Every timer in the game
// counts frames at this rate, like the NES did.
const FPS = 60

// MaxStartLevel is the highest level the title screen allows.
const MaxStartLevel = 19

// LineClearFrames is how long cleared rows are shown being wiped away.
const LineClearFrames = 20

// framesPerRow is the NES gravity table: how many frames a piece waits before
// falling one row, indexed by level. Level 29 and above use 1.
var framesPerRow = [...]int{
	48, 43, 38, 33, 28, 23, 18, 13, 8, 6, // 0-9
	5, 5, 5, // 10-12
	4, 4, 4, // 13-15
	3, 3, 3, // 16-18
	2, 2, 2, 2, 2, 2, 2, 2, 2, 2, // 19-28
	1, // 29+
}

// FramesPerRow returns how many frames pass between gravity drops at level.
func FramesPerRow(level int) int {
	if level < 0 {
		level = 0
	}
	if level >= len(framesPerRow) {
		level = len(framesPerRow) - 1
	}
	return framesPerRow[level]
}

// lineScores holds the NES base points for clearing 0-4 lines at once.
var lineScores = [...]int{0, 40, 100, 300, 1200}

// LineScore returns the points for clearing n lines at once at level.
func LineScore(level, n int) int {
	if n < 0 || n >= len(lineScores) {
		return 0
	}
	return lineScores[n] * (level + 1)
}

// LinesForNextLevel returns the total line count at which a game that began
// at startLevel moves past level. The first level-up uses the NES formula, so
// high start levels take longer to get going; after that it is every 10 lines.
func LinesForNextLevel(startLevel, level int) int {
	first := min(startLevel*10+10, max(100, startLevel*10-50))
	return first + (level-startLevel)*10
}

// EntryDelay returns how many frames pass between a piece locking and the next
// one appearing, given the row the locked piece's pivot was on. Like the NES,
// pieces locked near the bottom wait 10 frames, and the wait grows by 2 frames
// for every 4 rows higher, up to 18.
func EntryDelay(row int) int {
	fromBottom := Height - 1 - row
	if fromBottom < 0 {
		fromBottom = 0
	}
	return min(18, 10+2*((fromBottom+2)/4))
}
