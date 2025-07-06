// File: tui/app.go
package tui

import (
	"fmt"
	"habit-tracker/model"
	"strconv"
	"strings"
	"time"

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

type modelState struct {
	selected      int
	today         int
	dates         []time.Time
	habits        []model.Habit
	tasks         []model.Task
	selectedHabit int
	selectedTask  int
	mode          string
	newHabitName  string
	newTaskName   string
}

func initialModel() modelState {
	today := time.Now()
	start := today.AddDate(0, 0, -int(today.Weekday()))
	week := make([]time.Time, 7)
	for i := 0; i < 7; i++ {
		week[i] = start.AddDate(0, 0, i)
	}
	habits, _ := model.GetHabits()
	tasks, _ := model.GetTasks()
	return modelState{
		today:         int(today.Weekday()),
		selected:      int(today.Weekday()),
		dates:         week,
		habits:        habits,
		tasks:         tasks,
		selectedHabit: 0,
		selectedTask:  0,
		mode:          "calendar",
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
			if m.mode == "calendar" && m.selected > 0 {
				m.selected--
			}
		case "right":
			if m.mode == "calendar" && m.selected < 6 {
				m.selected++
			}
		case "up":
			if m.mode == "habits" && m.selectedHabit > 0 {
				m.selectedHabit--
			} else if m.mode == "tasks" && m.selectedTask > 0 {
				m.selectedTask--
			}
		case "down":
			if m.mode == "habits" && m.selectedHabit < len(m.habits)-1 {
				m.selectedHabit++
			} else if m.mode == "tasks" && m.selectedTask < len(m.tasks)-1 {
				m.selectedTask++
			}
		case "tab":
			if m.mode == "calendar" {
				m.mode = "habits"
			} else if m.mode == "habits" {
				m.mode = "tasks"
			} else if m.mode == "tasks" {
				m.mode = "stats"
			} else if m.mode == "stats" {
				m.mode = "calendar"
			}
		case "enter":
			if m.mode == "adding" && strings.TrimSpace(m.newHabitName) != "" {
				habitID := strconv.FormatInt(time.Now().UnixNano(), 10)
				model.AddHabit(habitID, strings.TrimSpace(m.newHabitName))
				m.habits, _ = model.GetHabits()
				m.mode = "habits"
				m.newHabitName = ""
			} else if m.mode == "adding_task" && strings.TrimSpace(m.newTaskName) != "" {
				taskID := strconv.FormatInt(time.Now().UnixNano(), 10)
				model.AddTask(taskID, strings.TrimSpace(m.newTaskName), "", "")
				m.tasks, _ = model.GetTasks()
				m.mode = "tasks"
				m.newTaskName = ""
			}
		case " ":
			if m.mode == "habits" && len(m.habits) > 0 {
				dateStr := m.dates[m.selected].Format("2006-01-02")
				habitID := m.habits[m.selectedHabit].ID
				model.ToggleHabitCompletion(habitID, dateStr)
			} else if m.mode == "tasks" && len(m.tasks) > 0 {
				taskID := m.tasks[m.selectedTask].ID
				model.ToggleTask(taskID)
				m.tasks, _ = model.GetTasks()
			} else if m.mode == "adding" {
				m.newHabitName += " "
			} else if m.mode == "adding_task" {
				m.newTaskName += " "
			}
		case "a":
			if m.mode == "habits" {
				m.mode = "adding"
				m.newHabitName = ""
			} else if m.mode == "tasks" {
				m.mode = "adding_task"
				m.newTaskName = ""
			} else if m.mode == "adding" {
				m.newHabitName += "a"
			} else if m.mode == "adding_task" {
				m.newTaskName += "a"
			}
		case "d":
			if m.mode == "habits" && len(m.habits) > 0 {
				model.DeleteHabit(m.habits[m.selectedHabit].ID)
				m.habits, _ = model.GetHabits()
				if m.selectedHabit >= len(m.habits) && len(m.habits) > 0 {
					m.selectedHabit = len(m.habits) - 1
				}
			} else if m.mode == "tasks" && len(m.tasks) > 0 {
				model.DeleteTask(m.tasks[m.selectedTask].ID)
				m.tasks, _ = model.GetTasks()
				if m.selectedTask >= len(m.tasks) && len(m.tasks) > 0 {
					m.selectedTask = len(m.tasks) - 1
				}
			}
		case "esc":
			if m.mode == "adding" {
				m.mode = "habits"
				m.newHabitName = ""
			} else if m.mode == "adding_task" {
				m.mode = "tasks"
				m.newTaskName = ""
			}
		case "backspace":
			if m.mode == "adding" && len(m.newHabitName) > 0 {
				m.newHabitName = m.newHabitName[:len(m.newHabitName)-1]
			} else if m.mode == "adding_task" && len(m.newTaskName) > 0 {
				m.newTaskName = m.newTaskName[:len(m.newTaskName)-1]
			}
		case "q":
			if m.mode != "adding" && m.mode != "adding_task" {
				return m, tea.Quit
			}
		default:
			if m.mode == "adding" && len(msg.String()) == 1 {
				char := msg.String()
				if (char >= "a" && char <= "z") || (char >= "A" && char <= "Z") || (char >= "0" && char <= "9") || 
				   char == "-" || char == "_" || char == "." || char == "," || char == "'" || char == "\"" {
					m.newHabitName += char
				}
			} else if m.mode == "adding_task" && len(msg.String()) == 1 {
				char := msg.String()
				if (char >= "a" && char <= "z") || (char >= "A" && char <= "Z") || (char >= "0" && char <= "9") || 
				   char == "-" || char == "_" || char == "." || char == "," || char == "'" || char == "\"" {
					m.newTaskName += char
				}
			}
		}
	}
	return m, nil
}

func (m modelState) View() string {
	var dayCards []string
	for i, date := range m.dates {
		dateStr := date.Format("2006-01-02")
		
		completed := 0
		total := len(m.habits)
		for _, h := range m.habits {
			if done, _ := model.IsHabitCompleted(h.ID, dateStr); done {
				completed++
			}
		}
		
		var progressBar string
		if total > 0 {
			percentage := float64(completed) / float64(total)
			if percentage == 1.0 {
				progressBar = "●●●"
			} else if percentage >= 0.66 {
				progressBar = "●●○"
			} else if percentage >= 0.33 {
				progressBar = "●○○"
			} else if percentage > 0 {
				progressBar = "◐○○"
			} else {
				progressBar = "○○○"
			}
		} else {
			progressBar = "   "
		}
		
		label := date.Format("Mon\n02 Jan") + "\n" + progressBar
		style := dayStyle
		if i == m.today {
			style = highlightedDay
		}
		if i == m.selected && m.mode == "calendar" {
			style = selectedDay
		}
		dayCards = append(dayCards, style.Render(label))
	}

	weekRow := lipgloss.JoinHorizontal(lipgloss.Top, dayCards...)

	var contentBuilder strings.Builder
	
	if m.mode == "habits" || m.mode == "adding" {
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
				
				var style lipgloss.Style
				if m.mode == "habits" && i == m.selectedHabit {
					style = selectedHabitStyle
				} else if completed {
					style = completedHabitStyle
				} else {
					style = incompleteHabitStyle
				}
				
				contentBuilder.WriteString(style.Render(habitLine) + "\n")
			}
		}
	} else if m.mode == "tasks" || m.mode == "adding_task" {
		contentBuilder.WriteString("Tasks\n\n")
		
		if len(m.tasks) == 0 {
			contentBuilder.WriteString("No tasks yet. Press 'a' to add one.")
		} else {
			for i, t := range m.tasks {
				var taskLine string
				if t.Completed {
					taskLine = "✓ " + t.Name
				} else {
					taskLine = "○ " + t.Name
				}
				
				var style lipgloss.Style
				if m.mode == "tasks" && i == m.selectedTask {
					style = selectedHabitStyle
				} else if t.Completed {
					style = completedHabitStyle
				} else {
					style = incompleteHabitStyle
				}
				
				contentBuilder.WriteString(style.Render(taskLine) + "\n")
			}
		}
	} else if m.mode == "stats" {
		contentBuilder.WriteString("Habit Statistics\n\n")
		
		if len(m.habits) == 0 {
			contentBuilder.WriteString("No habits to show statistics for.")
		} else {
			for _, h := range m.habits {
				currentStreak, _ := model.GetHabitStreak(h.ID)
				longestStreak, _ := model.GetHabitLongestStreak(h.ID)
				
				statsLine := fmt.Sprintf("%s\n  Current: %d days | Best: %d days", 
					h.Name, currentStreak, longestStreak)
				
				var style lipgloss.Style
				if currentStreak >= 7 {
					style = completedHabitStyle
				} else if currentStreak > 0 {
					style = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Padding(0, 1)
				} else {
					style = incompleteHabitStyle
				}
				
				contentBuilder.WriteString(style.Render(statsLine) + "\n\n")
			}
		}
	}

	var controls string
	if m.mode == "calendar" {
		controls = "Calendar: ←/→ navigate | Tab: switch to habits | q: quit"
	} else if m.mode == "habits" {
		controls = "Habits: ↑/↓ navigate | Space: toggle | d: delete | a: add | Tab: tasks | q: quit"
	} else if m.mode == "tasks" {
		controls = "Tasks: ↑/↓ navigate | Space: toggle | d: delete | a: add | Tab: stats | q: quit"
	} else if m.mode == "stats" {
		controls = "Statistics: View habit streaks and progress | Tab: calendar | q: quit"
	} else if m.mode == "adding" {
		contentBuilder.WriteString("\n" + lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11")).Render("Add new habit:") + "\n")
		inputStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("12")).
			Padding(0, 1).
			Foreground(lipgloss.Color("15"))
		contentBuilder.WriteString(inputStyle.Render(m.newHabitName + "█") + "\n")
		controls = "Type habit name | Enter: save | Esc: cancel"
	} else if m.mode == "adding_task" {
		contentBuilder.WriteString("\n" + lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11")).Render("Add new task:") + "\n")
		inputStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("12")).
			Padding(0, 1).
			Foreground(lipgloss.Color("15"))
		contentBuilder.WriteString(inputStyle.Render(m.newTaskName + "█") + "\n")
		controls = "Type task name | Enter: save | Esc: cancel"
	}

	contentSection := habitSectionStyle.Render(contentBuilder.String())
	styledControls := controlsStyle.Render(controls)

	return weekRow + contentSection + styledControls
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
