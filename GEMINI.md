# Gemini Guidelines

This document provides instructions for an AI assistant to help with development of this Habit Tracker application.

## Project Overview

This is a command-line habit tracking application written in Go. It uses the `bubbletea` library for the terminal user interface (TUI) and `bbolt` for the database.

## Commands

*   **Run:** `go run .`
*   **Test:** `go test ./...`
*   **Build:** `go build -o habit-tracker`

## Project Structure

*   `main.go`: The main entry point of the application.
*   `model/`: Contains the database logic and data structures.
    *   `db.go`: Handles all interactions with the `bbolt` database.
    *   `db_test.go`: Tests for the database logic.
*   `tui/`: Contains the terminal user interface logic.
    *   `app.go`: The main `bubbletea` application, handling UI and state.
    *   `app_test.go`: Tests for the TUI.
*   `tracker.db`: The `bbolt` database file.
*   `go.mod`, `go.sum`: Go module files.
