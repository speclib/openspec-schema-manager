package tui

import (
	"strings"
	"testing"
)

const box = "┌────┐\n│ ab │\n└────┘"

func TestViewportMeasuresItsContent(t *testing.T) {
	t.Parallel()

	v := newViewport(box)

	if len(v.lines) != 3 {
		t.Errorf("got %d lines, want 3", len(v.lines))
	}
	if v.width != 6 {
		t.Errorf("width = %d, want 6 runes", v.width)
	}
}

func TestAViewportThatFitsNeverMoves(t *testing.T) {
	t.Parallel()

	v := newViewport(box)

	if !v.fits(10, 20) {
		t.Fatal("a small diagram does not report that it fits")
	}

	v.scroll(5, 5, 10, 20)

	if v.top != 0 || v.left != 0 {
		t.Errorf("a diagram that fits scrolled to (%d, %d)", v.top, v.left)
	}
	if got := v.render(10, 20); got != box {
		t.Errorf("render changed a diagram that fits:\n%s", got)
	}
}

func TestScrollingStopsAtEveryEdge(t *testing.T) {
	t.Parallel()

	var lines []string
	for i := range 20 {
		lines = append(lines, strings.Repeat(string(rune('a'+i%26)), 40))
	}
	v := newViewport(strings.Join(lines, "\n"))

	v.scroll(-5, -5, 5, 10)
	if v.top != 0 || v.left != 0 {
		t.Errorf("scrolling up and left past the edge gave (%d, %d)", v.top, v.left)
	}

	v.scroll(100, 100, 5, 10)
	if v.top != 15 {
		t.Errorf("top = %d, want 15 so the last line sits at the pane's edge", v.top)
	}
	if v.left != 30 {
		t.Errorf("left = %d, want 30 so the last column sits at the pane's edge", v.left)
	}
}

func TestScrollingDownMovesTheContent(t *testing.T) {
	t.Parallel()

	v := newViewport("one\ntwo\nthree\nfour\nfive")

	before := v.render(2, 10)
	v.scroll(2, 0, 2, 10)
	after := v.render(2, 10)

	if before == after {
		t.Errorf("scrolling down changed nothing:\n%s", before)
	}
	if strings.Contains(after, "one") {
		t.Errorf("the top did not scroll out of sight:\n%s", after)
	}
	if !strings.Contains(after, "three") {
		t.Errorf("the expected line is not visible:\n%s", after)
	}
}

func TestScrollingRightMovesTheContent(t *testing.T) {
	t.Parallel()

	v := newViewport("abcdefghij")

	if got := v.render(1, 4); got != "abcd" {
		t.Errorf("render = %q", got)
	}

	v.scroll(0, 4, 1, 4)

	if got := v.render(1, 4); got != "efgh" {
		t.Errorf("after scrolling right, render = %q", got)
	}
}

func TestSlicingKeepsBoxDrawingCharactersWhole(t *testing.T) {
	t.Parallel()

	v := newViewport("┌────────┐")
	v.scroll(0, 3, 1, 4)

	got := v.render(1, 4)
	for _, r := range got {
		if r == '�' {
			t.Errorf("slicing produced a replacement character: %q", got)
		}
	}
	if len([]rune(got)) != 4 {
		t.Errorf("render gave %d runes, want 4: %q", len([]rune(got)), got)
	}
}

func TestSlicePastTheEndIsEmpty(t *testing.T) {
	t.Parallel()

	if got := slice("abc", 10, 4); got != "" {
		t.Errorf("slice past the end = %q", got)
	}
	if got := slice("abc", 1, 10); got != "bc" {
		t.Errorf("slice = %q", got)
	}
	if got := slice("", 0, 4); got != "" {
		t.Errorf("slice of an empty line = %q", got)
	}
}

func TestRenderingIntoNoSpace(t *testing.T) {
	t.Parallel()

	v := newViewport(box)

	if got := v.render(0, 10); got != "" {
		t.Errorf("render with no height = %q", got)
	}
	if got := v.render(10, 0); got != "" {
		t.Errorf("render with no width = %q", got)
	}
}

func TestRenderShowsOnlyTheVisibleLines(t *testing.T) {
	t.Parallel()

	v := newViewport("one\ntwo\nthree\nfour")

	got := v.render(2, 10)
	if strings.Count(got, "\n") != 1 {
		t.Errorf("render gave %d lines for a pane of 2:\n%s", strings.Count(got, "\n")+1, got)
	}
	if !strings.Contains(got, "one") || !strings.Contains(got, "two") {
		t.Errorf("render = %q", got)
	}
	if strings.Contains(got, "three") {
		t.Errorf("render showed a line past the pane:\n%s", got)
	}
}
