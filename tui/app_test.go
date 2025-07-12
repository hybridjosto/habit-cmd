package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"habit-tracker/model"
	"os"
	"testing"
)

func TestInitialModel(t *testing.T) {
	dbPath := "test.db"
	if err := model.InitDB(dbPath); err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	defer func() {
		model.CloseDB()
		os.Remove(dbPath)
	}()

	m := initialModel()
	if len(m.dates) != 7 {
		t.Fatalf("expected 7 days, got %d", len(m.dates))
	}
	if m.mode != "calendar" {
		t.Fatalf("expected mode calendar, got %s", m.mode)
	}
}

func TestTabCyclesModes(t *testing.T) {
	dbPath := "test.db"
	if err := model.InitDB(dbPath); err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	defer func() {
		model.CloseDB()
		os.Remove(dbPath)
	}()

	m := initialModel()
	km := tea.KeyMsg{Type: tea.KeyTab}
	next, _ := m.Update(km)
	m2, ok := next.(modelState)
	if !ok {
		t.Fatalf("model type mismatch")
	}
	if m2.mode != "habits" {
		t.Fatalf("expected habits, got %s", m2.mode)
	}
	next2, _ := m2.Update(km)
	m3, ok := next2.(modelState)
	if !ok {
		t.Fatalf("model type mismatch")
	}
	if m3.mode != "tasks" {
		t.Fatalf("expected tasks, got %s", m3.mode)
	}
}

func TestAddingTaskInput(t *testing.T) {
	dbPath := "test.db"
	if err := model.InitDB(dbPath); err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	defer func() {
		model.CloseDB()
		os.Remove(dbPath)
	}()

	m := initialModel()
	m.mode = "adding_task"

	letters := []rune{'d', 'n', 'e', 'q', 'x'}
	for _, r := range letters {
		km := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		next, _ := m.Update(km)
		m = next.(modelState)
	}

	expected := "dneqx"
	if m.newTaskName != expected {
		t.Fatalf("expected %s, got %s", expected, m.newTaskName)
	}
}

func TestArrowKeysMoveDaysInAllModes(t *testing.T) {
	dbPath := "test.db"
	if err := model.InitDB(dbPath); err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	defer func() {
		model.CloseDB()
		os.Remove(dbPath)
	}()

	m := initialModel()
	m.mode = "habits"
	m.selected = 3

	left := tea.KeyMsg{Type: tea.KeyLeft}
	next, _ := m.Update(left)
	m2 := next.(modelState)
	if m2.selected != 2 {
		t.Fatalf("expected selected 2, got %d", m2.selected)
	}

	right := tea.KeyMsg{Type: tea.KeyRight}
	next2, _ := m2.Update(right)
	m3 := next2.(modelState)
	if m3.selected != 3 {
		t.Fatalf("expected selected 3, got %d", m3.selected)
	}
}

func TestAddingHabit(t *testing.T) {
	dbPath := "test.db"
	if err := model.InitDB(dbPath); err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	defer func() {
		model.CloseDB()
		os.Remove(dbPath)
	}()

	m := initialModel()
	m.mode = "adding"
	m.newHabitName = "new habit"

	enter := tea.KeyMsg{Type: tea.KeyEnter}
	next, _ := m.Update(enter)
	m2 := next.(modelState)

	if m2.mode != "habits" {
		t.Fatalf("expected mode 'habits', got '%s'", m2.mode)
	}
	if len(m2.habits) != 1 {
		t.Fatalf("expected 1 habit, got %d", len(m2.habits))
	}
	if m2.habits[0].Name != "new habit" {
		t.Fatalf("expected habit name 'new habit', got '%s'", m2.habits[0].Name)
	}
}

func TestNoteEditing(t *testing.T) {
	dbPath := "test.db"
	if err := model.InitDB(dbPath); err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	defer func() {
		model.CloseDB()
		os.Remove(dbPath)
	}()

	model.AddHabit("1", "test habit", "general", nil)
	m := initialModel()
	m.mode = "habits"
	m.selectedHabit = 0
	m.showNotes = true
	m.editingNote = true
	m.editedNote = "initial note"

	// Simulate typing
	km := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	next, _ := m.Update(km)
	m = next.(modelState)
	km = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}}
	next, _ = m.Update(km)
	m = next.(modelState)
	km = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}}
	next, _ = m.Update(km)
	m = next.(modelState)
	if m.editedNote != "initial noteabc" {
		t.Fatalf("expected note 'initial noteabc', got '%s'", m.editedNote)
	}

	// Simulate saving
	enter := tea.KeyMsg{Type: tea.KeyEnter}
	next, _ = m.Update(enter)
	m = next.(modelState)

	if m.editingNote != false {
		t.Fatalf("expected editingNote to be false")
	}
	habits, _ := model.GetHabits()
	if habits[0].Notes["general"] != "initial noteabc" {
		t.Fatalf("expected note to be saved")
	}
}