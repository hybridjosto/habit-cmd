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
   "github.com/charmbracelet/bubbles/key"
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
// keybindings for application commands
var keys = struct {
   Left, Right, Up, Down, Tab, Enter, Escape, Backspace, Space,
   N, E, A, D, Q key.Binding
}{
   Left:      key.NewBinding(key.WithKeys("left"),      key.WithHelp("←",     "prev day")),
   Right:     key.NewBinding(key.WithKeys("right"),     key.WithHelp("→",     "next day")),
   Up:        key.NewBinding(key.WithKeys("up"),        key.WithHelp("↑",     "move up")),
   Down:      key.NewBinding(key.WithKeys("down"),      key.WithHelp("↓",     "move down")),
   Tab:       key.NewBinding(key.WithKeys("tab"),       key.WithHelp("⇥",     "switch section")),
   Enter:     key.NewBinding(key.WithKeys("enter"),     key.WithHelp("⏎",     "confirm")),
   Escape:    key.NewBinding(key.WithKeys("esc"),       key.WithHelp("esc",   "cancel")),
   Backspace: key.NewBinding(key.WithKeys("backspace"), key.WithHelp("⌫",     "delete char")),
   Space:     key.NewBinding(key.WithKeys(" "),         key.WithHelp("space", "toggle")),
   N:         key.NewBinding(key.WithKeys("n"),         key.WithHelp("n",     "toggle notes")),
   E:         key.NewBinding(key.WithKeys("e"),         key.WithHelp("e",     "edit note")),
   A:         key.NewBinding(key.WithKeys("a"),         key.WithHelp("a",     "add")),
   D:         key.NewBinding(key.WithKeys("d"),         key.WithHelp("d",     "delete")),
   Q:         key.NewBinding(key.WithKeys("q"),         key.WithHelp("q",     "quit")),
}

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
	showNotes     bool
	editingNote   bool
	editedNote    string
	newHabitType  string
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
        if m.editingNote {
            // handle note editing keybindings
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
                    // accumulate character input
                    if len(km.String()) == 1 {
                        m.editedNote += km.String()
                    }
                }
            }
            return m, nil
        }

       switch msg := msg.(type) {
       case tea.KeyMsg:
           // general keybindings via bubbletea key.Matches
           switch {
           case key.Matches(msg, keys.Left):
               if m.mode == "calendar" && m.selected > 0 {
                   m.selected--
               }
           case key.Matches(msg, keys.Right):
               if m.mode == "calendar" && m.selected < 6 {
                   m.selected++
               }
           case key.Matches(msg, keys.Up):
               if m.mode == "choosing_habit_type" {
                   m.newHabitType = "general"
               } else if m.mode == "habits" && m.selectedHabit > 0 {
                   m.selectedHabit--
               } else if m.mode == "tasks" && m.selectedTask > 0 {
                   m.selectedTask--
               }
           case key.Matches(msg, keys.Down):
               if m.mode == "choosing_habit_type" {
                   m.newHabitType = "daily"
               } else if m.mode == "habits" && m.selectedHabit < len(m.habits)-1 {
                   m.selectedHabit++
               } else if m.mode == "tasks" && m.selectedTask < len(m.tasks)-1 {
                   m.selectedTask++
               }
           case key.Matches(msg, keys.Tab):
               if m.mode == "calendar" {
                   m.mode = "habits"
               } else if m.mode == "habits" {
                   m.mode = "tasks"
               } else if m.mode == "tasks" {
                   m.mode = "stats"
               } else if m.mode == "stats" {
                   m.mode = "calendar"
               }
           case key.Matches(msg, keys.Enter):
               if m.mode == "choosing_habit_type" {
                   m.mode = "adding"
                   m.newHabitName = ""
               } else if m.mode == "adding" && strings.TrimSpace(m.newHabitName) != "" {
                   habitID := strconv.FormatInt(time.Now().UnixNano(), 10)
                   model.AddHabit(habitID, strings.TrimSpace(m.newHabitName), m.newHabitType, make(map[string]string))
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
           case key.Matches(msg, keys.N):
               if m.mode == "habits" && len(m.habits) > 0 {
                   m.showNotes = !m.showNotes
               }
           case key.Matches(msg, keys.E):
               if m.showNotes && !m.editingNote {
                   m.editingNote = true
                   habit := m.habits[m.selectedHabit]
                   if habit.Notes == nil {
                       habit.Notes = make(map[string]string)
                   }
                   if habit.Type == "general" {
                       m.editedNote = habit.Notes["general"]
                   } else {
                       day := m.dates[m.selected].Weekday().String()
                       m.editedNote = habit.Notes[day]
                   }
               }
           case key.Matches(msg, keys.Space):
               if m.mode == "habits" && len(m.habits) > 0 {
                   dateStr := m.dates[m.selected].Format("2006-01-02")
                   model.ToggleHabitCompletion(m.habits[m.selectedHabit].ID, dateStr)
               } else if m.mode == "tasks" && len(m.tasks) > 0 {
                   model.ToggleTask(m.tasks[m.selectedTask].ID)
                   m.tasks, _ = model.GetTasks()
               } else if m.mode == "adding" {
                   m.newHabitName += " "
               } else if m.mode == "adding_task" {
                   m.newTaskName += " "
               }
           case key.Matches(msg, keys.A):
               if m.mode == "habits" {
                   m.mode = "choosing_habit_type"
                   m.newHabitType = "general"
               } else if m.mode == "tasks" {
                   m.mode = "adding_task"
                   m.newTaskName = ""
               } else if m.mode == "adding" {
                   m.newHabitName += "a"
               } else if m.mode == "adding_task" {
                   m.newTaskName += "a"
               }
           case key.Matches(msg, keys.D):
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
           case key.Matches(msg, keys.Escape):
               if m.showNotes {
                   m.showNotes = false
               } else if m.mode == "adding" {
                   m.mode = "habits"
                   m.newHabitName = ""
               } else if m.mode == "adding_task" {
                   m.mode = "tasks"
                   m.newTaskName = ""
               }
           case key.Matches(msg, keys.Backspace):
               if m.mode == "adding" && len(m.newHabitName) > 0 {
                   m.newHabitName = m.newHabitName[:len(m.newHabitName)-1]
               } else if m.mode == "adding_task" && len(m.newTaskName) > 0 {
                   m.newTaskName = m.newTaskName[:len(m.newTaskName)-1]
               }
           case key.Matches(msg, keys.Q):
               if m.mode != "adding" && m.mode != "adding_task" {
                   return m, tea.Quit
               }
           default:
               // default character input for adding modes
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
       };


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

	var popup lipgloss.Style
	var contentBuilder strings.Builder

	if m.showNotes {
		habit := m.habits[m.selectedHabit]
		var note string
		if habit.Type == "general" {
			note = habit.Notes["general"]
		} else {
			day := m.dates[m.selected].Weekday().String()
			note = habit.Notes[day]
		}

		if m.editingNote {
			contentBuilder.WriteString("Editing note:\n")
			contentBuilder.WriteString(m.editedNote)
		} else {
			contentBuilder.WriteString("Note:\n")
			contentBuilder.WriteString(note)
		}

		popup = lipgloss.NewStyle().
			SetString(contentBuilder.String()).
			Border(lipgloss.RoundedBorder()).
			Padding(1, 2)

	}
	
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

	return weekRow + contentSection + styledControls + popup.String()
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
