package ui

import "github.com/charmbracelet/lipgloss"

// ── Color palette ──────────────────────────────────────

var (
	ColorPrimary   = lipgloss.Color("205") // magenta-pink
	ColorSecondary = lipgloss.Color("39")  // bright blue
	ColorAccent    = lipgloss.Color("57")  // purple
	ColorSuccess   = lipgloss.Color("42")  // green
	ColorWarning   = lipgloss.Color("208") // orange
	ColorDanger    = lipgloss.Color("196") // red
	ColorMuted     = lipgloss.Color("240") // dark grey
	ColorText      = lipgloss.Color("255") // white
	ColorBorder    = lipgloss.Color("238") // border grey
	ColorHighlight = lipgloss.Color("201") // bright magenta
	ColorBg        = lipgloss.Color("236") // dark bg for boxes
	ColorCalToday  = lipgloss.Color("42")  // calendar today
	ColorCalSelect = lipgloss.Color("205") // calendar selected
	ColorCalHasData = lipgloss.Color("39") // calendar day with data
)

// ── Tabs ───────────────────────────────────────────────

var (
	ActiveTabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      " ",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┘",
		BottomRight: "└",
	}
	TabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┴",
		BottomRight: "┴",
	}
	ActiveTabStyle = lipgloss.NewStyle().
			Border(ActiveTabBorder).
			BorderForeground(ColorPrimary).
			Foreground(ColorPrimary).
			Bold(true).
			Padding(0, 1)
	InactiveTabStyle = lipgloss.NewStyle().
				Border(TabBorder).
				BorderForeground(ColorBorder).
				Foreground(ColorMuted).
				Padding(0, 1)
	TabGapStyle = lipgloss.NewStyle().
			Border(lipgloss.Border{Bottom: "─"}).
			BorderForeground(ColorBorder).
			Padding(0, 0)
)

// ── Layout boxes ───────────────────────────────────────

var (
	HeaderBoxStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			Padding(0, 1)
	DateBoxStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Bold(true)
	TodayBadgeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).
			Background(ColorSuccess).
			Bold(true).
			Padding(0, 1)
	ContentBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(1, 4)
	FooterStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Padding(0, 1)
	FooterKeyStyle = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true)
	FooterDescStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)
)

// ── Content styles ─────────────────────────────────────

var (
	TitleStyle   = lipgloss.NewStyle().Foreground(ColorSecondary).Bold(true).MarginBottom(1)
	LabelStyle   = lipgloss.NewStyle().Foreground(ColorMuted)
	ValueStyle   = lipgloss.NewStyle().Foreground(ColorText)
	SuccessStyle = lipgloss.NewStyle().Foreground(ColorSuccess)
	WarningStyle = lipgloss.NewStyle().Foreground(ColorWarning)
	DangerStyle  = lipgloss.NewStyle().Foreground(ColorDanger)
	InfoStyle    = lipgloss.NewStyle().Foreground(ColorMuted)
	AccentStyle  = lipgloss.NewStyle().Foreground(ColorAccent).Bold(true)
)

// ── List styles ────────────────────────────────────────

var (
	SelectedStyle = lipgloss.NewStyle().
			Foreground(ColorHighlight).
			Bold(true)
	CursorStyle = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true)
	NormalStyle = lipgloss.NewStyle().
			Foreground(ColorText)
	EmptyStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Italic(true)
)

// ── Progress / Budget ──────────────────────────────────

var (
	DeficitStyle = lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true)
	SurplusStyle = lipgloss.NewStyle().Foreground(ColorDanger).Bold(true)
	ProgressFull = lipgloss.NewStyle().Foreground(ColorSuccess)
	ProgressWarn = lipgloss.NewStyle().Foreground(ColorWarning)
	ProgressOver = lipgloss.NewStyle().Foreground(ColorDanger)
)

// ── Calendar styles ────────────────────────────────────

var (
	CalendarBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.DoubleBorder()).
				BorderForeground(ColorPrimary).
				Padding(1, 4).
				Align(lipgloss.Center)
	CalendarTitleStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true).
				Align(lipgloss.Center).
				MarginBottom(1)
	CalendarHeaderStyle = lipgloss.NewStyle().
				Foreground(ColorAccent).
				Bold(true)
	CalendarDayStyle = lipgloss.NewStyle().
				Foreground(ColorText).
				Width(4).
				Align(lipgloss.Center)
	CalendarTodayStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("0")).
				Background(ColorCalToday).
				Bold(true).
				Width(4).
				Align(lipgloss.Center)
	CalendarSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("0")).
				Background(ColorCalSelect).
				Bold(true).
				Width(4).
				Align(lipgloss.Center)
	CalendarMutedStyle = lipgloss.NewStyle().
				Foreground(ColorMuted).
				Width(4).
				Align(lipgloss.Center)
	CalendarHasDataStyle = lipgloss.NewStyle().
				Foreground(ColorCalHasData).
				Bold(true).
				Width(4).
				Align(lipgloss.Center)
	StatusMsgStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Italic(true).
			Padding(0, 1)
	InputBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorAccent).
			Padding(0, 1)
)

// ── Mood ───────────────────────────────────────────────

var MoodStyles = []lipgloss.Style{
	lipgloss.NewStyle().Foreground(ColorDanger),  // 1 - terrible
	lipgloss.NewStyle().Foreground(ColorWarning),  // 2 - bad
	lipgloss.NewStyle().Foreground(lipgloss.Color("226")), // 3 - okay (yellow)
	lipgloss.NewStyle().Foreground(lipgloss.Color("82")),  // 4 - good
	lipgloss.NewStyle().Foreground(ColorSuccess),  // 5 - great
}

func MoodEmoji(rating int) string {
	switch rating {
	case 1:
		return "😞"
	case 2:
		return "😕"
	case 3:
		return "😐"
	case 4:
		return "😊"
	case 5:
		return "😄"
	default:
		return "❓"
	}
}

func MoodLabel(rating int) string {
	switch rating {
	case 1:
		return "Terrible"
	case 2:
		return "Bad"
	case 3:
		return "Okay"
	case 4:
		return "Good"
	case 5:
		return "Great"
	default:
		return "Unknown"
	}
}
