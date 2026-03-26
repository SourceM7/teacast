package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (a App) View() string {
	if a.width == 0 {
		return "Loading..."
	}

	tabBar := a.renderTabBar()
	playbar := a.renderPlaybar()
	statusBar := a.renderStatusBar()
	help := a.renderHelpBar()

	tabBarHeight := lipgloss.Height(tabBar)
	playbarHeight := lipgloss.Height(playbar)
	statusBarHeight := lipgloss.Height(statusBar)
	helpHeight := lipgloss.Height(help)

	contentHeight := a.height - tabBarHeight - playbarHeight - statusBarHeight - helpHeight
	if contentHeight < 1 {
		contentHeight = 1
	}

	var content string
	if a.detailPodcast != nil {
		content = a.renderDetail(contentHeight)
	} else {
		content = a.renderContent(contentHeight)
	}

	// Pad content to push playbar to the bottom
	contentRenderedHeight := lipgloss.Height(content)
	gap := contentHeight - contentRenderedHeight
	if gap < 0 {
		gap = 0
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		tabBar,
		content,
		strings.Repeat("\n", gap),
		statusBar,
		playbar,
		help,
	)
}

func (a App) renderTabBar() string {
	var tabs []string
	for i, name := range tabNames {
		var label string
		num := fmt.Sprintf("%d", i+1)
		if Tab(i) == a.activeTab {
			numPart := ActiveTabNumberStyle.Render(num)
			label = fmt.Sprintf(" %s%s %s ", numPart, tabIcons[i], name)
			tabs = append(tabs, ActiveTabStyle.Render(label))
		} else {
			numPart := TabNumberStyle.Render(num)
			label = fmt.Sprintf(" %s%s %s ", numPart, tabIcons[i], name)
			tabs = append(tabs, InactiveTabStyle.Render(label))
		}
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
	return TabBarStyle.Width(a.width).Render(row)
}

func (a App) renderStatusBar() string {
	if a.statusMsg == "" {
		return ""
	}
	return StatusBarStyle.Width(a.width).Render(a.statusMsg)
}

func (a App) renderHelpBar() string {
	var keys [][]string

	if a.detailPodcast != nil {
		keys = [][]string{
			{"esc", "back"},
			{"↑/↓", "navigate"},
			{"g/G", "first/last"},
			{"enter", "play"},
			{"s", "subscribe"},
			{"b", "bookmark"},
			{"space", "pause"},
			{"←/→", "seek"},
			{"+/-", "volume"},
			{"q", "quit"},
		}
	} else {
		keys = [][]string{
			{"tab/1-6", "tabs"},
			{"↑/↓", "navigate"},
			{"g/G", "first/last"},
			{"enter", "open"},
			{"s", "subscribe"},
			{"/", "search"},
			{"space", "pause"},
			{"←/→", "seek"},
			{"+/-", "volume"},
		}
		switch a.activeTab {
		case TabLibrary:
			keys = append(keys, []string{"u", "unsub"})
		case TabBookmarks:
			keys = append(keys, []string{"d", "remove"})
		}
		keys = append(keys, []string{"q", "quit"})
	}

	return "  " + RenderHelp(keys)
}

func (a App) renderPlaybar() string {
	return RenderPlaybar(a.currentEpisode, a.currentPodcast, a.playerStatus, a.width)
}

func (a App) renderContent(height int) string {
	switch a.activeTab {
	case TabHome:
		return a.renderHome(height)
	case TabSearch:
		return a.renderSearch(height)
	case TabLibrary:
		return a.renderLibrary(height)
	case TabBookmarks:
		return a.renderBookmarks(height)
	case TabHistory:
		return a.renderHistory(height)
	case TabDownloads:
		return a.renderDownloads(height)
	}
	return ""
}
