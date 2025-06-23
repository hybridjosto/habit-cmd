// File: tui/app.go
package tui

import (
	"fmt"
	"habit-tracker/model"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	dayStyle       = lipgloss.NewStyle().Padding(1, 2).Border(lipgloss.RoundedBorder())
	highlightedDay = dayStyle.BorderForeground(lipgloss.Color("2"))
	selectedDay    = highlightedDay.Bold(true).Underline(true)
)

type modelState struct {
	selected int
	today    int
	dates    []time.Time
	habits   []model.Habit
}

func initialModel() modelState {
	today := time.Now()
	start := today.AddDate(0, 0, -int(today.Weekday()))
	week := make([]time.Time, 7)
	for i := 0; i < 7; i++ {
		week[i] = start.AddDate(0, 0, i)
	}
	habits, _ := model.GetHabits()
	return modelState{
		today:    int(today.Weekday()),
		selected: int(today.Weekday()),
		dates:    week,
		habits:   habits,
	}
}

func (m modelState) Init() tea.Cmd {
	return nil
}

func (m modelState) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left":
			if m.selected > 0 {
				m.selected--
			}
		case "right":
			if m.selected < 6 {
				m.selected++
			}
		case "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m modelState) View() string {
	var dayCards []string
	for i, date := range m.dates {
		label := date.Format("Mon\n02 Jan")
		style := dayStyle
		if i == m.today {
			style = highlightedDay
		}
		if i == m.selected {
			style = selectedDay
		}
		dayCards = append(dayCards, style.Render(label))
	}

	weekRow := lipgloss.JoinHorizontal(lipgloss.Top, dayCards...)

	var habitsSection strings.Builder
	habitsSection.WriteString("\n\nHabits:\n")
	for _, h := range m.habits {
		habitsSection.WriteString(" - " + h.Name + "\n")
	}

	return weekRow + habitsSection.String() + "\n←/→ to move, q to quit"
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
