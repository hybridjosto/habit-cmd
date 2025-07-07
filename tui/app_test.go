package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"testing"
)

func TestInitialModel(t *testing.T) {
	m := initialModel()
	if len(m.dates) != 7 {
		t.Fatalf("expected 7 days, got %d", len(m.dates))
	}
	if m.mode != "calendar" {
		t.Fatalf("expected mode calendar, got %s", m.mode)
	}
}

func TestTabCyclesModes(t *testing.T) {
	m := initialModel()
	km := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'\t'}}
	m2, _ := m.Update(km).(modelState)
	if m2.mode != "habits" {
		t.Fatalf("expected habits, got %s", m2.mode)
	}
	m3, _ := m2.Update(km).(modelState)
	if m3.mode != "tasks" {
		t.Fatalf("expected tasks, got %s", m3.mode)
	}
}
