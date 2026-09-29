package tui

import "strings"

type viewport struct {
	lines []string
	top   int
	left  int
	width int
}

func newViewport(content string) viewport {
	lines := strings.Split(content, "\n")

	width := 0
	for _, line := range lines {
		if n := len([]rune(line)); n > width {
			width = n
		}
	}

	return viewport{lines: lines, width: width}
}

func (v *viewport) scroll(dy, dx, height, paneWidth int) {
	v.top += dy
	v.left += dx
	v.clamp(height, paneWidth)
}

func (v *viewport) clamp(height, paneWidth int) {
	maxTop := len(v.lines) - height
	if maxTop < 0 {
		maxTop = 0
	}
	if v.top > maxTop {
		v.top = maxTop
	}
	if v.top < 0 {
		v.top = 0
	}

	maxLeft := v.width - paneWidth
	if maxLeft < 0 {
		maxLeft = 0
	}
	if v.left > maxLeft {
		v.left = maxLeft
	}
	if v.left < 0 {
		v.left = 0
	}
}

func (v viewport) render(height, paneWidth int) string {
	if height < 1 || paneWidth < 1 {
		return ""
	}

	v.clamp(height, paneWidth)

	end := v.top + height
	if end > len(v.lines) {
		end = len(v.lines)
	}

	var b strings.Builder
	for i := v.top; i < end; i++ {
		if i > v.top {
			b.WriteString("\n")
		}
		b.WriteString(slice(v.lines[i], v.left, paneWidth))
	}

	return b.String()
}

func slice(line string, from, width int) string {
	runes := []rune(line)

	if from >= len(runes) {
		return ""
	}

	to := from + width
	if to > len(runes) {
		to = len(runes)
	}

	return string(runes[from:to])
}

func (v viewport) fits(height, paneWidth int) bool {
	return len(v.lines) <= height && v.width <= paneWidth
}
