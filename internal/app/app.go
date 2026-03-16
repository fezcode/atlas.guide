package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"atlas.guide/internal/data"
	"atlas.guide/internal/ui"
)

type tab int

const (
	tabOverview tab = iota
	tabMonthly
	tabCalories
	tabGym
	tabMood
	tabMedicine
	tabJournal
)

type inputMode int

const (
	inputNone inputMode = iota
	inputFoodName
	inputFoodCalories
	inputFoodProtein
	inputFoodFat
	inputFoodCarbs
	inputCalorieGoal
	inputFastStart
	inputFastEnd
	inputGymName
	inputGymDuration
	inputGymCalories
	inputMoodRating
	inputMoodNote
	inputMedicineName
	inputMedicineDosage
	inputJournalNote
)

type Model struct {
	activeTab tab
	date      time.Time
	width     int
	height    int

	// Data
	calories data.CalorieDay
	gym      data.GymDay
	mood     data.MoodEntry
	medicine data.MedicineDay
	journal    data.JournalDay

	// UI state
	cursor      int
	input       textinput.Model
	inputMode   inputMode
	isHelp      bool
	isCalendar  bool
	statusMsg   string
	progressBar progress.Model

	// Calendar state
	calYear  int
	calMonth time.Month
	calDay   int

	// Temp input buffers
	tmpFoodName     string
	tmpFoodCalories int
	tmpFoodProtein  int
	tmpFoodFat      int
	tmpGymName      string
	tmpGymDuration  int
	tmpMedicineName string
}

func NewModel() Model {
	ti := textinput.New()
	ti.CharLimit = 100
	ti.Width = 40

	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)

	now := time.Now()
	m := Model{
		activeTab:   tabOverview,
		date:        now,
		input:       ti,
		progressBar: p,
		width:       80,
		height:      24,
		calYear:     now.Year(),
		calMonth:    now.Month(),
		calDay:      now.Day(),
	}
	m.loadData()
	return m
}

func (m *Model) loadData() {
	m.calories = data.LoadCalories(m.date)
	m.gym = data.LoadGym(m.date)
	m.mood = data.LoadMood(m.date)
	m.medicine = data.LoadMedicine(m.date)
	m.journal = data.LoadJournal(m.date)
	m.cursor = 0
}

func (m Model) Init() tea.Cmd {
	return tea.WindowSize()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if m.isHelp {
			m.isHelp = false
			return m, nil
		}

		if m.isCalendar {
			return m.handleCalendar(msg)
		}

		if m.inputMode != inputNone {
			return m.handleInput(msg)
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "?":
			m.isHelp = true
		case "1":
			m.activeTab = tabOverview
			m.cursor = 0
		case "2":
			m.activeTab = tabMonthly
			m.cursor = 0
		case "3":
			m.activeTab = tabCalories
			m.cursor = 0
		case "4":
			m.activeTab = tabGym
			m.cursor = 0
		case "5":
			m.activeTab = tabMood
			m.cursor = 0
		case "6":
			m.activeTab = tabMedicine
			m.cursor = 0
		case "7":
			m.activeTab = tabJournal
			m.cursor = 0
		case "tab":
			m.activeTab = (m.activeTab + 1) % 7
			m.cursor = 0
		case "h", "left":
			m.date = m.date.AddDate(0, 0, -1)
			m.loadData()
			m.statusMsg = ""
		case "l", "right":
			m.date = m.date.AddDate(0, 0, 1)
			m.loadData()
			m.statusMsg = ""
		case "t":
			m.date = time.Now()
			m.loadData()
			m.statusMsg = "Jumped to today"
		case "c":
			m.isCalendar = true
			m.calYear = m.date.Year()
			m.calMonth = m.date.Month()
			m.calDay = m.date.Day()
		case "j", "down":
			m.cursorDown()
		case "k", "up":
			m.cursorUp()
		case "a":
			m.startAdd()
		case "d":
			m.deleteEntry()
		case "f":
			if m.activeTab == tabCalories {
				m.toggleFasting()
			}
		case "H":
			if m.activeTab == tabMonthly {
				m.date = m.date.AddDate(0, -1, 0)
				m.loadData()
				m.statusMsg = ""
			}
		case "L":
			if m.activeTab == tabMonthly {
				m.date = m.date.AddDate(0, 1, 0)
				m.loadData()
				m.statusMsg = ""
			}
		case "g":
			if m.activeTab == tabCalories {
				m.inputMode = inputCalorieGoal
				m.input.Placeholder = "Daily calorie goal (e.g. 2000)"
				m.input.SetValue(strconv.Itoa(m.calories.Goal))
				m.input.Focus()
			}
		}
	}
	return m, nil
}

// ── Calendar handling ──────────────────────────────────

func (m Model) handleCalendar(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "c", "q":
		m.isCalendar = false
	case "enter":
		m.date = time.Date(m.calYear, m.calMonth, m.calDay, 0, 0, 0, 0, time.Local)
		m.loadData()
		m.isCalendar = false
		m.statusMsg = fmt.Sprintf("Jumped to %s", m.date.Format("Jan 02, 2006"))
	case "t":
		now := time.Now()
		m.calYear = now.Year()
		m.calMonth = now.Month()
		m.calDay = now.Day()
	case "h", "left":
		d := time.Date(m.calYear, m.calMonth, m.calDay, 0, 0, 0, 0, time.Local)
		d = d.AddDate(0, 0, -1)
		m.calYear = d.Year()
		m.calMonth = d.Month()
		m.calDay = d.Day()
	case "l", "right":
		d := time.Date(m.calYear, m.calMonth, m.calDay, 0, 0, 0, 0, time.Local)
		d = d.AddDate(0, 0, 1)
		m.calYear = d.Year()
		m.calMonth = d.Month()
		m.calDay = d.Day()
	case "k", "up":
		d := time.Date(m.calYear, m.calMonth, m.calDay, 0, 0, 0, 0, time.Local)
		d = d.AddDate(0, 0, -7)
		m.calYear = d.Year()
		m.calMonth = d.Month()
		m.calDay = d.Day()
	case "j", "down":
		d := time.Date(m.calYear, m.calMonth, m.calDay, 0, 0, 0, 0, time.Local)
		d = d.AddDate(0, 0, 7)
		m.calYear = d.Year()
		m.calMonth = d.Month()
		m.calDay = d.Day()
	case "H":
		m.calMonth--
		if m.calMonth < time.January {
			m.calMonth = time.December
			m.calYear--
		}
		maxDay := daysInMonth(m.calYear, m.calMonth)
		if m.calDay > maxDay {
			m.calDay = maxDay
		}
	case "L":
		m.calMonth++
		if m.calMonth > time.December {
			m.calMonth = time.January
			m.calYear++
		}
		maxDay := daysInMonth(m.calYear, m.calMonth)
		if m.calDay > maxDay {
			m.calDay = maxDay
		}
	}
	return m, nil
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local).Day()
}

// ── Cursor / list ──────────────────────────────────────

func (m *Model) cursorDown() {
	max := m.listLen() - 1
	if max < 0 {
		max = 0
	}
	if m.cursor < max {
		m.cursor++
	}
}

func (m *Model) cursorUp() {
	if m.cursor > 0 {
		m.cursor--
	}
}

func (m Model) listLen() int {
	switch m.activeTab {
	case tabCalories:
		return len(m.calories.Entries)
	case tabGym:
		return len(m.gym.Activities)
	case tabMedicine:
		return len(m.medicine.Entries)
	case tabJournal:
		return len(m.journal.Entries)
	default:
		return 0
	}
}

// ── Actions ────────────────────────────────────────────

func (m *Model) startAdd() {
	m.input.Reset()
	m.input.Focus()
	m.statusMsg = ""
	switch m.activeTab {
	case tabOverview, tabMonthly:
		return
	case tabCalories:
		m.inputMode = inputFoodName
		m.input.Placeholder = "Food name (e.g. Chicken Breast)"
	case tabGym:
		m.inputMode = inputGymName
		m.input.Placeholder = "Activity name (e.g. Running)"
	case tabMood:
		m.inputMode = inputMoodRating
		m.input.Placeholder = "Rating 1-5 (1=terrible, 5=great)"
	case tabMedicine:
		m.inputMode = inputMedicineName
		m.input.Placeholder = "Medicine name (e.g. Vitamin D)"
	case tabJournal:
		m.inputMode = inputJournalNote
		m.input.Placeholder = "What's on your mind?"
	}
}

func (m Model) handleInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.inputMode = inputNone
		m.input.Reset()
		m.statusMsg = ""
		return m, nil

	case "enter":
		val := strings.TrimSpace(m.input.Value())
		if val == "" && m.inputMode != inputMoodNote {
			m.statusMsg = "Input cannot be empty"
			return m, nil
		}

		switch m.inputMode {
		case inputFoodName:
			m.tmpFoodName = val
			m.inputMode = inputFoodCalories
			m.input.Reset()
			m.input.Placeholder = "Calories (e.g. 165)"
			m.input.Focus()

		case inputFoodCalories:
			n, err := strconv.Atoi(val)
			if err != nil || n < 0 {
				m.statusMsg = "Enter a valid number"
				return m, nil
			}
			m.tmpFoodCalories = n
			m.inputMode = inputFoodProtein
			m.input.Reset()
			m.input.Placeholder = "Protein in grams (e.g. 31), or 0"
			m.input.Focus()

		case inputFoodProtein:
			n, err := strconv.Atoi(val)
			if err != nil || n < 0 {
				m.statusMsg = "Enter a valid number"
				return m, nil
			}
			m.tmpFoodProtein = n
			m.inputMode = inputFoodFat
			m.input.Reset()
			m.input.Placeholder = "Fat in grams (e.g. 5), or 0"
			m.input.Focus()

		case inputFoodFat:
			n, err := strconv.Atoi(val)
			if err != nil || n < 0 {
				m.statusMsg = "Enter a valid number"
				return m, nil
			}
			m.tmpFoodFat = n
			m.inputMode = inputFoodCarbs
			m.input.Reset()
			m.input.Placeholder = "Carbs in grams (e.g. 20), or 0"
			m.input.Focus()

		case inputFoodCarbs:
			n, err := strconv.Atoi(val)
			if err != nil || n < 0 {
				m.statusMsg = "Enter a valid number"
				return m, nil
			}
			entry := data.FoodEntry{
				Name:     m.tmpFoodName,
				Calories: m.tmpFoodCalories,
				Protein:  m.tmpFoodProtein,
				Fat:      m.tmpFoodFat,
				Carbs:    n,
				Time:     time.Now().Format("15:04"),
			}
			m.calories.Entries = append(m.calories.Entries, entry)
			_ = data.SaveCalories(m.date, m.calories)
			m.inputMode = inputNone
			m.input.Reset()
			m.statusMsg = fmt.Sprintf("Added: %s (%d cal)", entry.Name, entry.Calories)

		case inputCalorieGoal:
			n, err := strconv.Atoi(val)
			if err != nil || n <= 0 {
				m.statusMsg = "Enter a valid positive number"
				return m, nil
			}
			m.calories.Goal = n
			_ = data.SaveCalories(m.date, m.calories)
			m.inputMode = inputNone
			m.input.Reset()
			m.statusMsg = fmt.Sprintf("Goal set to %d cal", n)

		case inputFastStart:
			m.calories.Fasting.Start = val
			m.inputMode = inputFastEnd
			m.input.Reset()
			m.input.Placeholder = "Fasting ends at (e.g. 12:00)"
			m.input.Focus()

		case inputFastEnd:
			m.calories.Fasting.End = val
			m.calories.Fasting.Active = true
			_ = data.SaveCalories(m.date, m.calories)
			m.inputMode = inputNone
			m.input.Reset()
			m.statusMsg = fmt.Sprintf("Fasting: %s → %s", m.calories.Fasting.Start, m.calories.Fasting.End)

		case inputGymName:
			m.tmpGymName = val
			m.inputMode = inputGymDuration
			m.input.Reset()
			m.input.Placeholder = "Duration in minutes (e.g. 30)"
			m.input.Focus()

		case inputGymDuration:
			n, err := strconv.Atoi(val)
			if err != nil || n < 0 {
				m.statusMsg = "Enter a valid number"
				return m, nil
			}
			m.tmpGymDuration = n
			m.inputMode = inputGymCalories
			m.input.Reset()
			m.input.Placeholder = "Calories burnt (e.g. 300)"
			m.input.Focus()

		case inputGymCalories:
			n, err := strconv.Atoi(val)
			if err != nil || n < 0 {
				m.statusMsg = "Enter a valid number"
				return m, nil
			}
			activity := data.GymActivity{
				Name:          m.tmpGymName,
				DurationMin:   m.tmpGymDuration,
				CaloriesBurnt: n,
				Time:          time.Now().Format("15:04"),
			}
			m.gym.Activities = append(m.gym.Activities, activity)
			_ = data.SaveGym(m.date, m.gym)
			m.inputMode = inputNone
			m.input.Reset()
			m.statusMsg = fmt.Sprintf("Added: %s (%d min, %d cal)", activity.Name, activity.DurationMin, activity.CaloriesBurnt)

		case inputMoodRating:
			n, err := strconv.Atoi(val)
			if err != nil || n < 1 || n > 5 {
				m.statusMsg = "Enter 1-5"
				return m, nil
			}
			m.mood.Rating = n
			m.inputMode = inputMoodNote
			m.input.Reset()
			m.input.Placeholder = "How was your day? (optional, Enter to skip)"
			m.input.Focus()

		case inputMoodNote:
			m.mood.Note = val
			m.mood.Time = time.Now().Format("15:04")
			_ = data.SaveMood(m.date, m.mood)
			m.inputMode = inputNone
			m.input.Reset()
			m.statusMsg = fmt.Sprintf("Mood: %s %s", ui.MoodEmoji(m.mood.Rating), ui.MoodLabel(m.mood.Rating))

		case inputMedicineName:
			m.tmpMedicineName = val
			m.inputMode = inputMedicineDosage
			m.input.Reset()
			m.input.Placeholder = "Dosage (e.g. 1000 IU, 500mg)"
			m.input.Focus()

		case inputMedicineDosage:
			entry := data.MedicineEntry{
				Name:   m.tmpMedicineName,
				Dosage: val,
				Time:   time.Now().Format("15:04"),
			}
			m.medicine.Entries = append(m.medicine.Entries, entry)
			_ = data.SaveMedicine(m.date, m.medicine)
			m.inputMode = inputNone
			m.input.Reset()
			m.statusMsg = fmt.Sprintf("Added: %s (%s)", entry.Name, entry.Dosage)

		case inputJournalNote:
			entry := data.JournalEntry{
				Note: val,
				Time: time.Now().Format("15:04"),
			}
			m.journal.Entries = append(m.journal.Entries, entry)
			_ = data.SaveJournal(m.date, m.journal)
			m.inputMode = inputNone
			m.input.Reset()
			m.statusMsg = "Journal entry added"
		}
		return m, nil

	default:
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
}

func (m *Model) deleteEntry() {
	switch m.activeTab {
	case tabCalories:
		if m.cursor < len(m.calories.Entries) {
			name := m.calories.Entries[m.cursor].Name
			m.calories.Entries = append(m.calories.Entries[:m.cursor], m.calories.Entries[m.cursor+1:]...)
			_ = data.SaveCalories(m.date, m.calories)
			if m.cursor > 0 && m.cursor >= len(m.calories.Entries) {
				m.cursor--
			}
			m.statusMsg = fmt.Sprintf("Deleted: %s", name)
		}
	case tabGym:
		if m.cursor < len(m.gym.Activities) {
			name := m.gym.Activities[m.cursor].Name
			m.gym.Activities = append(m.gym.Activities[:m.cursor], m.gym.Activities[m.cursor+1:]...)
			_ = data.SaveGym(m.date, m.gym)
			if m.cursor > 0 && m.cursor >= len(m.gym.Activities) {
				m.cursor--
			}
			m.statusMsg = fmt.Sprintf("Deleted: %s", name)
		}
	case tabMood:
		m.mood = data.MoodEntry{}
		_ = data.SaveMood(m.date, m.mood)
		m.statusMsg = "Mood cleared"
	case tabMedicine:
		if m.cursor < len(m.medicine.Entries) {
			name := m.medicine.Entries[m.cursor].Name
			m.medicine.Entries = append(m.medicine.Entries[:m.cursor], m.medicine.Entries[m.cursor+1:]...)
			_ = data.SaveMedicine(m.date, m.medicine)
			if m.cursor > 0 && m.cursor >= len(m.medicine.Entries) {
				m.cursor--
			}
			m.statusMsg = fmt.Sprintf("Deleted: %s", name)
		}
	case tabJournal:
		if m.cursor < len(m.journal.Entries) {
			note := truncate(m.journal.Entries[m.cursor].Note, 30)
			m.journal.Entries = append(m.journal.Entries[:m.cursor], m.journal.Entries[m.cursor+1:]...)
			_ = data.SaveJournal(m.date, m.journal)
			if m.cursor > 0 && m.cursor >= len(m.journal.Entries) {
				m.cursor--
			}
			m.statusMsg = fmt.Sprintf("Deleted: %s", note)
		}
	}
}

func (m *Model) toggleFasting() {
	if m.calories.Fasting.Active {
		m.calories.Fasting.Active = false
		_ = data.SaveCalories(m.date, m.calories)
		m.statusMsg = "Fasting disabled"
	} else if m.calories.Fasting.Start != "" && m.calories.Fasting.End != "" {
		m.calories.Fasting.Active = true
		_ = data.SaveCalories(m.date, m.calories)
		m.statusMsg = "Fasting enabled"
	} else {
		m.inputMode = inputFastStart
		m.input.Reset()
		m.input.Placeholder = "Fasting starts at (e.g. 20:00)"
		m.input.Focus()
	}
}

// ── View ───────────────────────────────────────────────

func (m Model) View() string {
	if m.isHelp {
		return m.viewHelp()
	}
	if m.isCalendar {
		return m.viewCalendar()
	}

	contentWidth := min(m.width-4, 70)

	var sections []string

	// ── Header ──
	sections = append(sections, m.viewHeader(contentWidth))

	// ── Tabs ──
	sections = append(sections, m.viewTabs(contentWidth))

	// ── Content ──
	var content string
	switch m.activeTab {
	case tabOverview:
		content = m.viewOverview(contentWidth)
	case tabMonthly:
		content = m.viewMonthly(contentWidth)
	case tabCalories:
		content = m.viewCalories(contentWidth)
	case tabGym:
		content = m.viewGym(contentWidth)
	case tabMood:
		content = m.viewMood()
	case tabMedicine:
		content = m.viewMedicine(contentWidth)
	case tabJournal:
		content = m.viewJournal(contentWidth)
	}
	contentBox := ui.ContentBoxStyle.Width(contentWidth).Render(content)
	sections = append(sections, contentBox)

	// ── Input ──
	if m.inputMode != inputNone {
		inputContent := ui.AccentStyle.Render(m.inputPromptLabel()) + "\n" + m.input.View()
		inputBox := ui.InputBoxStyle.Width(contentWidth).Render(inputContent)
		sections = append(sections, inputBox)
	}

	// ── Status ──
	if m.statusMsg != "" {
		sections = append(sections, ui.StatusMsgStyle.Render("  "+m.statusMsg))
	}

	// ── Footer ──
	sections = append(sections, m.viewFooter(contentWidth))

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m Model) viewHeader(width int) string {
	title := ui.HeaderBoxStyle.Render("atlas.guide")

	dateStr := m.date.Format("Mon, Jan 02 2006")
	dateDisplay := ui.DateBoxStyle.Render("◀ " + dateStr + " ▶")

	isToday := m.date.Format("2006-01-02") == time.Now().Format("2006-01-02")

	right := dateDisplay
	if isToday {
		right = dateDisplay + "  " + ui.TodayBadgeStyle.Render(" TODAY ")
	}

	gap := width - lipgloss.Width(title) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}

	return title + strings.Repeat(" ", gap) + right
}

func (m Model) viewTabs(width int) string {
	tabNames := []string{"📊 Overview", "📅 Monthly", "🍎 Calories", "💪 Gym", "😊 Mood", "💊 Medicine", "🥛 Journal"}
	var tabs []string
	for i, name := range tabNames {
		if tab(i) == m.activeTab {
			tabs = append(tabs, ui.ActiveTabStyle.Render(name))
		} else {
			tabs = append(tabs, ui.InactiveTabStyle.Render(name))
		}
	}

	row := lipgloss.JoinHorizontal(lipgloss.Bottom, tabs...)

	// Fill remaining width with bottom border
	rowWidth := lipgloss.Width(row)
	gap := width - rowWidth
	if gap > 0 {
		gapStr := ui.TabGapStyle.Render(strings.Repeat(" ", gap))
		row = lipgloss.JoinHorizontal(lipgloss.Bottom, row, gapStr)
	}

	return row
}

func (m Model) viewFooter(width int) string {
	if m.inputMode != inputNone {
		return ui.FooterStyle.Render(
			fmtKey("Enter", "confirm") + "  " + fmtKey("Esc", "cancel"),
		)
	}

	common := fmtKey("1-7", "tabs") + "  " +
		fmtKey("←/→", "date") + "  " +
		fmtKey("t", "today") + "  " +
		fmtKey("c", "calendar") + "  "

	var specific string
	switch m.activeTab {
	case tabOverview:
		specific = ""
	case tabMonthly:
		specific = fmtKey("H/L", "month")
	case tabCalories:
		specific = fmtKey("a", "add") + "  " +
			fmtKey("d", "del") + "  " +
			fmtKey("g", "goal") + "  " +
			fmtKey("f", "fast")
	case tabGym:
		specific = fmtKey("a", "add") + "  " +
			fmtKey("d", "del")
	case tabMood:
		specific = fmtKey("a", "mood") + "  " +
			fmtKey("d", "clear")
	case tabMedicine:
		specific = fmtKey("a", "add") + "  " +
			fmtKey("d", "del")
	case tabJournal:
		specific = fmtKey("a", "add") + "  " +
			fmtKey("d", "del")
	}

	right := fmtKey("?", "help") + "  " + fmtKey("q", "quit")

	return ui.FooterStyle.Render(common + specific + "  " + right)
}

func fmtKey(key, desc string) string {
	return ui.FooterKeyStyle.Render(key) + " " + ui.FooterDescStyle.Render(desc)
}

func (m Model) inputPromptLabel() string {
	switch m.inputMode {
	case inputFoodName:
		return "Enter food name:"
	case inputFoodCalories:
		return fmt.Sprintf("Calories for '%s':", m.tmpFoodName)
	case inputFoodProtein:
		return fmt.Sprintf("Protein (g) for '%s':", m.tmpFoodName)
	case inputFoodFat:
		return fmt.Sprintf("Fat (g) for '%s':", m.tmpFoodName)
	case inputFoodCarbs:
		return fmt.Sprintf("Carbs (g) for '%s':", m.tmpFoodName)
	case inputCalorieGoal:
		return "Set daily calorie goal:"
	case inputFastStart:
		return "Fasting window start time:"
	case inputFastEnd:
		return "Fasting window end time:"
	case inputGymName:
		return "Enter activity name:"
	case inputGymDuration:
		return fmt.Sprintf("Duration (min) for '%s':", m.tmpGymName)
	case inputGymCalories:
		return fmt.Sprintf("Calories burnt for '%s':", m.tmpGymName)
	case inputMoodRating:
		return "Rate your day (1-5):"
	case inputMoodNote:
		return fmt.Sprintf("Note for %s %s:", ui.MoodEmoji(m.mood.Rating), ui.MoodLabel(m.mood.Rating))
	case inputMedicineName:
		return "Enter medicine name:"
	case inputMedicineDosage:
		return fmt.Sprintf("Dosage for '%s':", m.tmpMedicineName)
	case inputJournalNote:
		return "Write a journal note:"
	default:
		return ""
	}
}

// ── Overview view ──────────────────────────────────────

func (m Model) viewOverview(width int) string {
	var b strings.Builder

	consumed := data.TotalCaloriesConsumed(m.calories)
	protein := data.TotalProtein(m.calories)
	burnt := data.TotalCaloriesBurnt(m.gym)
	net := data.NetCalories(m.calories, m.gym)
	remaining := data.CalorieRemaining(m.calories, m.gym)
	totalDuration := data.TotalGymDuration(m.gym)

	// ── Calories card ──
	b.WriteString(ui.TitleStyle.Render("🍎 Calories"))
	b.WriteString("\n")

	if len(m.calories.Entries) == 0 && !m.calories.Fasting.Active {
		b.WriteString(ui.EmptyStyle.Render("  No food logged yet."))
	} else {
		// Budget bar
		pct := float64(consumed) / float64(m.calories.Goal)
		if pct > 1 {
			pct = 1
		}
		barWidth := min(width-6, 50)
		m.progressBar.Width = barWidth
		if pct > 0.9 {
			m.progressBar.FullColor = string(ui.ColorDanger)
		} else if pct > 0.7 {
			m.progressBar.FullColor = string(ui.ColorWarning)
		} else {
			m.progressBar.FullColor = string(ui.ColorSuccess)
		}
		b.WriteString("  " + m.progressBar.ViewAs(pct) + fmt.Sprintf("  %d%%", int(pct*100)))
		b.WriteString("\n\n")

		budgetText := fmt.Sprintf("%d / %d cal", consumed, m.calories.Goal)
		b.WriteString("  " + ui.LabelStyle.Render("Budget ") + ui.ValueStyle.Render(budgetText) + "  ")
		if remaining >= 0 {
			b.WriteString(ui.DeficitStyle.Render(fmt.Sprintf("▼ %d remaining", remaining)))
		} else {
			b.WriteString(ui.SurplusStyle.Render(fmt.Sprintf("▲ %d over!", -remaining)))
		}
		b.WriteString("\n")
		b.WriteString("  " +
			ui.LabelStyle.Render("Protein ") + ui.ValueStyle.Render(fmt.Sprintf("%dg", protein)) + "  │  " +
			ui.LabelStyle.Render("Fat ") + ui.ValueStyle.Render(fmt.Sprintf("%dg", data.TotalFat(m.calories))) + "  │  " +
			ui.LabelStyle.Render("Carbs ") + ui.ValueStyle.Render(fmt.Sprintf("%dg", data.TotalCarbs(m.calories))) + "  │  " +
			ui.LabelStyle.Render("Net ") + ui.ValueStyle.Render(fmt.Sprintf("%d cal", net)))
		b.WriteString("\n")

		// Top 3 foods
		if len(m.calories.Entries) > 0 {
			b.WriteString("  " + ui.LabelStyle.Render("Foods "))
			limit := len(m.calories.Entries)
			if limit > 3 {
				limit = 3
			}
			var names []string
			for _, e := range m.calories.Entries[:limit] {
				names = append(names, fmt.Sprintf("%s (%d)", e.Name, e.Calories))
			}
			b.WriteString(ui.InfoStyle.Render(strings.Join(names, ", ")))
			if len(m.calories.Entries) > 3 {
				b.WriteString(ui.InfoStyle.Render(fmt.Sprintf(" +%d more", len(m.calories.Entries)-3)))
			}
			b.WriteString("\n")
		}

		// Fasting
		if m.calories.Fasting.Active {
			b.WriteString("  🕐 " +
				ui.WarningStyle.Render(m.calories.Fasting.Start) +
				ui.LabelStyle.Render(" → ") +
				ui.SuccessStyle.Render(m.calories.Fasting.End) +
				"  " + ui.SuccessStyle.Render("● fasting"))
			b.WriteString("\n")
		}
	}

	// ── Separator ──
	b.WriteString("\n")
	b.WriteString(ui.InfoStyle.Render(strings.Repeat("─", min(width-4, 60))))
	b.WriteString("\n\n")

	// ── Gym card ──
	b.WriteString(ui.TitleStyle.Render("💪 Gym"))
	b.WriteString("\n")

	if len(m.gym.Activities) == 0 {
		b.WriteString(ui.EmptyStyle.Render("  No activities logged yet."))
	} else {
		b.WriteString("  " +
			ui.LabelStyle.Render("Burnt ") + ui.SuccessStyle.Render(fmt.Sprintf("%d cal", burnt)) + "  │  " +
			ui.LabelStyle.Render("Duration ") + ui.ValueStyle.Render(fmt.Sprintf("%d min", totalDuration)) + "  │  " +
			ui.LabelStyle.Render("Sessions ") + ui.ValueStyle.Render(fmt.Sprintf("%d", len(m.gym.Activities))))
		b.WriteString("\n")

		// List activities
		b.WriteString("  " + ui.LabelStyle.Render("Log "))
		var acts []string
		for _, a := range m.gym.Activities {
			acts = append(acts, fmt.Sprintf("%s (%dm, %d cal)", a.Name, a.DurationMin, a.CaloriesBurnt))
		}
		b.WriteString(ui.InfoStyle.Render(strings.Join(acts, ", ")))
		b.WriteString("\n")
	}

	// ── Separator ──
	b.WriteString("\n")
	b.WriteString(ui.InfoStyle.Render(strings.Repeat("─", min(width-4, 60))))
	b.WriteString("\n\n")

	// ── Mood card ──
	b.WriteString(ui.TitleStyle.Render("😊 Mood"))
	b.WriteString("\n")

	if m.mood.Rating == 0 {
		b.WriteString(ui.EmptyStyle.Render("  No mood logged yet."))
	} else {
		style := ui.MoodStyles[m.mood.Rating-1]
		emoji := ui.MoodEmoji(m.mood.Rating)
		label := ui.MoodLabel(m.mood.Rating)

		var dots []string
		for i := 1; i <= 5; i++ {
			if i <= m.mood.Rating {
				dots = append(dots, style.Render("●"))
			} else {
				dots = append(dots, ui.InfoStyle.Render("○"))
			}
		}

		b.WriteString("  " + style.Render(fmt.Sprintf("%s %s", emoji, label)) +
			"  " + strings.Join(dots, " ") +
			"  " + ui.LabelStyle.Render(fmt.Sprintf("%d/5", m.mood.Rating)))
		b.WriteString("\n")

		if m.mood.Note != "" {
			b.WriteString("  " + ui.LabelStyle.Render("Note ") + ui.ValueStyle.Render(m.mood.Note))
			b.WriteString("\n")
		}
	}

	// ── Separator ──
	b.WriteString("\n")
	b.WriteString(ui.InfoStyle.Render(strings.Repeat("─", min(width-4, 60))))
	b.WriteString("\n\n")

	// ── Medicine card ──
	b.WriteString(ui.TitleStyle.Render("💊 Medicine"))
	b.WriteString("\n")

	if len(m.medicine.Entries) == 0 {
		b.WriteString(ui.EmptyStyle.Render("  No medicine logged yet."))
	} else {
		b.WriteString("  " +
			ui.LabelStyle.Render("Doses ") + ui.ValueStyle.Render(fmt.Sprintf("%d", len(m.medicine.Entries))))
		b.WriteString("\n")

		b.WriteString("  " + ui.LabelStyle.Render("Log "))
		var meds []string
		for _, e := range m.medicine.Entries {
			meds = append(meds, fmt.Sprintf("%s (%s)", e.Name, e.Dosage))
		}
		b.WriteString(ui.InfoStyle.Render(strings.Join(meds, ", ")))
		b.WriteString("\n")
	}

	// ── Separator ──
	b.WriteString("\n")
	b.WriteString(ui.InfoStyle.Render(strings.Repeat("─", min(width-4, 60))))
	b.WriteString("\n\n")

	// ── Journal card ──
	b.WriteString(ui.TitleStyle.Render("📓 Journal"))
	b.WriteString("\n")

	if len(m.journal.Entries) == 0 {
		b.WriteString(ui.EmptyStyle.Render("  No journal entries yet."))
	} else {
		b.WriteString("  " +
			ui.LabelStyle.Render("Entries ") + ui.ValueStyle.Render(fmt.Sprintf("%d", len(m.journal.Entries))))
		b.WriteString("\n")

		// Show latest entry
		latest := m.journal.Entries[len(m.journal.Entries)-1]
		b.WriteString("  " + ui.LabelStyle.Render("Latest ") + ui.InfoStyle.Render(truncate(latest.Note, 40)))
		b.WriteString("\n")
	}

	// ── Separator ──
	b.WriteString("\n")
	b.WriteString(ui.InfoStyle.Render(strings.Repeat("─", min(width-4, 60))))
	b.WriteString("\n\n")

	// ── Day score ──
	b.WriteString(ui.TitleStyle.Render("📈 Day Score"))
	b.WriteString("\n")

	filled := 0
	if len(m.calories.Entries) > 0 {
		filled++
	}
	if len(m.gym.Activities) > 0 {
		filled++
	}
	if m.mood.Rating > 0 {
		filled++
	}
	if len(m.medicine.Entries) > 0 {
		filled++
	}
	if len(m.journal.Entries) > 0 {
		filled++
	}

	var checks []string
	if len(m.calories.Entries) > 0 {
		checks = append(checks, ui.SuccessStyle.Render("✓ Calories"))
	} else {
		checks = append(checks, ui.InfoStyle.Render("○ Calories"))
	}
	if len(m.gym.Activities) > 0 {
		checks = append(checks, ui.SuccessStyle.Render("✓ Gym"))
	} else {
		checks = append(checks, ui.InfoStyle.Render("○ Gym"))
	}
	if m.mood.Rating > 0 {
		checks = append(checks, ui.SuccessStyle.Render("✓ Mood"))
	} else {
		checks = append(checks, ui.InfoStyle.Render("○ Mood"))
	}
	if len(m.medicine.Entries) > 0 {
		checks = append(checks, ui.SuccessStyle.Render("✓ Medicine"))
	} else {
		checks = append(checks, ui.InfoStyle.Render("○ Medicine"))
	}
	if len(m.journal.Entries) > 0 {
		checks = append(checks, ui.SuccessStyle.Render("✓ Journal"))
	} else {
		checks = append(checks, ui.InfoStyle.Render("○ Journal"))
	}

	b.WriteString("  " + strings.Join(checks, "   "))
	b.WriteString("   " + ui.AccentStyle.Render(fmt.Sprintf("%d/5 tracked", filled)))

	return b.String()
}

// ── Calories view ──────────────────────────────────────

func (m Model) viewCalories(width int) string {
	var b strings.Builder

	consumed := data.TotalCaloriesConsumed(m.calories)
	protein := data.TotalProtein(m.calories)
	fat := data.TotalFat(m.calories)
	carbs := data.TotalCarbs(m.calories)
	burnt := data.TotalCaloriesBurnt(m.gym)
	net := data.NetCalories(m.calories, m.gym)
	remaining := data.CalorieRemaining(m.calories, m.gym)

	b.WriteString(ui.TitleStyle.Render("Daily Summary"))
	b.WriteString("\n")

	// Budget line
	budgetText := fmt.Sprintf("%d / %d cal", consumed, m.calories.Goal)
	b.WriteString("  " + ui.LabelStyle.Render("Budget: ") + ui.ValueStyle.Render(budgetText) + "  ")
	if remaining >= 0 {
		b.WriteString(ui.DeficitStyle.Render(fmt.Sprintf("▼ %d remaining", remaining)))
	} else {
		b.WriteString(ui.SurplusStyle.Render(fmt.Sprintf("▲ %d over!", -remaining)))
	}
	b.WriteString("\n\n")

	// Progress bar
	pct := float64(consumed) / float64(m.calories.Goal)
	if pct > 1 {
		pct = 1
	}
	barWidth := min(width-6, 50)
	m.progressBar.Width = barWidth
	if pct > 0.9 {
		m.progressBar.FullColor = string(ui.ColorDanger)
	} else if pct > 0.7 {
		m.progressBar.FullColor = string(ui.ColorWarning)
	} else {
		m.progressBar.FullColor = string(ui.ColorSuccess)
	}
	b.WriteString("  " + m.progressBar.ViewAs(pct) + fmt.Sprintf("  %d%%", int(pct*100)))
	b.WriteString("\n\n")

	// Stats row
	stats := []string{
		ui.LabelStyle.Render("Protein ") + ui.ValueStyle.Render(fmt.Sprintf("%dg", protein)),
		ui.LabelStyle.Render("Fat ") + ui.ValueStyle.Render(fmt.Sprintf("%dg", fat)),
		ui.LabelStyle.Render("Carbs ") + ui.ValueStyle.Render(fmt.Sprintf("%dg", carbs)),
		ui.LabelStyle.Render("Burnt ") + ui.ValueStyle.Render(fmt.Sprintf("%d cal", burnt)),
		ui.LabelStyle.Render("Net ") + ui.ValueStyle.Render(fmt.Sprintf("%d cal", net)),
	}
	b.WriteString("  " + strings.Join(stats, "  │  "))
	b.WriteString("\n")

	// Fasting
	if m.calories.Fasting.Active {
		b.WriteString("\n")
		b.WriteString("  🕐 " + ui.LabelStyle.Render("Fasting ") +
			ui.WarningStyle.Render(m.calories.Fasting.Start) +
			ui.LabelStyle.Render(" → ") +
			ui.SuccessStyle.Render(m.calories.Fasting.End) +
			"  " + ui.SuccessStyle.Render("● active"))
		b.WriteString("\n")
	}

	// Food list
	b.WriteString("\n")
	b.WriteString(ui.TitleStyle.Render("Food Log"))
	b.WriteString("\n")

	if len(m.calories.Entries) == 0 {
		b.WriteString(ui.EmptyStyle.Render("  No food entries yet. Press 'a' to add."))
	} else {
		for i, e := range m.calories.Entries {
			cursor := "  "
			style := ui.NormalStyle
			if i == m.cursor {
				cursor = ui.CursorStyle.Render("▸ ")
				style = ui.SelectedStyle
			}
			name := fmt.Sprintf("%-22s", truncate(e.Name, 22))
			line := fmt.Sprintf("%s%s %s  %s  %s  %s  %s",
				cursor,
				style.Render(name),
				ui.ValueStyle.Render(fmt.Sprintf("%4d cal", e.Calories)),
				ui.LabelStyle.Render(fmt.Sprintf("%3dg P", e.Protein)),
				ui.LabelStyle.Render(fmt.Sprintf("%3dg F", e.Fat)),
				ui.LabelStyle.Render(fmt.Sprintf("%3dg C", e.Carbs)),
				ui.InfoStyle.Render("@"+e.Time),
			)
			b.WriteString(line + "\n")
		}
	}

	return b.String()
}

// ── Gym view ───────────────────────────────────────────

func (m Model) viewGym(width int) string {
	var b strings.Builder

	totalBurnt := data.TotalCaloriesBurnt(m.gym)
	totalDuration := data.TotalGymDuration(m.gym)

	b.WriteString(ui.TitleStyle.Render("Gym Summary"))
	b.WriteString("\n")

	stats := []string{
		ui.LabelStyle.Render("Burnt ") + ui.SuccessStyle.Render(fmt.Sprintf("%d cal", totalBurnt)),
		ui.LabelStyle.Render("Duration ") + ui.ValueStyle.Render(fmt.Sprintf("%d min", totalDuration)),
		ui.LabelStyle.Render("Activities ") + ui.ValueStyle.Render(fmt.Sprintf("%d", len(m.gym.Activities))),
	}
	b.WriteString("  " + strings.Join(stats, "  │  "))
	b.WriteString("\n\n")

	b.WriteString(ui.TitleStyle.Render("Activities"))
	b.WriteString("\n")

	if len(m.gym.Activities) == 0 {
		b.WriteString(ui.EmptyStyle.Render("  No activities yet. Press 'a' to add."))
	} else {
		for i, a := range m.gym.Activities {
			cursor := "  "
			style := ui.NormalStyle
			if i == m.cursor {
				cursor = ui.CursorStyle.Render("▸ ")
				style = ui.SelectedStyle
			}
			name := fmt.Sprintf("%-22s", truncate(a.Name, 22))
			line := fmt.Sprintf("%s%s %s  %s  %s",
				cursor,
				style.Render(name),
				ui.ValueStyle.Render(fmt.Sprintf("%3d min", a.DurationMin)),
				ui.SuccessStyle.Render(fmt.Sprintf("%4d cal", a.CaloriesBurnt)),
				ui.InfoStyle.Render("@"+a.Time),
			)
			b.WriteString(line + "\n")
		}
	}

	return b.String()
}

// ── Medicine view ──────────────────────────────────────

func (m Model) viewMedicine(width int) string {
	var b strings.Builder

	b.WriteString(ui.TitleStyle.Render("Medicine Summary"))
	b.WriteString("\n")

	stats := []string{
		ui.LabelStyle.Render("Doses ") + ui.ValueStyle.Render(fmt.Sprintf("%d", len(m.medicine.Entries))),
	}
	b.WriteString("  " + strings.Join(stats, "  │  "))
	b.WriteString("\n\n")

	b.WriteString(ui.TitleStyle.Render("Medicine Log"))
	b.WriteString("\n")

	if len(m.medicine.Entries) == 0 {
		b.WriteString(ui.EmptyStyle.Render("  No medicine logged yet. Press 'a' to add."))
	} else {
		for i, e := range m.medicine.Entries {
			cursor := "  "
			style := ui.NormalStyle
			if i == m.cursor {
				cursor = ui.CursorStyle.Render("▸ ")
				style = ui.SelectedStyle
			}
			name := fmt.Sprintf("%-22s", truncate(e.Name, 22))
			line := fmt.Sprintf("%s%s %s  %s",
				cursor,
				style.Render(name),
				ui.ValueStyle.Render(fmt.Sprintf("%-12s", e.Dosage)),
				ui.InfoStyle.Render("@"+e.Time),
			)
			b.WriteString(line + "\n")
		}
	}

	return b.String()
}

// ── Journal view ─────────────────────────────────────────

func (m Model) viewJournal(width int) string {
	var b strings.Builder

	b.WriteString(ui.TitleStyle.Render("Journal"))
	b.WriteString("\n")

	stats := []string{
		ui.LabelStyle.Render("Entries ") + ui.ValueStyle.Render(fmt.Sprintf("%d", len(m.journal.Entries))),
	}
	b.WriteString("  " + strings.Join(stats, "  │  "))
	b.WriteString("\n\n")

	b.WriteString(ui.TitleStyle.Render("Notes"))
	b.WriteString("\n")

	if len(m.journal.Entries) == 0 {
		b.WriteString(ui.EmptyStyle.Render("  No journal entries yet. Press 'a' to write."))
	} else {
		for i, e := range m.journal.Entries {
			cursor := "  "
			style := ui.NormalStyle
			if i == m.cursor {
				cursor = ui.CursorStyle.Render("▸ ")
				style = ui.SelectedStyle
			}
			note := fmt.Sprintf("%-40s", truncate(e.Note, 40))
			line := fmt.Sprintf("%s%s  %s",
				cursor,
				style.Render(note),
				ui.InfoStyle.Render("@"+e.Time),
			)
			b.WriteString(line + "\n")
		}
	}

	return b.String()
}

// ── Mood view ──────────────────────────────────────────

func (m Model) viewMood() string {
	var b strings.Builder

	b.WriteString(ui.TitleStyle.Render("Today's Mood"))
	b.WriteString("\n")

	if m.mood.Rating == 0 {
		b.WriteString(ui.EmptyStyle.Render("  No mood entry yet. Press 'a' to log your mood."))
		b.WriteString("\n\n")
		b.WriteString("  " + ui.LabelStyle.Render("Rating scale:") + "\n")
		for i := 1; i <= 5; i++ {
			style := ui.MoodStyles[i-1]
			bar := strings.Repeat("█", i*3) + strings.Repeat("░", 15-i*3)
			b.WriteString(fmt.Sprintf("   %s  %s %s  %s\n",
				style.Render(fmt.Sprintf("%d", i)),
				ui.MoodEmoji(i),
				style.Render(fmt.Sprintf("%-10s", ui.MoodLabel(i))),
				style.Render(bar),
			))
		}
	} else {
		style := ui.MoodStyles[m.mood.Rating-1]
		emoji := ui.MoodEmoji(m.mood.Rating)
		label := ui.MoodLabel(m.mood.Rating)

		// Big centered emoji + label
		b.WriteString("\n")
		b.WriteString(style.Render(fmt.Sprintf("      %s", emoji)) + "\n")
		b.WriteString(style.Render(fmt.Sprintf("      %s", label)) + "\n\n")

		// Rating dots row
		var dots []string
		for i := 1; i <= 5; i++ {
			if i <= m.mood.Rating {
				dots = append(dots, style.Render("●"))
			} else {
				dots = append(dots, ui.InfoStyle.Render("○"))
			}
		}
		b.WriteString("      " + strings.Join(dots, " ") + "  " + ui.LabelStyle.Render(fmt.Sprintf("%d/5", m.mood.Rating)) + "\n")

		// Note
		if m.mood.Note != "" {
			b.WriteString("\n")
			b.WriteString("  " + ui.LabelStyle.Render("Note") + "\n")
			b.WriteString("  " + ui.ValueStyle.Render(m.mood.Note) + "\n")
		}

		// Timestamp
		if m.mood.Time != "" {
			b.WriteString("\n")
			b.WriteString("  " + ui.InfoStyle.Render("Logged at @"+m.mood.Time))
		}
	}

	return b.String()
}

// ── Monthly view ──────────────────────────────────────

func (m Model) viewMonthly(width int) string {
	var b strings.Builder

	year := m.date.Year()
	month := m.date.Month()
	summaries := data.LoadMonthSummaries(year, month)
	today := time.Now()

	b.WriteString(ui.TitleStyle.Render(fmt.Sprintf("📅 %s %d", month.String(), year)))
	b.WriteString("\n\n")

	// ── Aggregate stats ──
	var totalConsumed, totalBurnt, totalProtein, totalGymMin, totalSessions int
	var daysTracked, daysWithGym, moodSum, moodCount int
	var totalMedicineDoses, daysWithMedicine, totalJournalItems, daysWithJournal int
	for _, s := range summaries {
		if !s.HasData {
			continue
		}
		daysTracked++
		totalConsumed += s.CalConsumed
		totalBurnt += s.CalBurnt
		totalProtein += s.Protein
		totalGymMin += s.GymDuration
		totalSessions += s.GymSessions
		if s.GymSessions > 0 {
			daysWithGym++
		}
		if s.MoodRating > 0 {
			moodSum += s.MoodRating
			moodCount++
		}
		if s.MedicineCount > 0 {
			totalMedicineDoses += s.MedicineCount
			daysWithMedicine++
		}
		if s.JournalCount > 0 {
			totalJournalItems += s.JournalCount
			daysWithJournal++
		}
	}

	// ── Summary cards ──
	// Calories
	b.WriteString(ui.TitleStyle.Render("🍎 Calories"))
	b.WriteString("\n")
	if daysTracked == 0 {
		b.WriteString(ui.EmptyStyle.Render("  No data this month."))
	} else {
		avgConsumed := totalConsumed / daysTracked
		avgNet := (totalConsumed - totalBurnt) / daysTracked
		avgProtein := totalProtein / daysTracked

		b.WriteString("  " +
			ui.LabelStyle.Render("Total ") + ui.ValueStyle.Render(fmt.Sprintf("%d cal", totalConsumed)) + "  │  " +
			ui.LabelStyle.Render("Avg/day ") + ui.ValueStyle.Render(fmt.Sprintf("%d cal", avgConsumed)) + "  │  " +
			ui.LabelStyle.Render("Days ") + ui.ValueStyle.Render(fmt.Sprintf("%d", daysTracked)))
		b.WriteString("\n")
		b.WriteString("  " +
			ui.LabelStyle.Render("Avg Net ") + ui.ValueStyle.Render(fmt.Sprintf("%d cal", avgNet)) + "  │  " +
			ui.LabelStyle.Render("Avg Protein ") + ui.ValueStyle.Render(fmt.Sprintf("%dg", avgProtein)) + "  │  " +
			ui.LabelStyle.Render("Total Protein ") + ui.ValueStyle.Render(fmt.Sprintf("%dg", totalProtein)))
	}
	b.WriteString("\n\n")

	b.WriteString(ui.InfoStyle.Render(strings.Repeat("─", min(width-4, 60))))
	b.WriteString("\n\n")

	// Gym
	b.WriteString(ui.TitleStyle.Render("💪 Gym"))
	b.WriteString("\n")
	if daysWithGym == 0 {
		b.WriteString(ui.EmptyStyle.Render("  No gym data this month."))
	} else {
		avgDuration := totalGymMin / daysWithGym
		b.WriteString("  " +
			ui.LabelStyle.Render("Total Burnt ") + ui.SuccessStyle.Render(fmt.Sprintf("%d cal", totalBurnt)) + "  │  " +
			ui.LabelStyle.Render("Total Time ") + ui.ValueStyle.Render(fmt.Sprintf("%dh %dm", totalGymMin/60, totalGymMin%60)))
		b.WriteString("\n")
		b.WriteString("  " +
			ui.LabelStyle.Render("Sessions ") + ui.ValueStyle.Render(fmt.Sprintf("%d", totalSessions)) + "  │  " +
			ui.LabelStyle.Render("Active Days ") + ui.ValueStyle.Render(fmt.Sprintf("%d", daysWithGym)) + "  │  " +
			ui.LabelStyle.Render("Avg/session ") + ui.ValueStyle.Render(fmt.Sprintf("%d min", avgDuration)))
	}
	b.WriteString("\n\n")

	b.WriteString(ui.InfoStyle.Render(strings.Repeat("─", min(width-4, 60))))
	b.WriteString("\n\n")

	// Mood
	b.WriteString(ui.TitleStyle.Render("😊 Mood"))
	b.WriteString("\n")
	if moodCount == 0 {
		b.WriteString(ui.EmptyStyle.Render("  No mood data this month."))
	} else {
		avgMood := float64(moodSum) / float64(moodCount)
		avgRounded := int(avgMood + 0.5)
		if avgRounded < 1 {
			avgRounded = 1
		}
		if avgRounded > 5 {
			avgRounded = 5
		}
		style := ui.MoodStyles[avgRounded-1]

		b.WriteString("  " +
			ui.LabelStyle.Render("Average ") + style.Render(fmt.Sprintf("%s %.1f/5", ui.MoodEmoji(avgRounded), avgMood)) + "  │  " +
			ui.LabelStyle.Render("Days Logged ") + ui.ValueStyle.Render(fmt.Sprintf("%d", moodCount)))
		b.WriteString("\n")

		// Mood distribution
		b.WriteString("  " + ui.LabelStyle.Render("Distribution "))
		dist := [5]int{}
		for _, s := range summaries {
			if s.MoodRating >= 1 && s.MoodRating <= 5 {
				dist[s.MoodRating-1]++
			}
		}
		for i := 0; i < 5; i++ {
			ms := ui.MoodStyles[i]
			b.WriteString(ms.Render(fmt.Sprintf("%s%d ", ui.MoodEmoji(i+1), dist[i])))
		}
	}
	b.WriteString("\n\n")

	b.WriteString(ui.InfoStyle.Render(strings.Repeat("─", min(width-4, 60))))
	b.WriteString("\n\n")

	// Medicine
	b.WriteString(ui.TitleStyle.Render("💊 Medicine"))
	b.WriteString("\n")
	if daysWithMedicine == 0 {
		b.WriteString(ui.EmptyStyle.Render("  No medicine data this month."))
	} else {
		b.WriteString("  " +
			ui.LabelStyle.Render("Total Doses ") + ui.ValueStyle.Render(fmt.Sprintf("%d", totalMedicineDoses)) + "  │  " +
			ui.LabelStyle.Render("Days Logged ") + ui.ValueStyle.Render(fmt.Sprintf("%d", daysWithMedicine)))
	}
	b.WriteString("\n\n")

	b.WriteString(ui.InfoStyle.Render(strings.Repeat("─", min(width-4, 60))))
	b.WriteString("\n\n")

	// Journal
	b.WriteString(ui.TitleStyle.Render("📓 Journal"))
	b.WriteString("\n")
	if daysWithJournal == 0 {
		b.WriteString(ui.EmptyStyle.Render("  No journal data this month."))
	} else {
		b.WriteString("  " +
			ui.LabelStyle.Render("Total Entries ") + ui.ValueStyle.Render(fmt.Sprintf("%d", totalJournalItems)) + "  │  " +
			ui.LabelStyle.Render("Days Logged ") + ui.ValueStyle.Render(fmt.Sprintf("%d", daysWithJournal)))
	}
	b.WriteString("\n\n")

	b.WriteString(ui.InfoStyle.Render(strings.Repeat("─", min(width-4, 60))))
	b.WriteString("\n\n")
	// ── Day-by-day heatmap ──
	b.WriteString(ui.TitleStyle.Render("Day-by-Day"))
	b.WriteString("\n")

	daysInMo := len(summaries)
	for i, s := range summaries {
		day := i + 1
		isToday := s.Date.Year() == today.Year() && s.Date.Month() == today.Month() && s.Date.Day() == today.Day()

		dayLabel := fmt.Sprintf("%2d", day)

		// Build indicator: [Cal][Gym][Mood]
		var indicator string
		if !s.HasData {
			indicator = ui.InfoStyle.Render(dayLabel + "  ·  ·  ·")
		} else {
			var calPart, gymPart, moodPart string

			// Cal indicator
			if s.CalConsumed > 0 {
				pct := float64(s.CalConsumed) / float64(s.CalGoal)
				if pct > 1 {
					calPart = ui.DangerStyle.Render(fmt.Sprintf("%4d", s.CalConsumed))
				} else if pct > 0.8 {
					calPart = ui.WarningStyle.Render(fmt.Sprintf("%4d", s.CalConsumed))
				} else {
					calPart = ui.SuccessStyle.Render(fmt.Sprintf("%4d", s.CalConsumed))
				}
			} else {
				calPart = ui.InfoStyle.Render("   ·")
			}

			// Gym indicator
			if s.GymSessions > 0 {
				gymPart = ui.SuccessStyle.Render(fmt.Sprintf("%3dm", s.GymDuration))
			} else {
				gymPart = ui.InfoStyle.Render("   ·")
			}

			// Mood indicator
			if s.MoodRating > 0 {
				ms := ui.MoodStyles[s.MoodRating-1]
				moodPart = ms.Render(ui.MoodEmoji(s.MoodRating))
			} else {
				moodPart = ui.InfoStyle.Render(" ·")
			}

			indicator = ui.ValueStyle.Render(dayLabel) + "  " + calPart + "  " + gymPart + "  " + moodPart
		}

		if isToday {
			indicator = ui.TodayBadgeStyle.Render("▸") + " " + indicator
		} else {
			indicator = "  " + indicator
		}

		b.WriteString(indicator + "\n")

		// Print a week separator after every Sunday
		wd := s.Date.Weekday()
		if wd == time.Sunday && day < daysInMo {
			b.WriteString(ui.InfoStyle.Render("  " + strings.Repeat("·", min(width-8, 30))) + "\n")
		}
	}

	// Legend
	b.WriteString("\n")
	b.WriteString(ui.InfoStyle.Render("  Day   Cal   Gym  Mood"))

	return b.String()
}

// ── Calendar view ──────────────────────────────────────

func (m Model) viewCalendar() string {
	today := time.Now()
	todayStr := today.Format("2006-01-02")

	// Title
	title := ui.CalendarTitleStyle.Render(
		fmt.Sprintf("◀  %s %d  ▶", m.calMonth.String(), m.calYear),
	)

	// Day-of-week header
	days := []string{"Mo", "Tu", "We", "Th", "Fr", "Sa", "Su"}
	var headerCells []string
	for _, d := range days {
		headerCells = append(headerCells, ui.CalendarHeaderStyle.Width(4).Align(lipgloss.Center).Render(d))
	}
	header := strings.Join(headerCells, "")

	// Build weeks
	firstDay := time.Date(m.calYear, m.calMonth, 1, 0, 0, 0, 0, time.Local)
	lastDay := daysInMonth(m.calYear, m.calMonth)
	weekday := int(firstDay.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	weekday-- // Monday=0

	var weeks []string
	var week []string

	// Leading blanks
	for i := 0; i < weekday; i++ {
		week = append(week, ui.CalendarMutedStyle.Render(""))
	}

	for day := 1; day <= lastDay; day++ {
		dateStr := fmt.Sprintf("%04d-%02d-%02d", m.calYear, int(m.calMonth), day)
		dayStr := fmt.Sprintf("%d", day)

		isSelected := day == m.calDay
		isToday := dateStr == todayStr
		hasData := m.calDayHasData(day)

		var cell string
		switch {
		case isSelected:
			cell = ui.CalendarSelectedStyle.Render(dayStr)
		case isToday:
			cell = ui.CalendarTodayStyle.Render(dayStr)
		case hasData:
			cell = ui.CalendarHasDataStyle.Render(dayStr)
		default:
			cell = ui.CalendarDayStyle.Render(dayStr)
		}
		week = append(week, cell)

		if len(week) == 7 {
			weeks = append(weeks, strings.Join(week, ""))
			week = nil
		}
	}
	if len(week) > 0 {
		for len(week) < 7 {
			week = append(week, ui.CalendarMutedStyle.Render(""))
		}
		weeks = append(weeks, strings.Join(week, ""))
	}

	grid := strings.Join(weeks, "\n")

	// Legend
	legend := "\n" +
		ui.CalendarSelectedStyle.Render("██") + " selected  " +
		ui.CalendarTodayStyle.Render("██") + " today  " +
		ui.CalendarHasDataStyle.Render("██") + " has data"

	// Controls
	controls := "\n\n" + ui.FooterStyle.Render(
		fmtKey("←/→", "day") + "  " +
			fmtKey("↑/↓", "week") + "  " +
			fmtKey("H/L", "month") + "  " +
			fmtKey("t", "today") + "  " +
			fmtKey("Enter", "select") + "  " +
			fmtKey("Esc", "close"),
	)

	calContent := title + "\n" + header + "\n" + grid + "\n" + legend + controls
	calBox := ui.CalendarBoxStyle.Render(calContent)

	// Center on screen
	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		calBox,
	)
}

func (m Model) calDayHasData(day int) bool {
	date := time.Date(m.calYear, m.calMonth, day, 0, 0, 0, 0, time.Local)
	dir := m.dataDirForDate(date)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	return len(entries) > 0
}

func (m Model) dataDirForDate(date time.Time) string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".atlas", "atlas.guide.data", date.Format("2006-01-02"))
}

// ── Help view ──────────────────────────────────────────

func (m Model) viewHelp() string {
	sections := []struct {
		title string
		rows  []table.Row
	}{
		{"Navigation", []table.Row{
			{"1-7, Tab", "Switch tabs (Overview/Monthly/Cal/Gym/Mood/Med/Journal)"},
			{"←/→, h/l", "Previous / Next day"},
			{"↑/↓, j/k", "Move cursor in lists"},
			{"t", "Jump to today"},
			{"c", "Open calendar picker"},
		}},
		{"Actions", []table.Row{
			{"a", "Add new entry"},
			{"d", "Delete selected / clear mood"},
			{"Enter", "Confirm input"},
			{"Esc", "Cancel input / close overlay"},
		}},
		{"Calories", []table.Row{
			{"g", "Set daily calorie goal"},
			{"f", "Toggle / set fasting window"},
		}},
		{"Medicine", []table.Row{
			{"a", "Add medicine (name + dosage)"},
			{"d", "Delete selected medicine"},
		}},
		{"Journal", []table.Row{
			{"a", "Write a journal note"},
			{"d", "Delete selected note"},
		}},
		{"Monthly", []table.Row{
			{"H/L", "Previous / Next month"},
		}},
		{"Calendar", []table.Row{
			{"←/→", "Move by day"},
			{"↑/↓", "Move by week"},
			{"H/L", "Previous / Next month"},
			{"t", "Jump to today"},
			{"Enter", "Select date"},
		}},
		{"General", []table.Row{
			{"?", "Toggle this help"},
			{"q, Ctrl+C", "Quit"},
		}},
	}

	cols := []table.Column{
		{Title: "Key", Width: 14},
		{Title: "Action", Width: 40},
	}

	tableStyle := table.DefaultStyles()
	tableStyle.Header = tableStyle.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(ui.ColorBorder).
		Foreground(ui.ColorPrimary).
		Bold(true)
	tableStyle.Selected = lipgloss.NewStyle()
	tableStyle.Cell = tableStyle.Cell.
		Foreground(ui.ColorText)

	var b strings.Builder
	b.WriteString(ui.HeaderBoxStyle.Render("atlas.guide — Help") + "\n\n")

	for _, s := range sections {
		b.WriteString(ui.TitleStyle.Render(s.title) + "\n")

		t := table.New(
			table.WithColumns(cols),
			table.WithRows(s.rows),
			table.WithHeight(len(s.rows)),
			table.WithStyles(tableStyle),
		)
		b.WriteString(t.View())
		b.WriteString("\n\n")
	}

	b.WriteString(ui.InfoStyle.Render("Press any key to close"))

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		ui.CalendarBoxStyle.Render(b.String()),
	)
}

// ── Helpers ────────────────────────────────────────────

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
