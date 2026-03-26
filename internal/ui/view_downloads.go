package ui

import (
	"github.com/charmbracelet/lipgloss"
)

func (a App) renderDownloads(height int) string {
	if len(a.downloads) == 0 {
		return RenderEmptyState("↓", "No Downloads", "Download feature coming soon", a.width)
	}

	var lines []string
	lines = append(lines, SectionHeader.Render("  Downloads"))
	lines = append(lines, "")

	visible := height - 3
	if visible < 3 {
		visible = 3
	}

	listStart := 0
	if a.downloadCursor >= listStart+visible {
		listStart = a.downloadCursor - visible + 1
	}

	end := listStart + visible
	if end > len(a.downloads) {
		end = len(a.downloads)
	}

	for i := listStart; i < end; i++ {
		lines = append(lines, RenderDownloadRow(a.downloads[i], i == a.downloadCursor, a.width))
	}
	below := len(a.downloads) - end
	if below > 0 || listStart > 0 {
		lines = append(lines, RenderScrollHint(listStart, below))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}
