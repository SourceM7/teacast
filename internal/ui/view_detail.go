package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"gocast/internal/model"
)

func (a App) renderDetail(height int) string {
	p := a.detailPodcast
	if p == nil {
		return ""
	}

	// Dual-pane layout: Left 40%, Right 60%
	if a.width < 60 {
		return a.renderDetailSingle(height)
	}

	leftPanelWidth := (a.width * 2) / 5
	rightPanelWidth := a.width - leftPanelWidth

	// Inner content width: subtract border(2) + padding(2) = 4
	leftInner := leftPanelWidth - 4
	rightInner := rightPanelWidth - 4
	// Inner height: subtract 2 for top+bottom panel borders
	innerHeight := height - 2
	if innerHeight < 3 {
		innerHeight = 3
	}

	leftContent := a.buildDetailLeft(p, leftInner, innerHeight)
	rightContent := a.buildDetailRight(p, rightInner, innerHeight)

	leftPanel := PanelStyle.Width(leftPanelWidth).Height(innerHeight).Render(leftContent)
	rightPanel := FocusedPanelStyle.Width(rightPanelWidth).Height(innerHeight).Render(rightContent)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
}

// buildDetailLeft builds the left panel content (podcast info).
func (a App) buildDetailLeft(p *model.Podcast, width, innerHeight int) string {
	var lines []string

	// Title
	lines = append(lines, TitleStyle.Render(truncate(p.Title, width)))
	lines = append(lines, "")

	// Author
	if p.Author != "" {
		lines = append(lines, PrimaryTextStyle.Render(truncate("by "+p.Author, width)))
		lines = append(lines, "")
	}

	// Description (word-wrapped)
	if p.Description != "" {
		desc := p.Description
		maxDescChars := width * 6
		if len(desc) > maxDescChars {
			desc = desc[:maxDescChars-1] + "…"
		}
		wrapped := wrapText(desc, width)
		maxDescLines := innerHeight / 2
		if maxDescLines < 2 {
			maxDescLines = 2
		}
		for i, l := range wrapped {
			if i >= maxDescLines {
				break
			}
			lines = append(lines, SubtitleStyle.Render(l))
		}
		lines = append(lines, "")
	}

	// Episode count
	if p.EpisodeCount > 0 {
		lines = append(lines, MutedTextStyle.Render(fmt.Sprintf("%d episodes", p.EpisodeCount)))
	} else if len(p.Episodes) > 0 {
		lines = append(lines, MutedTextStyle.Render(fmt.Sprintf("%d episodes", len(p.Episodes))))
	}

	// Subscribe status
	subscribed := false
	for _, existing := range a.podcasts {
		if existing.ID == p.ID {
			subscribed = true
			break
		}
	}
	lines = append(lines, "")
	if subscribed {
		lines = append(lines, AccentTextStyle.Render("✓ Subscribed"))
	} else {
		lines = append(lines, MutedTextStyle.Render("Press 's' to subscribe"))
	}

	return strings.Join(lines, "\n")
}

// buildDetailRight builds the right panel content (episode list).
func (a App) buildDetailRight(p *model.Podcast, width, innerHeight int) string {
	if a.loadingEps {
		return RenderLoading(a.spinner.View()+" Loading episodes...", width)
	}

	if len(a.detailEpisodes) == 0 {
		msg := "No episodes found."
		if p.FeedURL == "" {
			msg = "No feed URL available."
		}
		return RenderEmptyState("♪", "No Episodes", msg, width)
	}

	var lines []string

	// Header (1 line)
	header := fmt.Sprintf("Episodes (%d)", len(a.detailEpisodes))
	lines = append(lines, SectionHeader.Render(header))

	// Available lines for episode rows
	episodeLines := innerHeight - 1
	if episodeLines < 1 {
		episodeLines = 1
	}

	listStart := 0
	if a.detailCursor >= listStart+episodeLines {
		listStart = a.detailCursor - episodeLines + 1
	}
	if a.detailCursor < listStart {
		listStart = a.detailCursor
	}
	end := listStart + episodeLines
	if end > len(a.detailEpisodes) {
		end = len(a.detailEpisodes)
	}

	// Episode rows (pass width+4 to account for the -2 that RenderEpisodeRow subtracts)
	for i := listStart; i < end; i++ {
		lines = append(lines, RenderEpisodeRow(a.detailEpisodes[i], *p, i == a.detailCursor, width+2))
	}

	// Pad to innerHeight
	for len(lines) < innerHeight {
		lines = append(lines, "")
	}

	// Scroll hint overwrites last line if there are hidden items
	below := len(a.detailEpisodes) - end
	above := listStart
	if below > 0 || above > 0 {
		hint := RenderScrollHint(above, below)
		if hint != "" {
			lines[innerHeight-1] = hint
		}
	}

	return strings.Join(lines[:innerHeight], "\n")
}

// renderDetailSingle is the fallback single-column detail view for narrow terminals.
func (a App) renderDetailSingle(height int) string {
	p := a.detailPodcast
	var lines []string

	header := lipgloss.JoinVertical(lipgloss.Left,
		TitleStyle.Render("  "+p.Title),
		PrimaryTextStyle.Render("  "+p.Author),
	)
	if p.Description != "" {
		desc := truncate(p.Description, 200)
		header = lipgloss.JoinVertical(lipgloss.Left, header,
			SubtitleStyle.Width(a.width-6).Render("  "+desc),
		)
	}

	subscribed := false
	for _, existing := range a.podcasts {
		if existing.ID == p.ID {
			subscribed = true
			break
		}
	}
	if subscribed {
		header = lipgloss.JoinVertical(lipgloss.Left, header, AccentTextStyle.Render("  ✓ Subscribed"))
	} else {
		header = lipgloss.JoinVertical(lipgloss.Left, header, MutedTextStyle.Render("  Press 's' to subscribe"))
	}

	card := CardStyle.Width(a.width - 4).Render(header)
	lines = append(lines, card)

	if a.loadingEps {
		lines = append(lines, RenderLoading(a.spinner.View()+" Loading episodes...", a.width))
		return lipgloss.JoinVertical(lipgloss.Left, lines...)
	}

	if len(a.detailEpisodes) == 0 {
		if p.FeedURL == "" {
			lines = append(lines, RenderEmptyState("♪", "No Feed", "No feed URL available.", a.width))
		} else {
			lines = append(lines, RenderEmptyState("♪", "No Episodes", "No episodes found.", a.width))
		}
		return lipgloss.JoinVertical(lipgloss.Left, lines...)
	}

	lines = append(lines, SectionHeader.Render(fmt.Sprintf("  Episodes (%d)", len(a.detailEpisodes))))

	visible := height - lipgloss.Height(strings.Join(lines, "\n")) - 1
	if visible < 3 {
		visible = 3
	}

	listStart := 0
	if a.detailCursor >= listStart+visible {
		listStart = a.detailCursor - visible + 1
	}
	if a.detailCursor < listStart {
		listStart = a.detailCursor
	}

	end := listStart + visible
	if end > len(a.detailEpisodes) {
		end = len(a.detailEpisodes)
	}

	for i := listStart; i < end; i++ {
		lines = append(lines, RenderEpisodeRow(a.detailEpisodes[i], *p, i == a.detailCursor, a.width))
	}

	below := len(a.detailEpisodes) - end
	if below > 0 || listStart > 0 {
		lines = append(lines, RenderScrollHint(listStart, below))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}
