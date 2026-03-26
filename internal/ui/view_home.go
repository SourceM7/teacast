package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (a App) renderHome(height int) string {
	var lines []string

	if a.loading {
		return RenderLoading(a.spinner.View()+" Loading top podcasts...", a.width)
	}

	if len(a.topPodcasts) == 0 {
		return RenderEmptyState("⌂", "No Top Podcasts", "Could not load top podcasts. Check your internet connection.", a.width)
	}

	// Only show featured card when cursor is near the top (first 3 items)
	if a.homeCursor < 3 {
		featured := a.topPodcasts[0]
		card := RenderFeaturedPodcastCard(featured, a.width)
		lines = append(lines, card)
	}

	lines = append(lines, SectionHeader.Render("  Top Podcasts"))

	listStart := 0
	visible := height - lipgloss.Height(strings.Join(lines, "\n")) - 1
	if visible < 3 {
		visible = 3
	}

	if a.homeCursor >= listStart+visible {
		listStart = a.homeCursor - visible + 1
	}
	if a.homeCursor < listStart {
		listStart = a.homeCursor
	}

	end := listStart + visible
	if end > len(a.topPodcasts) {
		end = len(a.topPodcasts)
	}

	for i := listStart; i < end; i++ {
		lines = append(lines, RenderPodcastCard(a.topPodcasts[i], i == a.homeCursor, a.width))
	}

	below := len(a.topPodcasts) - end
	if below > 0 || listStart > 0 {
		lines = append(lines, RenderScrollHint(listStart, below))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// wrapText word-wraps text to the given width, returning individual lines.
func wrapText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}
	var lines []string
	words := strings.Fields(text)
	var current string
	for _, word := range words {
		if current == "" {
			current = word
		} else if len(current)+1+len(word) <= width {
			current += " " + word
		} else {
			lines = append(lines, current)
			current = word
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}
