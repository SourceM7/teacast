package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gocast/internal/api"
	"gocast/internal/model"
	"gocast/internal/player"
)

type Tab int

const (
	TabHome Tab = iota
	TabSearch
	TabLibrary
	TabBookmarks
	TabHistory
	TabDownloads
)

var tabNames = []string{"Home", "Search", "Library", "Bookmarks", "History", "Downloads"}
var tabIcons = []string{"⌂", "⌕", "♫", "★", "↻", "↓"}

// --- Messages ---

type topPodcastsMsg struct {
	podcasts []model.Podcast
	err      error
}

type searchResultsMsg struct {
	podcasts []model.Podcast
	err      error
}

type episodesMsg struct {
	podcastID string
	episodes  []model.Episode
	err       error
}

type tickMsg time.Time

// --- Commands ---

func fetchTopPodcasts() tea.Msg {
	podcasts, err := api.FetchTopPodcasts()
	return topPodcastsMsg{podcasts: podcasts, err: err}
}

func searchPodcasts(query string) tea.Cmd {
	return func() tea.Msg {
		podcasts, err := api.SearchPodcasts(query)
		return searchResultsMsg{podcasts: podcasts, err: err}
	}
}

func fetchEpisodes(podcastID, feedURL string) tea.Cmd {
	return func() tea.Msg {
		episodes, err := api.FetchEpisodes(feedURL)
		return episodesMsg{podcastID: podcastID, episodes: episodes, err: err}
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// --- App ---

type App struct {
	activeTab Tab
	width     int
	height    int

	// Data
	topPodcasts []model.Podcast
	podcasts    []model.Podcast // library (subscribed)
	bookmarks   []model.Bookmark
	history     []model.HistoryEntry
	downloads   []model.Download

	// Cursors for each tab
	homeCursor     int
	searchCursor   int
	libraryCursor  int
	bookmarkCursor int
	historyCursor  int
	downloadCursor int

	// Search
	searchInput   textinput.Model
	searchActive  bool
	searchResults []model.Podcast
	searching     bool

	// Episode detail view
	detailPodcast  *model.Podcast
	detailEpisodes []model.Episode // episodes for detail view (separate from podcast struct)
	detailCursor   int
	loadingEps     bool

	// Playback
	player         *player.Player
	currentEpisode *model.Episode
	currentPodcast *model.Podcast
	playerStatus   player.Status

	// Status
	statusMsg string
	loading   bool
}

func NewApp() App {
	ti := textinput.New()
	ti.Placeholder = "Search podcasts..."
	ti.CharLimit = 100

	return App{
		activeTab:   TabHome,
		searchInput: ti,
		loading:     true,
		statusMsg:   "Loading top podcasts...",
		player:      player.New(),
	}
}

func (a App) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, fetchTopPodcasts, tickCmd())
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		return a, nil

	case tickMsg:
		a.playerStatus = a.player.GetStatus()
		return a, tickCmd()

	case topPodcastsMsg:
		a.loading = false
		if msg.err != nil {
			a.statusMsg = fmt.Sprintf("Error: %v", msg.err)
		} else {
			a.topPodcasts = msg.podcasts
			a.statusMsg = fmt.Sprintf("Loaded %d top podcasts", len(msg.podcasts))
		}
		return a, nil

	case searchResultsMsg:
		a.searching = false
		if msg.err != nil {
			a.statusMsg = fmt.Sprintf("Search error: %v", msg.err)
		} else {
			a.searchResults = msg.podcasts
			a.searchCursor = 0
			a.statusMsg = fmt.Sprintf("Found %d podcasts", len(msg.podcasts))
		}
		return a, nil

	case episodesMsg:
		a.loadingEps = false
		if msg.err != nil {
			a.statusMsg = fmt.Sprintf("Feed error: %v", msg.err)
		} else {
			// Update detail view episodes
			if a.detailPodcast != nil && a.detailPodcast.ID == msg.podcastID {
				a.detailEpisodes = msg.episodes
			}
			// Cache episodes in source lists
			for i := range a.topPodcasts {
				if a.topPodcasts[i].ID == msg.podcastID {
					a.topPodcasts[i].Episodes = msg.episodes
				}
			}
			for i := range a.podcasts {
				if a.podcasts[i].ID == msg.podcastID {
					a.podcasts[i].Episodes = msg.episodes
				}
			}
			for i := range a.searchResults {
				if a.searchResults[i].ID == msg.podcastID {
					a.searchResults[i].Episodes = msg.episodes
				}
			}
			a.statusMsg = fmt.Sprintf("Loaded %d episodes", len(msg.episodes))
		}
		return a, nil

	case tea.KeyMsg:
		// If in detail view, handle detail keys
		if a.detailPodcast != nil {
			return a.updateDetail(msg)
		}

		// If search input is active, let it consume most keys
		if a.activeTab == TabSearch && a.searchActive {
			return a.updateSearchInput(msg)
		}

		return a.updateMain(msg)
	}

	return a, nil
}

func (a App) updateSearchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		query := strings.TrimSpace(a.searchInput.Value())
		a.searchActive = false
		a.searchInput.Blur()
		if query == "" {
			return a, nil
		}
		a.searching = true
		a.statusMsg = "Searching iTunes..."
		return a, searchPodcasts(query)
	case "esc":
		a.searchActive = false
		a.searchInput.Blur()
		return a, nil
	case "ctrl+c":
		a.player.Cleanup()
		return a, tea.Quit
	}

	// Pass everything else to the text input
	var cmd tea.Cmd
	a.searchInput, cmd = a.searchInput.Update(msg)
	return a, cmd
}

func (a App) updateMain(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		a.player.Cleanup()
		return a, tea.Quit

	case "tab":
		a.activeTab = (a.activeTab + 1) % 6
		return a, nil
	case "shift+tab":
		a.activeTab = (a.activeTab + 5) % 6
		return a, nil

	case "1":
		a.activeTab = TabHome
		return a, nil
	case "2":
		a.activeTab = TabSearch
		return a, nil
	case "3":
		a.activeTab = TabLibrary
		return a, nil
	case "4":
		a.activeTab = TabBookmarks
		return a, nil
	case "5":
		a.activeTab = TabHistory
		return a, nil
	case "6":
		a.activeTab = TabDownloads
		return a, nil

	case " ":
		a.player.TogglePause()
		a.playerStatus = a.player.GetStatus()
		return a, nil

	case "up", "k":
		a.moveCursor(-1)
		return a, nil
	case "down", "j":
		a.moveCursor(1)
		return a, nil
	case "g":
		a.moveCursorTo(0)
		return a, nil
	case "G":
		a.moveCursorTo(-1) // -1 means last
		return a, nil

	case "enter":
		return a.handleEnter()

	case "/":
		a.activeTab = TabSearch
		a.searchActive = true
		a.searchInput.Focus()
		return a, textinput.Blink

	case "s":
		return a.subscribeCurrent()

	case "left", "h":
		a.player.Seek(-10)
		return a, nil
	case "right", "l":
		a.player.Seek(10)
		return a, nil
	}
	return a, nil
}

func (a App) updateDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "backspace":
		a.detailPodcast = nil
		a.detailEpisodes = nil
		a.detailCursor = 0
		return a, nil
	case "up", "k":
		if a.detailCursor > 0 {
			a.detailCursor--
		}
		return a, nil
	case "down", "j":
		if a.detailCursor < len(a.detailEpisodes)-1 {
			a.detailCursor++
		}
		return a, nil
	case "g":
		a.detailCursor = 0
		return a, nil
	case "G":
		if len(a.detailEpisodes) > 0 {
			a.detailCursor = len(a.detailEpisodes) - 1
		}
		return a, nil
	case "enter":
		if a.detailCursor < len(a.detailEpisodes) {
			ep := a.detailEpisodes[a.detailCursor]
			podcast := *a.detailPodcast
			a.playEpisode(ep, podcast)
		}
		return a, nil
	case "b":
		if a.detailPodcast != nil && a.detailCursor < len(a.detailEpisodes) {
			ep := a.detailEpisodes[a.detailCursor]
			a.bookmarks = append([]model.Bookmark{{
				Episode: ep,
				Podcast: *a.detailPodcast,
				AddedAt: time.Now(),
			}}, a.bookmarks...)
			a.statusMsg = "Bookmarked: " + ep.Title
		}
		return a, nil
	case "s":
		return a.subscribeDetail()
	case "ctrl+c", "q":
		a.player.Cleanup()
		return a, tea.Quit
	case " ":
		a.player.TogglePause()
		a.playerStatus = a.player.GetStatus()
		return a, nil
	case "left", "h":
		a.player.Seek(-10)
		return a, nil
	case "right", "l":
		a.player.Seek(10)
		return a, nil
	}
	return a, nil
}

func (a *App) playEpisode(ep model.Episode, podcast model.Podcast) {
	a.currentEpisode = &ep
	a.currentPodcast = &podcast

	if ep.AudioURL != "" {
		if err := a.player.Play(ep.AudioURL, ep.Title); err != nil {
			a.statusMsg = fmt.Sprintf("Playback error: %v", err)
			return
		}
		a.statusMsg = "Now playing: " + ep.Title
	} else {
		a.statusMsg = "No audio URL for this episode"
		return
	}

	// Add to history
	a.history = append([]model.HistoryEntry{{
		Episode:  ep,
		Podcast:  podcast,
		PlayedAt: time.Now(),
	}}, a.history...)
}

func (a App) handleEnter() (tea.Model, tea.Cmd) {
	switch a.activeTab {
	case TabSearch:
		if !a.searchActive && a.searchCursor < len(a.searchResults) {
			return a.openPodcastDetail(a.searchResults[a.searchCursor])
		}
		// If no results, activate search
		if !a.searchActive {
			a.searchActive = true
			a.searchInput.Focus()
			return a, textinput.Blink
		}

	case TabHome:
		if a.homeCursor < len(a.topPodcasts) {
			return a.openPodcastDetail(a.topPodcasts[a.homeCursor])
		}

	case TabLibrary:
		if a.libraryCursor < len(a.podcasts) {
			return a.openPodcastDetail(a.podcasts[a.libraryCursor])
		}

	case TabBookmarks:
		if a.bookmarkCursor < len(a.bookmarks) {
			bm := a.bookmarks[a.bookmarkCursor]
			a.playEpisode(bm.Episode, bm.Podcast)
		}

	case TabHistory:
		if a.historyCursor < len(a.history) {
			h := a.history[a.historyCursor]
			a.playEpisode(h.Episode, h.Podcast)
		}
	}

	return a, nil
}

func (a App) openPodcastDetail(p model.Podcast) (tea.Model, tea.Cmd) {
	pCopy := p
	a.detailPodcast = &pCopy
	a.detailCursor = 0

	if len(p.Episodes) > 0 {
		a.detailEpisodes = p.Episodes
		return a, nil
	}

	if p.FeedURL != "" {
		a.loadingEps = true
		a.detailEpisodes = nil
		a.statusMsg = "Loading episodes..."
		return a, fetchEpisodes(p.ID, p.FeedURL)
	}

	a.detailEpisodes = nil
	return a, nil
}

func (a App) subscribeCurrent() (tea.Model, tea.Cmd) {
	var p *model.Podcast
	switch a.activeTab {
	case TabHome:
		if a.homeCursor < len(a.topPodcasts) {
			p = &a.topPodcasts[a.homeCursor]
		}
	case TabSearch:
		if a.searchCursor < len(a.searchResults) {
			p = &a.searchResults[a.searchCursor]
		}
	}
	if p != nil {
		for _, existing := range a.podcasts {
			if existing.ID == p.ID {
				a.statusMsg = "Already subscribed to " + p.Title
				return a, nil
			}
		}
		a.podcasts = append(a.podcasts, *p)
		a.statusMsg = "Subscribed to " + p.Title
	}
	return a, nil
}

func (a App) subscribeDetail() (tea.Model, tea.Cmd) {
	if a.detailPodcast == nil {
		return a, nil
	}
	for _, existing := range a.podcasts {
		if existing.ID == a.detailPodcast.ID {
			a.statusMsg = "Already subscribed to " + a.detailPodcast.Title
			return a, nil
		}
	}
	sub := *a.detailPodcast
	sub.Episodes = a.detailEpisodes
	a.podcasts = append(a.podcasts, sub)
	a.statusMsg = "Subscribed to " + a.detailPodcast.Title
	return a, nil
}

func (a *App) moveCursor(delta int) {
	switch a.activeTab {
	case TabHome:
		a.homeCursor = clamp(a.homeCursor+delta, 0, maxIdx(len(a.topPodcasts)))
	case TabSearch:
		a.searchCursor = clamp(a.searchCursor+delta, 0, maxIdx(len(a.searchResults)))
	case TabLibrary:
		a.libraryCursor = clamp(a.libraryCursor+delta, 0, maxIdx(len(a.podcasts)))
	case TabBookmarks:
		a.bookmarkCursor = clamp(a.bookmarkCursor+delta, 0, maxIdx(len(a.bookmarks)))
	case TabHistory:
		a.historyCursor = clamp(a.historyCursor+delta, 0, maxIdx(len(a.history)))
	case TabDownloads:
		a.downloadCursor = clamp(a.downloadCursor+delta, 0, maxIdx(len(a.downloads)))
	}
}

func (a *App) moveCursorTo(pos int) {
	// pos=-1 means last item
	last := func(n int) int {
		if pos < 0 {
			return maxIdx(n)
		}
		return clamp(pos, 0, maxIdx(n))
	}
	switch a.activeTab {
	case TabHome:
		a.homeCursor = last(len(a.topPodcasts))
	case TabSearch:
		a.searchCursor = last(len(a.searchResults))
	case TabLibrary:
		a.libraryCursor = last(len(a.podcasts))
	case TabBookmarks:
		a.bookmarkCursor = last(len(a.bookmarks))
	case TabHistory:
		a.historyCursor = last(len(a.history))
	case TabDownloads:
		a.downloadCursor = last(len(a.downloads))
	}
}

// --- View ---

func (a App) View() string {
	if a.width == 0 {
		return "Loading..."
	}

	contentHeight := a.height - 7

	var sections []string
	sections = append(sections, a.renderTabBar())

	if a.detailPodcast != nil {
		sections = append(sections, a.renderDetail(contentHeight))
	} else {
		sections = append(sections, a.renderContent(contentHeight))
	}

	main := lipgloss.JoinVertical(lipgloss.Left, sections...)

	mainHeight := lipgloss.Height(main)
	playbar := a.renderPlaybar()
	help := a.renderHelpBar()
	playbarHeight := lipgloss.Height(playbar)
	helpHeight := lipgloss.Height(help)

	gap := a.height - mainHeight - playbarHeight - helpHeight
	if gap < 0 {
		gap = 0
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		main,
		strings.Repeat("\n", gap),
		playbar,
		help,
	)
}

func (a App) renderTabBar() string {
	var tabs []string
	for i, name := range tabNames {
		var label string
		if Tab(i) == a.activeTab {
			label = fmt.Sprintf(" %s %s ", tabIcons[i], name)
			tabs = append(tabs, ActiveTabStyle.Render(label))
		} else {
			label = fmt.Sprintf(" %s %s ", tabIcons[i], name)
			tabs = append(tabs, InactiveTabStyle.Render(label))
		}
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
	return TabBarStyle.Width(a.width).Render(row)
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

func (a App) renderHome(height int) string {
	var lines []string

	if a.loading {
		lines = append(lines, "")
		lines = append(lines, SectionHeader.Render("  Loading top podcasts..."))
		return lipgloss.JoinVertical(lipgloss.Left, lines...)
	}

	if len(a.topPodcasts) > 0 {
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

func (a App) renderSearch(height int) string {
	var lines []string

	lines = append(lines, SectionHeader.Render("  Search Podcasts (iTunes)"))

	inputStyle := SearchInputStyle.Width(a.width - 8)
	lines = append(lines, "  "+inputStyle.Render(a.searchInput.View()))
	lines = append(lines, "")

	if a.searching {
		lines = append(lines, MutedTextStyle.Render("  Searching iTunes..."))
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
		lines = append(lines, MutedTextStyle.Render("  No results found. Press / to search again."))
	} else if !a.searchActive {
		lines = append(lines, MutedTextStyle.Render("  Press / or Enter to start searching."))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (a App) renderLibrary(height int) string {
	var lines []string

	if len(a.podcasts) == 0 {
		lines = append(lines, SectionHeader.Render("  Your Library"))
		lines = append(lines, "")
		lines = append(lines, MutedTextStyle.Render("  No subscriptions yet. Press 's' on a podcast to subscribe."))
		return lipgloss.JoinVertical(lipgloss.Left, lines...)
	}

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

func (a App) renderBookmarks(height int) string {
	var lines []string
	lines = append(lines, SectionHeader.Render(
		fmt.Sprintf("  Bookmarks (%d saved)", len(a.bookmarks))))
	lines = append(lines, "")

	if len(a.bookmarks) == 0 {
		lines = append(lines, MutedTextStyle.Render("  No bookmarks yet. Press 'b' on an episode to bookmark it."))
		return lipgloss.JoinVertical(lipgloss.Left, lines...)
	}

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

func (a App) renderHistory(height int) string {
	var lines []string
	lines = append(lines, SectionHeader.Render(
		fmt.Sprintf("  Listening History (%d played)", len(a.history))))
	lines = append(lines, "")

	if len(a.history) == 0 {
		lines = append(lines, MutedTextStyle.Render("  No listening history yet."))
		return lipgloss.JoinVertical(lipgloss.Left, lines...)
	}

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

func (a App) renderDownloads(height int) string {
	var lines []string
	lines = append(lines, SectionHeader.Render(
		fmt.Sprintf("  Downloads (%d)", len(a.downloads))))
	lines = append(lines, "")

	if len(a.downloads) == 0 {
		lines = append(lines, MutedTextStyle.Render("  No downloads yet."))
		return lipgloss.JoinVertical(lipgloss.Left, lines...)
	}

	visible := height - 3
	if visible < 3 {
		visible = 3
	}
	end := visible
	if end > len(a.downloads) {
		end = len(a.downloads)
	}

	for i := 0; i < end; i++ {
		lines = append(lines, RenderDownloadRow(a.downloads[i], i == a.downloadCursor, a.width))
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

func (a App) renderDetail(height int) string {
	p := a.detailPodcast
	if p == nil {
		return ""
	}

	// Dual-pane layout (Golden Rule #4: weight-based sizing)
	// Left: podcast info (2/5 ≈ 40%), Right: episode list (3/5 ≈ 60%)
	if a.width < 60 {
		// Narrow terminal: fall back to single-column
		return a.renderDetailSingle(height)
	}

	leftPanelWidth := (a.width * 2) / 5
	rightPanelWidth := a.width - leftPanelWidth

	// Inner content width: panel border(2) + padding(2) = subtract 4 (Golden Rule #1)
	leftInner := leftPanelWidth - 4
	rightInner := rightPanelWidth - 4
	// Inner height: subtract 2 for top+bottom panel borders (Golden Rule #1)
	innerHeight := height - 2
	if innerHeight < 3 {
		innerHeight = 3
	}

	leftContent := a.buildDetailLeft(p, leftInner, innerHeight)
	rightContent := a.buildDetailRight(p, rightInner, innerHeight)

	leftPanel := PanelStyle.Width(leftPanelWidth).Render(leftContent)
	rightPanel := FocusedPanelStyle.Width(rightPanelWidth).Render(rightContent)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
}

// buildDetailLeft builds the left panel content (podcast info), padded to innerHeight lines.
func (a App) buildDetailLeft(p *model.Podcast, width, innerHeight int) string {
	var lines []string

	// Title
	title := p.Title
	if len(title) > width {
		title = title[:width-1] + "…"
	}
	lines = append(lines, TitleStyle.Render(title))
	lines = append(lines, "")

	// Author
	if p.Author != "" {
		author := "by " + p.Author
		if len(author) > width {
			author = author[:width-1] + "…"
		}
		lines = append(lines, PrimaryTextStyle.Render(author))
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

	// Pad to innerHeight (fill content, let border add naturally — Golden Rule anti-pattern avoided)
	for len(lines) < innerHeight {
		lines = append(lines, "")
	}
	if len(lines) > innerHeight {
		lines = lines[:innerHeight]
	}

	return strings.Join(lines, "\n")
}

// buildDetailRight builds the right panel content (episode list), padded to innerHeight lines.
func (a App) buildDetailRight(p *model.Podcast, width, innerHeight int) string {
	var lines []string

	if a.loadingEps {
		lines = append(lines, MutedTextStyle.Render("Loading episodes from feed..."))
		for len(lines) < innerHeight {
			lines = append(lines, "")
		}
		return strings.Join(lines, "\n")
	}

	if len(a.detailEpisodes) == 0 {
		msg := "No episodes found."
		if p.FeedURL == "" {
			msg = "No feed URL available."
		}
		lines = append(lines, MutedTextStyle.Render(msg))
		for len(lines) < innerHeight {
			lines = append(lines, "")
		}
		return strings.Join(lines, "\n")
	}

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

	// Scroll hint overwrites last line if there are hidden items below
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
		desc := p.Description
		if len(desc) > 200 {
			desc = desc[:200] + "…"
		}
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
		lines = append(lines, MutedTextStyle.Render("  Loading episodes..."))
		return lipgloss.JoinVertical(lipgloss.Left, lines...)
	}

	if len(a.detailEpisodes) == 0 {
		if p.FeedURL == "" {
			lines = append(lines, MutedTextStyle.Render("  No feed URL available."))
		} else {
			lines = append(lines, MutedTextStyle.Render("  No episodes found."))
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

func (a App) renderPlaybar() string {
	return RenderPlaybar(a.currentEpisode, a.currentPodcast, a.playerStatus, a.width)
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
			{"q", "quit"},
		}
	}

	help := "  " + RenderHelp(keys)
	if a.statusMsg != "" {
		help += "    " + MutedTextStyle.Render("│ "+a.statusMsg)
	}
	return help
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func maxIdx(length int) int {
	if length == 0 {
		return 0
	}
	return length - 1
}
