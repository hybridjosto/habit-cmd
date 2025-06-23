// File: model/db.go
package model

import (
	"encoding/json"
	"time"

	bolt "go.etcd.io/bbolt"
)

var db *bolt.DB

var habitsBucket = []byte("habits")

type Habit struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func InitDB(path string) error {
	var err error
	db, err = bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}
	return db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(habitsBucket)
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

func AddHabit(id, name string) error {
	h := Habit{ID: id, Name: name}
	data, err := json.Marshal(h)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(habitsBucket)
		return b.Put([]byte(id), data)
	})
}

func CloseDB() {
	if db != nil {
		db.Close()
	}
}
