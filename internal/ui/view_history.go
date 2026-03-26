package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func (a App) renderHistory(height int) string {
	if len(a.history) == 0 {
		return RenderEmptyState("↻", "No History", "Episodes you play will show up here", a.width)
	}

	var lines []string
	lines = append(lines, SectionHeader.Render(
		fmt.Sprintf("  Listening History (%d played)", len(a.history))))
	lines = append(lines, "")

	visible := height - 3
	if visible < 3 {
		visible = 3
	}

	listStart := 0
	if a.historyCursor >= listStart+visible {
		listStart = a.historyCursor - visible + 1
	}

	end := listStart + visible
	if end > len(a.history) {
		end = len(a.history)
	}

	for i := listStart; i < end; i++ {
		lines = append(lines, RenderHistoryRow(a.history[i], i == a.historyCursor, a.width))
	}
	below := len(a.history) - end
	if below > 0 || listStart > 0 {
		lines = append(lines, RenderScrollHint(listStart, below))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}
