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
