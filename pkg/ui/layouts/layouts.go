package layouts

import (
	"bytes"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"github.com/muesli/ansi"
	"github.com/muesli/reflow/truncate"
)

type Position struct {
	x int
	y int
}

var (
	TopLeft     = &Position{}
	TopRight    = &Position{}
	BottomLeft  = &Position{}
	BottomRight = &Position{}
	Center      = &Position{}
)

func NewPosition(x, y int) *Position {
	return &Position{x: x, y: y}
}

func PlaceOverlay(pos *Position, fg, bg string) string {
	x, y := calculateXY(pos, fg, bg)

	fgLines, fgWidth := getLines(fg)
	bgLines, bgWidth := getLines(bg)
	bgHeight := len(bgLines)
	fgHeight := len(fgLines)

	if fgWidth >= bgWidth && fgHeight >= bgHeight {
		return fg
	}

	x = clamp(x, 0, bgWidth-fgWidth)
	y = clamp(y, 0, bgHeight-fgHeight)

	var b strings.Builder
	for i, bgLine := range bgLines {
		if i > 0 {
			b.WriteByte('\n')
		}
		if i < y || i >= y+fgHeight {
			b.WriteString(bgLine)
			continue
		}

		pos := 0
		if x > 0 {
			left := truncate.String(bgLine, uint(x))
			pos = ansi.PrintableRuneWidth(left)
			b.WriteString(left)
			if pos < x {
				b.WriteString(strings.Repeat(" ", x-pos))
				pos = x
			}
		}

		fgLine := fgLines[i-y]
		b.WriteString(fgLine)
		pos += ansi.PrintableRuneWidth(fgLine)

		right := cutLeft(bgLine, pos)
		bgWidth := ansi.PrintableRuneWidth(bgLine)
		rightWidth := ansi.PrintableRuneWidth(right)
		if rightWidth <= bgWidth-pos {
			b.WriteString(strings.Repeat(" ", bgWidth-rightWidth-pos))
		}

		b.WriteString(right)
	}

	return b.String()
}

func getLines(s string) (lines []string, widest int) {
	lines = strings.Split(s, "\n")

	for _, l := range lines {
		w := ansi.PrintableRuneWidth(l)
		if widest < w {
			widest = w
		}
	}

	return lines, widest
}

func calculateXY(pos *Position, fg, bg string) (x int, y int) {
	switch pos {
	case TopLeft:
		return 0, 0
	case TopRight:
		return lipgloss.Width(bg) - lipgloss.Width(fg), 0
	case BottomLeft:
		return 0, lipgloss.Height(bg) - lipgloss.Height(fg)
	case BottomRight:
		return lipgloss.Width(bg) - lipgloss.Width(fg), lipgloss.Height(bg) - lipgloss.Height(fg)
	case Center:
		return (lipgloss.Width(bg) - lipgloss.Width(fg)) / 2, (lipgloss.Height(bg) - lipgloss.Height(fg)) / 2
	default:
		return pos.x, pos.y
	}
}

func cutLeft(s string, cutWidth int) string {
	var (
		pos    int
		isAnsi bool
		ab     bytes.Buffer
		b      bytes.Buffer
	)

	for _, c := range s {
		if c == ansi.Marker || isAnsi {
			isAnsi = true
			ab.WriteRune(c)
			if ansi.IsTerminator(c) {
				isAnsi = false
				b.Write(ab.Bytes())
				ab.Reset()
			}
		} else {
			w := runewidth.RuneWidth(c)
			if pos >= cutWidth {
				b.WriteRune(c)
			}
			pos += w
		}
	}

	b.Write(ab.Bytes())

	return b.String()
}

func clamp(v, lower, upper int) int {
	return min(max(v, lower), upper)
}
