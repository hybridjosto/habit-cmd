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
