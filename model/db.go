// File: model/db.go
package model

import (
	"encoding/json"
	"time"

	bolt "go.etcd.io/bbolt"
)

var db *bolt.DB

var (
	habitsBucket      = []byte("habits")
	completionsBucket = []byte("completions")
	tasksBucket       = []byte("tasks")
)

type Habit struct {
	ID    string            `json:"id"`
	Name  string            `json:"name"`
	Type  string            `json:"type"` // "general" or "daily"
	Notes map[string]string `json:"notes"`
}

type HabitCompletion struct {
	HabitID string `json:"habit_id"`
	Date    string `json:"date"`
}

type Task struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	DueDate     string `json:"due_date"`
	Completed   bool   `json:"completed"`
	CreatedAt   string `json:"created_at"`
}

func InitDB(path string) error {
	var err error
	db, err = bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}
	return db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(habitsBucket)
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists(completionsBucket)
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists(tasksBucket)
		return err
	})
}

func GetHabits() ([]Habit, error) {
	var habits []Habit
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(habitsBucket)
		return b.ForEach(func(k, v []byte) error {
			var h Habit
			if err := json.Unmarshal(v, &h); err != nil {
				return err
			}
			habits = append(habits, h)
			return nil
		})
	})
	return habits, err
}

func AddHabit(id, name, habitType string, notes map[string]string) error {
	h := Habit{ID: id, Name: name, Type: habitType, Notes: notes}
	data, err := json.Marshal(h)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(habitsBucket)
		return b.Put([]byte(id), data)
	})
}

func UpdateHabit(id string, habit Habit) error {
	data, err := json.Marshal(habit)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(habitsBucket)
		return b.Put([]byte(id), data)
	})
}

func ToggleHabitCompletion(habitID, date string) error {
	key := habitID + "_" + date
	keyBytes := []byte(key)
	
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(completionsBucket)
		if b.Get(keyBytes) != nil {
			return b.Delete(keyBytes)
		}
		completion := HabitCompletion{HabitID: habitID, Date: date}
		data, err := json.Marshal(completion)
		if err != nil {
			return err
		}
		return b.Put(keyBytes, data)
	})
}

func IsHabitCompleted(habitID, date string) (bool, error) {
	key := habitID + "_" + date
	keyBytes := []byte(key)
	
	var exists bool
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(completionsBucket)
		exists = b.Get(keyBytes) != nil
		return nil
	})
	return exists, err
}

func DeleteHabit(id string) error {
	return db.Update(func(tx *bolt.Tx) error {
		habitsB := tx.Bucket(habitsBucket)
		completionsB := tx.Bucket(completionsBucket)
		
		if err := habitsB.Delete([]byte(id)); err != nil {
			return err
		}
		
		return completionsB.ForEach(func(k, v []byte) error {
			var completion HabitCompletion
			if err := json.Unmarshal(v, &completion); err != nil {
				return nil
			}
			if completion.HabitID == id {
				return completionsB.Delete(k)
			}
			return nil
		})
	})
}

func AddTask(id, name, description, dueDate string) error {
	task := Task{
		ID:          id,
		Name:        name,
		Description: description,
		DueDate:     dueDate,
		Completed:   false,
		CreatedAt:   time.Now().Format("2006-01-02 15:04:05"),
	}
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(tasksBucket)
		return b.Put([]byte(id), data)
	})
}

func GetTasks() ([]Task, error) {
	var tasks []Task
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(tasksBucket)
		return b.ForEach(func(k, v []byte) error {
			var t Task
			if err := json.Unmarshal(v, &t); err != nil {
				return err
			}
			tasks = append(tasks, t)
			return nil
		})
	})
	return tasks, err
}

func ToggleTask(id string) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(tasksBucket)
		data := b.Get([]byte(id))
		if data == nil {
			return nil
		}
		
		var task Task
		if err := json.Unmarshal(data, &task); err != nil {
			return err
		}
		
		task.Completed = !task.Completed
		
		updatedData, err := json.Marshal(task)
		if err != nil {
			return err
		}
		
		return b.Put([]byte(id), updatedData)
	})
}

func DeleteTask(id string) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(tasksBucket)
		return b.Delete([]byte(id))
	})
}

func GetHabitStreak(habitID string) (int, error) {
	today := time.Now()
	streak := 0
	
	for i := 0; i < 365; i++ {
		date := today.AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")
		
		completed, err := IsHabitCompleted(habitID, dateStr)
		if err != nil {
			return 0, err
		}
		
		if completed {
			streak++
		} else {
			break
		}
	}
	
	return streak, nil
}

func GetHabitLongestStreak(habitID string) (int, error) {
	longestStreak := 0
	currentStreak := 0
	
	today := time.Now()
	for i := 365; i >= 0; i-- {
		date := today.AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")
		
		completed, err := IsHabitCompleted(habitID, dateStr)
		if err != nil {
			return 0, err
		}
		
		if completed {
			currentStreak++
			if currentStreak > longestStreak {
				longestStreak = currentStreak
			}
		} else {
			currentStreak = 0
		}
	}
	
	return longestStreak, nil
}

func CloseDB() {
	if db != nil {
		db.Close()
	}
}
