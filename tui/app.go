// File: tui/app.go
package tui

import (
	"fmt"
	"habit-tracker/model"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	dayStyle       = lipgloss.NewStyle().Padding(1, 2).Border(lipgloss.RoundedBorder())
	highlightedDay = dayStyle.BorderForeground(lipgloss.Color("2"))
	selectedDay    = dayStyle.Bold(true).Underline(true)

	completedHabitStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("10")).
				Background(lipgloss.Color("22")).
				Padding(0, 1).
				Bold(true)

	incompleteHabitStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("8")).
				Background(lipgloss.Color("0")).
				Padding(0, 1)

	selectedHabitStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("4")).
				Padding(0, 1).
				Bold(true)

	habitSectionStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("6")).
				Padding(1).
				MarginTop(1)

	controlsStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("7")).
			Italic(true).
			MarginTop(1)
)

// keybindings for application commands
var keys = struct {
	Left, Right, Up, Down, Tab, Enter, Escape, Backspace, Space,
	N, E, A, D, C, U, V, Quit key.Binding
}{
	Left:      key.NewBinding(key.WithKeys("left"), key.WithHelp("←", "prev")),
	Right:     key.NewBinding(key.WithKeys("right"), key.WithHelp("→", "next")),
	Up:        key.NewBinding(key.WithKeys("up"), key.WithHelp("↑", "move up")),
	Down:      key.NewBinding(key.WithKeys("down"), key.WithHelp("↓", "move down")),
	Tab:       key.NewBinding(key.WithKeys("tab"), key.WithHelp("⇥", "switch section")),
	Enter:     key.NewBinding(key.WithKeys("enter"), key.WithHelp("⏎", "confirm")),
	Escape:    key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel/back")),
	Backspace: key.NewBinding(key.WithKeys("backspace"), key.WithHelp("⌫", "delete char")),
	Space:     key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "toggle")),
	N:         key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "toggle notes")),
	E:         key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
	A:         key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add")),
	D:         key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "archive")),
	C:         key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "calendar view")),
	U:         key.NewBinding(key.WithKeys("u"), key.WithHelp("u", "unarchive")),
	V:         key.NewBinding(key.WithKeys("v"), key.WithHelp("v", "view archived")),
	Quit:      key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit")),
}

type modelState struct {
	selected            int
	today               int
	dates               []time.Time
	habits              []model.Habit
	archivedHabits      []model.Habit
	tasks               []model.Task
	selectedHabit       int
	selectedTask        int
	selectedArchived    int
	mode                string // week, habits, tasks, stats, archived, adding_habit, editing_habit, calendar
	newHabitName        string
	newHabitDescription string
	newTaskName         string
	showNotes           bool
	editingNote         bool
	editedNote          string
	newHabitType        string
	calendarMonth       time.Time
	editingField        string // "name" or "description"
}

func initialModel() modelState {
	today := time.Now()
	start := today.AddDate(0, 0, -int(today.Weekday()))
	week := make([]time.Time, 7)
	for i := 0; i < 7; i++ {
		week[i] = start.AddDate(0, 0, i)
	}
	habits, _ := model.GetHabits()
	archivedHabits, _ := model.GetArchivedHabits()
	tasks, _ := model.GetTasks()
	return modelState{
		today:          int(today.Weekday()),
		selected:       int(today.Weekday()),
		dates:          week,
		habits:         habits,
		archivedHabits: archivedHabits,
		tasks:          tasks,
		selectedHabit:  0,
		selectedTask:   0,
		mode:           "week",
		calendarMonth:  today,
	}
}

func (m modelState) Init() tea.Cmd {
	return nil
}

func (m modelState) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.editingNote {
		if km, ok := msg.(tea.KeyMsg); ok {
			switch {
			case key.Matches(km, keys.Enter):
				habit := m.habits[m.selectedHabit]
				if habit.Notes == nil {
					habit.Notes = make(map[string]string)
				}
				if habit.Type == "general" {
					habit.Notes["general"] = m.editedNote
				} else {
					day := m.dates[m.selected].Weekday().String()
					habit.Notes[day] = m.editedNote
				}
				model.UpdateHabit(habit.ID, habit)
				m.habits, _ = model.GetHabits()
				m.editingNote = false
			case key.Matches(km, keys.Escape):
				m.editingNote = false
			case key.Matches(km, keys.Backspace):
				if len(m.editedNote) > 0 {
					m.editedNote = m.editedNote[:len(m.editedNote)-1]
				}
			default:
				if len(km.String()) == 1 {
					m.editedNote += km.String()
				}
			}
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.mode == "adding_habit" || m.mode == "editing_habit" {
			switch {
			case key.Matches(msg, keys.Enter):
				if m.editingField == "name" {
					m.editingField = "description"
				} else {
					if m.mode == "adding_habit" {
						habitID := strconv.FormatInt(time.Now().UnixNano(), 10)
						model.AddHabit(habitID, m.newHabitName, m.newHabitDescription, m.newHabitType, make(map[string]string))
					} else {
						habit := m.habits[m.selectedHabit]
						habit.Name = m.newHabitName
						habit.Description = m.newHabitDescription
						model.UpdateHabit(habit.ID, habit)
					}
					m.habits, _ = model.GetHabits()
					m.mode = "habits"
					m.newHabitName = ""
					m.newHabitDescription = ""
				}
			case key.Matches(msg, keys.Escape):
				m.mode = "habits"
				m.newHabitName = ""
				m.newHabitDescription = ""
			case key.Matches(msg, keys.Backspace):
				if m.editingField == "name" {
					if len(m.newHabitName) > 0 {
						m.newHabitName = m.newHabitName[:len(m.newHabitName)-1]
					}
				} else {
					if len(m.newHabitDescription) > 0 {
						m.newHabitDescription = m.newHabitDescription[:len(m.newHabitDescription)-1]
					}
				}
			default:
				if msg.Type == tea.KeyRunes {
					if m.editingField == "name" {
						m.newHabitName += msg.String()
					} else {
						m.newHabitDescription += msg.String()
					}
				}
			}
			return m, nil
		}

		if m.mode == "adding_task" {
			// ... (existing task adding logic)
		}

		switch {
		case key.Matches(msg, keys.Left):
			if m.mode == "week" && m.selected > 0 {
				m.selected--
			} else if m.mode == "calendar" {
				m.calendarMonth = m.calendarMonth.AddDate(0, -1, 0)
			}
		case key.Matches(msg, keys.Right):
			if m.mode == "week" && m.selected < len(m.dates)-1 {
				m.selected++
			} else if m.mode == "calendar" {
				m.calendarMonth = m.calendarMonth.AddDate(0, 1, 0)
			}
		case key.Matches(msg, keys.Up):
			if m.mode == "habits" && m.selectedHabit > 0 {
				m.selectedHabit--
			} else if m.mode == "tasks" && m.selectedTask > 0 {
				m.selectedTask--
			} else if m.mode == "archived" && m.selectedArchived > 0 {
				m.selectedArchived--
			}
		case key.Matches(msg, keys.Down):
			if m.mode == "habits" && m.selectedHabit < len(m.habits)-1 {
				m.selectedHabit++
			} else if m.mode == "tasks" && m.selectedTask < len(m.tasks)-1 {
				m.selectedTask++
			} else if m.mode == "archived" && m.selectedArchived < len(m.archivedHabits)-1 {
				m.selectedArchived++
			}
		case key.Matches(msg, keys.Tab):
			modes := []string{"week", "habits", "tasks", "stats", "archived"}
			currentModeIndex := -1
			for i, mode := range modes {
				if m.mode == mode {
					currentModeIndex = i
					break
				}
			}
			if currentModeIndex != -1 {
				m.mode = modes[(currentModeIndex+1)%len(modes)]
			}
		case key.Matches(msg, keys.Enter):
			if m.mode == "choosing_habit_type" {
				m.mode = "adding_habit"
				m.editingField = "name"
			}
		case key.Matches(msg, keys.N):
			if m.mode == "habits" && len(m.habits) > 0 {
				m.showNotes = !m.showNotes
			}
		case key.Matches(msg, keys.E):
			if m.mode == "habits" && len(m.habits) > 0 {
				m.mode = "editing_habit"
				habit := m.habits[m.selectedHabit]
				m.newHabitName = habit.Name
				m.newHabitDescription = habit.Description
				m.editingField = "name"
			}
		case key.Matches(msg, keys.Space):
			if m.mode == "habits" && len(m.habits) > 0 {
				dateStr := m.dates[m.selected].Format("2006-01-02")
				model.ToggleHabitCompletion(m.habits[m.selectedHabit].ID, dateStr)
			} else if m.mode == "tasks" && len(m.tasks) > 0 {
				model.ToggleTask(m.tasks[m.selectedTask].ID)
				m.tasks, _ = model.GetTasks()
			}
		case key.Matches(msg, keys.A):
			if m.mode == "habits" {
				m.mode = "choosing_habit_type"
				m.newHabitType = "general"
			} else if m.mode == "tasks" {
				m.mode = "adding_task"
				m.newTaskName = ""
			}
		case key.Matches(msg, keys.D):
			if m.mode == "habits" && len(m.habits) > 0 {
				model.ArchiveHabit(m.habits[m.selectedHabit].ID)
				m.habits, _ = model.GetHabits()
				m.archivedHabits, _ = model.GetArchivedHabits()
				if m.selectedHabit >= len(m.habits) && len(m.habits) > 0 {
					m.selectedHabit = len(m.habits) - 1
				}
			}
		case key.Matches(msg, keys.C):
			if m.mode == "habits" && len(m.habits) > 0 {
				m.mode = "calendar"
			}
		case key.Matches(msg, keys.U):
			if m.mode == "archived" && len(m.archivedHabits) > 0 {
				model.UnarchiveHabit(m.archivedHabits[m.selectedArchived].ID)
				m.habits, _ = model.GetHabits()
				m.archivedHabits, _ = model.GetArchivedHabits()
				if m.selectedArchived >= len(m.archivedHabits) && len(m.archivedHabits) > 0 {
					m.selectedArchived = len(m.archivedHabits) - 1
				}
			}
		case key.Matches(msg, keys.V):
			m.mode = "archived"
		case key.Matches(msg, keys.Escape):
			if m.mode == "calendar" || m.mode == "archived" {
				m.mode = "habits"
			} else if m.showNotes {
				m.showNotes = false
			}
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m modelState) View() string {
	var s strings.Builder

	// Week View
	if m.mode == "week" {
		var dayCards []string
		for range m.dates {
			// ... (existing week view logic)
		}
		s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, dayCards...))
	}

	// Main Content
	var contentBuilder strings.Builder
	switch m.mode {
	case "habits", "adding_habit", "editing_habit":
		contentBuilder.WriteString("Habits for " + m.dates[m.selected].Format("Mon Jan 02") + "\n\n")
		if len(m.habits) == 0 {
			contentBuilder.WriteString("No habits yet. Press 'a' to add one.")
		} else {
			dateStr := m.dates[m.selected].Format("2006-01-02")
			for i, h := range m.habits {
				completed, _ := model.IsHabitCompleted(h.ID, dateStr)
				var habitLine string
				if completed {
					habitLine = "✓ " + h.Name
				} else {
					habitLine = "○ " + h.Name
				}
				style := incompleteHabitStyle
				if m.mode == "habits" && i == m.selectedHabit {
					style = selectedHabitStyle
				} else if completed {
					style = completedHabitStyle
				}
				contentBuilder.WriteString(style.Render(habitLine) + "\n")
				if i == m.selectedHabit {
					contentBuilder.WriteString("  " + h.Description + "\n")
				}
			}
		}
	case "tasks", "adding_task":
		// ... (existing task view logic)
	case "stats":
		contentBuilder.WriteString("Habit Statistics\n\n")
		if len(m.habits) == 0 {
			contentBuilder.WriteString("No habits to show statistics for.")
		} else {
			for _, h := range m.habits {
				currentStreak, _ := model.GetHabitStreak(h.ID)
				longestStreak, _ := model.GetHabitLongestStreak(h.ID)
				statsLine := fmt.Sprintf("%s\n  Current: %d days | Best: %d days", h.Name, currentStreak, longestStreak)
				contentBuilder.WriteString(statsLine + "\n\n")
			}
		}
	case "calendar":
		habit := m.habits[m.selectedHabit]
		contentBuilder.WriteString(fmt.Sprintf("Calendar for: %s (%s)\n", habit.Name, m.calendarMonth.Format("January 2006")))
		contentBuilder.WriteString(renderCalendar(m.calendarMonth, habit.ID))
	case "archived":
		contentBuilder.WriteString("Archived Habits\n\n")
		if len(m.archivedHabits) == 0 {
			contentBuilder.WriteString("No archived habits.")
		} else {
			for i, h := range m.archivedHabits {
				style := incompleteHabitStyle
				if i == m.selectedArchived {
					style = selectedHabitStyle
				}
				contentBuilder.WriteString(style.Render(h.Name) + "\n")
			}
		}
	}

	s.WriteString(habitSectionStyle.Render(contentBuilder.String()))

	// Controls
	// ... (add controls for new modes)

	// Popups
	if m.mode == "adding_habit" || m.mode == "editing_habit" {
		var popupBuilder strings.Builder
		title := "Add New Habit"
		if m.mode == "editing_habit" {
			title = "Edit Habit"
		}
		popupBuilder.WriteString(title + "\n")
		name_field := m.newHabitName
		desc_field := m.newHabitDescription
		if m.editingField == "name" {
			name_field += "█"
		} else {
			desc_field += "█"
		}
		popupBuilder.WriteString("Name: " + name_field + "\n")
		popupBuilder.WriteString("Description: " + desc_field + "\n")
		s.WriteString(lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1).Render(popupBuilder.String()))
	}

	return s.String()
}

func renderCalendar(month time.Time, habitID string) string {
	var cal strings.Builder
	startOfMonth := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, -1)
	startDay := int(startOfMonth.Weekday())

	cal.WriteString(" Su Mo Tu We Th Fr Sa\n")
	cal.WriteString(strings.Repeat("   ", startDay))

	for day := 1; day <= endOfMonth.Day(); day++ {
		date := time.Date(month.Year(), month.Month(), day, 0, 0, 0, 0, time.UTC)
		dateStr := date.Format("2006-01-02")
		completed, _ := model.IsHabitCompleted(habitID, dateStr)
		dayStr := " "
		if completed {
			dayStr = "✓"
		}
		cal.WriteString(fmt.Sprintf(" %s ", dayStr))
		if date.Weekday() == time.Saturday {
			cal.WriteString("\n")
		}
	}
	cal.WriteString("\n")
	return cal.String()
}

func StartApp() {
	err := model.InitDB("tracker.db")
	if err != nil {
		fmt.Println("Failed to open DB:", err)
		return
	}
	defer model.CloseDB()

	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
	}
}
