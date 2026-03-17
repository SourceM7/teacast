package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	ColorPrimary    = lipgloss.Color("#7C3AED") // Purple
	ColorSecondary  = lipgloss.Color("#A78BFA") // Light purple
	ColorAccent     = lipgloss.Color("#10B981") // Green
	ColorWarning    = lipgloss.Color("#F59E0B") // Amber
	ColorDanger     = lipgloss.Color("#EF4444") // Red
	ColorMuted      = lipgloss.Color("#6B7280") // Gray
	ColorSubtle     = lipgloss.Color("#374151") // Dark gray
	ColorBg         = lipgloss.Color("#111827") // Near black
	ColorCardBg     = lipgloss.Color("#1F2937") // Dark card
	ColorText       = lipgloss.Color("#F9FAFB") // White-ish
	ColorTextDim    = lipgloss.Color("#9CA3AF") // Dim text
	ColorHighlight  = lipgloss.Color("#312E81") // Selected bg

	// Tab styles
	ActiveTabStyle = lipgloss.NewStyle().
			Foreground(ColorText).
			Background(lipgloss.Color("#1E1B4B")).
			Bold(true).
			Padding(0, 2)

	InactiveTabStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Padding(0, 2)

	TabBarStyle = lipgloss.NewStyle().
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottomForeground(ColorSubtle).
			MarginBottom(1)

	// Card styles
	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSubtle).
			Padding(1, 2)

	FeaturedCardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(1, 2)

	// Playback bar
	PlaybarStyle = lipgloss.NewStyle().
			BorderTop(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderTopForeground(ColorSubtle).
			Padding(0, 2)

	// List items
	SelectedItemStyle = lipgloss.NewStyle().
			Foreground(ColorText).
			Background(ColorHighlight).
			Bold(true).
			Padding(0, 1)

	NormalItemStyle = lipgloss.NewStyle().
			Foreground(ColorTextDim).
			Padding(0, 1)

	// Text styles
	TitleStyle = lipgloss.NewStyle().
			Foreground(ColorText).
			Bold(true)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorTextDim)

	AccentTextStyle = lipgloss.NewStyle().
			Foreground(ColorAccent)

	PrimaryTextStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary)

	MutedTextStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	WarningTextStyle = lipgloss.NewStyle().
			Foreground(ColorWarning)

	// Badge styles
	BadgeLive = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFF")).
			Background(ColorDanger).
			Padding(0, 1).
			Bold(true)

	BadgeNew = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFF")).
			Background(ColorAccent).
			Padding(0, 1).
			Bold(true)

	BadgeDownloaded = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFF")).
			Background(ColorPrimary).
			Padding(0, 1)

	// Section header
	SectionHeader = lipgloss.NewStyle().
			Foreground(ColorText).
			Bold(true).
			MarginBottom(1).
			MarginTop(1)

	// Help bar
	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Bold(true)

	// Progress bar colors
	ProgressFilled = lipgloss.NewStyle().
			Foreground(ColorPrimary)

	ProgressEmpty = lipgloss.NewStyle().
			Foreground(ColorSubtle)

	// Search
	SearchInputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(0, 1)

	// Panel for dual-pane layout (no Height — fill content to match, per golden rule)
	PanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSubtle).
			Padding(0, 1)

	// Focused panel (e.g. episode list when cursor is in it)
	FocusedPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(0, 1)

	// Scroll hint (↑ 3 more / ↓ 7 more)
	ScrollHintStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Italic(true)
)
