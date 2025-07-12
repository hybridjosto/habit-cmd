package model

import (
	"os"
	"testing"
	"time"
)

func setupTestDB(t *testing.T) func() {
	f, err := os.CreateTemp("", "tracker_test_*.db")
	if err != nil {
		t.Fatalf("create temp db: %v", err)
	}
	f.Close()
	if err := InitDB(f.Name()); err != nil {
		os.Remove(f.Name())
		t.Fatalf("init db: %v", err)
	}
	return func() {
		CloseDB()
		os.Remove(f.Name())
	}
}

func TestAddAndGetHabits(t *testing.T) {
	teardown := setupTestDB(t)
	defer teardown()

	if err := AddHabit("1", "drink water", "general", nil); err != nil {
		t.Fatalf("AddHabit err: %v", err)
	}

	habits, err := GetHabits()
	if err != nil {
		t.Fatalf("GetHabits err: %v", err)
	}
	if len(habits) != 1 || habits[0].Name != "drink water" {
		t.Fatalf("unexpected habits: %+v", habits)
	}
}

func TestToggleHabitCompletion(t *testing.T) {
	teardown := setupTestDB(t)
	defer teardown()

	AddHabit("1", "exercise", "general", nil)
	date := time.Now().Format("2006-01-02")
	if err := ToggleHabitCompletion("1", date); err != nil {
		t.Fatalf("toggle: %v", err)
	}
	done, err := IsHabitCompleted("1", date)
	if err != nil || !done {
		t.Fatalf("completion not recorded")
	}
}

func TestDeleteHabit(t *testing.T) {
	teardown := setupTestDB(t)
	defer teardown()

	AddHabit("1", "read", "general", nil)
	if err := DeleteHabit("1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	habits, _ := GetHabits()
	if len(habits) != 0 {
		t.Fatalf("habit not deleted")
	}
}

func TestAddAndToggleTask(t *testing.T) {
	teardown := setupTestDB(t)
	defer teardown()

	if err := AddTask("1", "task", "", ""); err != nil {
		t.Fatalf("add task: %v", err)
	}
	if err := ToggleTask("1"); err != nil {
		t.Fatalf("toggle task: %v", err)
	}
	tasks, err := GetTasks()
	if err != nil {
		t.Fatalf("get tasks: %v", err)
	}
	if len(tasks) != 1 || !tasks[0].Completed {
		t.Fatalf("task not toggled")
	}
}

func TestDeleteTask(t *testing.T) {
	teardown := setupTestDB(t)
	defer teardown()

	AddTask("1", "task", "", "")
	if err := DeleteTask("1"); err != nil {
		t.Fatalf("delete task: %v", err)
	}
	tasks, _ := GetTasks()
	if len(tasks) != 0 {
		t.Fatalf("task not deleted")
	}
}

func TestUpdateHabit(t *testing.T) {
	teardown := setupTestDB(t)
	defer teardown()

	AddHabit("1", "meditate", "daily", nil)
	updatedHabit := Habit{
		ID:    "1",
		Name:  "meditate daily",
		Type:  "daily",
		Notes: map[string]string{"Monday": "10 minutes"},
	}
	if err := UpdateHabit("1", updatedHabit); err != nil {
		t.Fatalf("update habit: %v", err)
	}
	habits, _ := GetHabits()
	if len(habits) != 1 || habits[0].Name != "meditate daily" || habits[0].Notes["Monday"] != "10 minutes" {
		t.Fatalf("habit not updated: %+v", habits[0])
	}
}

func TestGetHabitStreak(t *testing.T) {
	teardown := setupTestDB(t)
	defer teardown()

	AddHabit("1", "walk", "general", nil)
	today := time.Now()
	ToggleHabitCompletion("1", today.Format("2006-01-02"))
	ToggleHabitCompletion("1", today.AddDate(0, 0, -1).Format("2006-01-02"))
	ToggleHabitCompletion("1", today.AddDate(0, 0, -2).Format("2006-01-02"))

	streak, err := GetHabitStreak("1")
	if err != nil {
		t.Fatalf("get streak: %v", err)
	}
	if streak != 3 {
		t.Fatalf("expected streak 3, got %d", streak)
	}

	// test with a break in the streak
	AddHabit("2", "run", "general", nil)
	ToggleHabitCompletion("2", today.Format("2006-01-02"))
	ToggleHabitCompletion("2", today.AddDate(0, 0, -2).Format("2006-01-02"))
	streak, _ = GetHabitStreak("2")
	if streak != 1 {
		t.Fatalf("expected streak 1 after a break, got %d", streak)
	}
}

func TestGetHabitLongestStreak(t *testing.T) {
	teardown := setupTestDB(t)
	defer teardown()

	AddHabit("1", "code", "general", nil)
	today := time.Now()

	// 5 day streak
	ToggleHabitCompletion("1", today.Format("2006-01-02"))
	ToggleHabitCompletion("1", today.AddDate(0, 0, -1).Format("2006-01-02"))
	ToggleHabitCompletion("1", today.AddDate(0, 0, -2).Format("2006-01-02"))
	ToggleHabitCompletion("1", today.AddDate(0, 0, -3).Format("2006-01-02"))
	ToggleHabitCompletion("1", today.AddDate(0, 0, -4).Format("2006-01-02"))

	// 3 day streak
	ToggleHabitCompletion("1", today.AddDate(0, 0, -6).Format("2006-01-02"))
	ToggleHabitCompletion("1", today.AddDate(0, 0, -7).Format("2006-01-02"))
	ToggleHabitCompletion("1", today.AddDate(0, 0, -8).Format("2006-01-02"))

	longest, err := GetHabitLongestStreak("1")
	if err != nil {
		t.Fatalf("get longest streak: %v", err)
	}
	if longest != 5 {
		t.Fatalf("expected longest streak 5, got %d", longest)
	}
}