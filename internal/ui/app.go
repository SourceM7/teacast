package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
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

	spinner spinner.Model
}

func NewApp() App {
	ti := textinput.New()
	ti.Placeholder = "Search podcasts..."
	ti.CharLimit = 100

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = LoadingStyle

	return App{
		activeTab:   TabHome,
		searchInput: ti,
		loading:     true,
		statusMsg:   "Loading top podcasts...",
		player:      player.New(),
		spinner:     sp,
	}
}

func (a App) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, fetchTopPodcasts, tickCmd(), a.spinner.Tick)
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

	case spinner.TickMsg:
		var cmd tea.Cmd
		a.spinner, cmd = a.spinner.Update(msg)
		return a, cmd

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
	case "+", "=":
		a.player.SetVolume(10)
		a.statusMsg = "Volume up"
		return a, nil
	case "-", "_":
		a.player.SetVolume(-10)
		a.statusMsg = "Volume down"
		return a, nil
	case "d":
		if a.activeTab == TabBookmarks && len(a.bookmarks) > 0 && a.bookmarkCursor < len(a.bookmarks) {
			removed := a.bookmarks[a.bookmarkCursor]
			a.bookmarks = append(a.bookmarks[:a.bookmarkCursor], a.bookmarks[a.bookmarkCursor+1:]...)
			if a.bookmarkCursor >= len(a.bookmarks) && a.bookmarkCursor > 0 {
				a.bookmarkCursor--
			}
			a.statusMsg = "Removed bookmark: " + removed.Episode.Title
		}
		return a, nil
	case "u":
		if a.activeTab == TabLibrary && len(a.podcasts) > 0 && a.libraryCursor < len(a.podcasts) {
			removed := a.podcasts[a.libraryCursor]
			a.podcasts = append(a.podcasts[:a.libraryCursor], a.podcasts[a.libraryCursor+1:]...)
			if a.libraryCursor >= len(a.podcasts) && a.libraryCursor > 0 {
				a.libraryCursor--
			}
			a.statusMsg = "Unsubscribed from " + removed.Title
		}
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
	case "+", "=":
		a.player.SetVolume(10)
		a.statusMsg = "Volume up"
		return a, nil
	case "-", "_":
		a.player.SetVolume(-10)
		a.statusMsg = "Volume down"
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
