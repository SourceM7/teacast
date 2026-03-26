package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func (a App) renderBookmarks(height int) string {
	if len(a.bookmarks) == 0 {
		return RenderEmptyState("★", "No Bookmarks", "Press 'b' on an episode to save it for later", a.width)
	}

	var lines []string
	lines = append(lines, SectionHeader.Render(
		fmt.Sprintf("  Bookmarks (%d saved)", len(a.bookmarks))))
	lines = append(lines, "")

	visible := height - 3
	if visible < 3 {
		visible = 3
	}

	listStart := 0
	if a.bookmarkCursor >= listStart+visible {
		listStart = a.bookmarkCursor - visible + 1
	}

	end := listStart + visible
	if end > len(a.bookmarks) {
		end = len(a.bookmarks)
	}

	for i := listStart; i < end; i++ {
		lines = append(lines, RenderBookmarkRow(a.bookmarks[i], i == a.bookmarkCursor, a.width))
	}
	below := len(a.bookmarks) - end
	if below > 0 || listStart > 0 {
		lines = append(lines, RenderScrollHint(listStart, below))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}
