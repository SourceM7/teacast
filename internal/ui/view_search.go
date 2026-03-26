package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func (a App) renderSearch(height int) string {
	var lines []string

	lines = append(lines, SectionHeader.Render("  Search Podcasts (iTunes)"))

	inputStyle := SearchInputStyle.Width(a.width - 8)
	lines = append(lines, "  "+inputStyle.Render(a.searchInput.View()))
	lines = append(lines, "")

	if a.searching {
		lines = append(lines, RenderLoading(a.spinner.View()+" Searching iTunes...", a.width))
	} else if len(a.searchResults) > 0 {
		lines = append(lines, SectionHeader.Render(
			fmt.Sprintf("  Results (%d)", len(a.searchResults))))

		visible := height - len(lines) - 1
		if visible < 3 {
			visible = 3
		}

		listStart := 0
		if a.searchCursor >= listStart+visible {
			listStart = a.searchCursor - visible + 1
		}

		end := listStart + visible
		if end > len(a.searchResults) {
			end = len(a.searchResults)
		}

		for i := listStart; i < end; i++ {
			lines = append(lines, RenderPodcastCard(a.searchResults[i], i == a.searchCursor, a.width))
		}
		below := len(a.searchResults) - end
		if below > 0 || listStart > 0 {
			lines = append(lines, RenderScrollHint(listStart, below))
		}
	} else if a.searchInput.Value() != "" && !a.searchActive {
		lines = append(lines, RenderEmptyState("🔍", "No Results", "No podcasts found. Try a different search term.", a.width))
	} else if !a.searchActive {
		lines = append(lines, RenderEmptyState("⌕", "Search iTunes", "Press / or Enter to start searching for podcasts.", a.width))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}
