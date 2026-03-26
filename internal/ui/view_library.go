package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func (a App) renderLibrary(height int) string {
	if len(a.podcasts) == 0 {
		return RenderEmptyState("♫", "Your Library is Empty", "Press 's' on any podcast to subscribe", a.width)
	}

	var lines []string
	lines = append(lines, SectionHeader.Render(
		fmt.Sprintf("  Your Library (%d podcasts)", len(a.podcasts))))
	lines = append(lines, "")

	visible := height - 3
	if visible < 3 {
		visible = 3
	}

	listStart := 0
	if a.libraryCursor >= listStart+visible {
		listStart = a.libraryCursor - visible + 1
	}

	end := listStart + visible
	if end > len(a.podcasts) {
		end = len(a.podcasts)
	}

	for i := listStart; i < end; i++ {
		lines = append(lines, RenderPodcastCard(a.podcasts[i], i == a.libraryCursor, a.width))
	}
	below := len(a.podcasts) - end
	if below > 0 || listStart > 0 {
		lines = append(lines, RenderScrollHint(listStart, below))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}
