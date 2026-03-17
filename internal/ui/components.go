package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"gocast/internal/model"
	"gocast/internal/player"
)

func FormatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh%02dm", h, m)
	}
	if m > 0 {
		return fmt.Sprintf("%dm%02ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

func TimeAgo(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	diff := time.Since(t)
	days := int(diff.Hours() / 24)
	if days == 0 {
		hours := int(diff.Hours())
		if hours == 0 {
			return "just now"
		}
		return fmt.Sprintf("%dh ago", hours)
	} else if days == 1 {
		return "Yesterday"
	} else if days < 7 {
		return fmt.Sprintf("%dd ago", days)
	} else if days < 30 {
		return fmt.Sprintf("%dw ago", days/7)
	}
	return fmt.Sprintf("%dmo ago", days/30)
}

func ProgressBar(width int, progress float64) string {
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}
	filled := int(float64(width) * progress)
	if filled > width {
		filled = width
	}
	empty := width - filled

	bar := ProgressFilled.Render(strings.Repeat("━", filled)) +
		ProgressEmpty.Render(strings.Repeat("─", empty))
	return bar
}

// RenderScrollHint returns a muted scroll hint when there are hidden items.
// Pass above=0 or below=0 to omit that direction.
func RenderScrollHint(above, below int) string {
	var parts []string
	if above > 0 {
		parts = append(parts, fmt.Sprintf("↑ %d more", above))
	}
	if below > 0 {
		parts = append(parts, fmt.Sprintf("↓ %d more", below))
	}
	if len(parts) == 0 {
		return ""
	}
	return ScrollHintStyle.Render("  " + strings.Join(parts, "   "))
}

func RenderFeaturedPodcastCard(p model.Podcast, width int) string {
	cardWidth := width - 6
	if cardWidth < 40 {
		cardWidth = 40
	}

	epCountStr := ""
	if len(p.Episodes) > 0 {
		epCountStr = fmt.Sprintf("  %d eps", len(p.Episodes))
	} else if p.EpisodeCount > 0 {
		epCountStr = fmt.Sprintf("  %d eps", p.EpisodeCount)
	}

	badgeText := " ♫ #1 TOP PODCAST" + epCountStr + " "
	badge := BadgeLive.Render(badgeText)
	title := TitleStyle.Width(cardWidth).Render(p.Title)
	author := PrimaryTextStyle.Render("by " + p.Author)

	desc := ""
	if p.Description != "" {
		d := p.Description
		if len(d) > 160 {
			d = d[:160] + "…"
		}
		desc = SubtitleStyle.Width(cardWidth).Render(d)
	}

	parts := []string{badge, "", title, author}
	if desc != "" {
		parts = append(parts, "", desc)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, parts...)
	return FeaturedCardStyle.Width(cardWidth).Render(content)
}

func RenderEpisodeRow(ep model.Episode, podcast model.Podcast, selected bool, width int) string {
	style := NormalItemStyle
	if selected {
		style = SelectedItemStyle
	}

	indicator := "  "
	if selected {
		indicator = "▶ "
	}

	titleWidth := width - 30
	if titleWidth < 20 {
		titleWidth = 20
	}

	title := ep.Title
	if len(title) > titleWidth {
		title = title[:titleWidth-1] + "…"
	}

	var duration string
	if ep.Duration > 0 {
		duration = FormatDuration(ep.Duration)
	} else {
		duration = "—"
	}
	ago := TimeAgo(ep.PublishedAt)

	row := fmt.Sprintf("%s%-*s  %8s  %s", indicator, titleWidth, title, duration, ago)
	return style.Width(width - 2).Render(row)
}

func RenderPodcastCard(p model.Podcast, selected bool, width int) string {
	style := NormalItemStyle
	if selected {
		style = SelectedItemStyle
	}

	indicator := "  "
	if selected {
		indicator = "▶ "
	}

	// Show episode count from metadata or loaded episodes
	epCount := ""
	if len(p.Episodes) > 0 {
		epCount = fmt.Sprintf("%d ep", len(p.Episodes))
	} else if p.EpisodeCount > 0 {
		epCount = fmt.Sprintf("%d ep", p.EpisodeCount)
	}

	titleWidth := width - 30
	if titleWidth < 20 {
		titleWidth = 20
	}

	title := p.Title
	if len(title) > titleWidth {
		title = title[:titleWidth-1] + "…"
	}

	author := p.Author
	if len(author) > 18 {
		author = author[:17] + "…"
	}

	row := fmt.Sprintf("%s%-*s  %-18s  %s", indicator, titleWidth, title, author, epCount)
	return style.Width(width - 2).Render(row)
}

func RenderBookmarkRow(b model.Bookmark, selected bool, width int) string {
	style := NormalItemStyle
	if selected {
		style = SelectedItemStyle
	}

	indicator := "  "
	if selected {
		indicator = "★ "
	}

	titleWidth := width - 35
	if titleWidth < 20 {
		titleWidth = 20
	}

	title := b.Episode.Title
	if len(title) > titleWidth {
		title = title[:titleWidth-1] + "…"
	}

	podName := b.Podcast.Title
	if len(podName) > 15 {
		podName = podName[:14] + "…"
	}

	row := fmt.Sprintf("%s%-*s  %-15s  %s", indicator, titleWidth, title, podName, TimeAgo(b.AddedAt))
	return style.Width(width - 2).Render(row)
}

func RenderHistoryRow(h model.HistoryEntry, selected bool, width int) string {
	style := NormalItemStyle
	if selected {
		style = SelectedItemStyle
	}

	indicator := "  "
	if selected {
		indicator = "▶ "
	}

	titleWidth := width - 40
	if titleWidth < 20 {
		titleWidth = 20
	}

	title := h.Episode.Title
	if len(title) > titleWidth {
		title = title[:titleWidth-1] + "…"
	}

	podName := h.Podcast.Title
	if len(podName) > 15 {
		podName = podName[:14] + "…"
	}

	row := fmt.Sprintf("%s%-*s  %-15s  %s", indicator, titleWidth, title, podName, TimeAgo(h.PlayedAt))
	return style.Width(width - 2).Render(row)
}

func RenderDownloadRow(d model.Download, selected bool, width int) string {
	style := NormalItemStyle
	if selected {
		style = SelectedItemStyle
	}

	indicator := "  "
	if selected {
		indicator = "▶ "
	}

	titleWidth := width - 45
	if titleWidth < 20 {
		titleWidth = 20
	}

	title := d.Episode.Title
	if len(title) > titleWidth {
		title = title[:titleWidth-1] + "…"
	}

	var statusStr string
	switch d.Status {
	case model.Downloaded:
		statusStr = AccentTextStyle.Render("✓ Complete")
	case model.Downloading:
		bar := ProgressBar(15, d.Progress)
		statusStr = fmt.Sprintf("%s %3.0f%%", bar, d.Progress*100)
	case model.NotDownloaded:
		statusStr = MutedTextStyle.Render("  Queued")
	}

	row := fmt.Sprintf("%s%-*s  %-15s  %s", indicator, titleWidth, title, d.Podcast.Title, statusStr)
	return style.Width(width - 2).Render(row)
}

func RenderPlaybar(ep *model.Episode, podcast *model.Podcast, status player.Status, width int) string {
	if ep == nil {
		content := MutedTextStyle.Render("  ♪  Nothing playing — press Enter on an episode to start")
		return PlaybarStyle.Width(width).Render(content)
	}

	var stateIcon string
	var stateLabel string
	switch status.State {
	case player.Playing:
		stateIcon = PrimaryTextStyle.Render("▶")
		stateLabel = AccentTextStyle.Render("PLAYING")
	case player.Paused:
		stateIcon = WarningTextStyle.Render("⏸")
		stateLabel = WarningTextStyle.Render("PAUSED")
	default:
		stateIcon = MutedTextStyle.Render("■")
		stateLabel = MutedTextStyle.Render("STOPPED")
	}

	// Use mpv-reported position/duration if available, fall back to episode metadata
	position := status.Position
	duration := status.Duration
	if duration == 0 {
		duration = ep.Duration
	}

	var progress float64
	if duration > 0 {
		progress = float64(position) / float64(duration)
	}

	timeStr := fmt.Sprintf("%s / %s", FormatDuration(position), FormatDuration(duration))

	// Titles — truncate to avoid wrapping
	maxTitleLen := width/3 - 4
	if maxTitleLen < 10 {
		maxTitleLen = 10
	}
	if maxTitleLen > 45 {
		maxTitleLen = 45
	}
	title := ep.Title
	if len(title) > maxTitleLen {
		title = title[:maxTitleLen-1] + "…"
	}

	maxPodLen := 22
	podTitle := podcast.Title
	if len(podTitle) > maxPodLen {
		podTitle = podTitle[:maxPodLen-1] + "…"
	}

	// Progress bar fills the gap between time info and right edge
	// Fixed elements: icon(1) + spaces(2) + stateLabel(~7) + spaces(2) + title + " · " + pod + "  " + time + spaces
	fixedLen := 1 + 2 + 7 + 2 + len(title) + 3 + len(podTitle) + 2 + len(timeStr) + 4 + 4
	barWidth := width - fixedLen
	if barWidth < 8 {
		barWidth = 8
	}
	if barWidth > 40 {
		barWidth = 40
	}

	bar := ProgressBar(barWidth, progress)

	line1 := fmt.Sprintf(" %s  %s  %s  ·  %s",
		stateIcon,
		stateLabel,
		TitleStyle.Render(title),
		SubtitleStyle.Render(podTitle),
	)
	line2 := fmt.Sprintf("         %s  %s", bar, MutedTextStyle.Render(timeStr))

	content := lipgloss.JoinVertical(lipgloss.Left, line1, line2)
	return PlaybarStyle.Width(width).Render(content)
}

func RenderHelp(keys [][]string) string {
	var parts []string
	for _, kv := range keys {
		parts = append(parts, HelpKeyStyle.Render(kv[0])+" "+HelpStyle.Render(kv[1]))
	}
	return HelpStyle.Render(strings.Join(parts, "    "))
}
