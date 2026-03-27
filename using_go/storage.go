package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// TaskStore is the root structure of our JSON file.
// It holds all tasks and users, plus counters for generating unique IDs.
type TaskStore struct {
	NextID     int    `json:"next_id"`
	NextUserID int    `json:"next_user_id"`
	Tasks      []Task `json:"tasks"`
	Users      []User `json:"users"`
}

// ============================================================
// UNEXPORTED FIELDS & POINTERS
// ============================================================
// `filePath` starts with a lowercase letter, so it's unexported
// (private). Code outside this package can't access it directly.
//
// In the function signatures below, you'll see *Storage (pointer
// to Storage). Pointers in Go are simpler than in C/C++:
//   - & takes the address of a value   (like "give me a reference")
//   - * dereferences a pointer          (like "give me the value")
//   - *Type in a function signature means "pointer to Type"
//
// Why use pointers? Two reasons:
//   1. Avoid copying large structs (performance)
//   2. Allow methods to modify the struct they're called on
//
// TypeScript/Python always pass objects by reference.
// Go passes structs by VALUE (copies them), so you need
// pointers when you want reference-like behavior.
// ============================================================

type Storage struct {
	filePath string
}

// NewStorage creates a new Storage instance.
// It returns *Storage (a pointer) so all methods can share
// the same instance without copying.
func NewStorage(filePath string) *Storage {
	return &Storage{filePath: filePath}
}

// ============================================================
// METHODS (functions attached to a type)
// ============================================================
// The `(s *Storage)` before the function name is the "receiver".
// It makes this function a method on Storage, similar to:
//   TypeScript: class Storage { load() { this.filePath ... } }
//   Java:       class Storage { TaskStore load() { this.filePath ... } }
//
// In Go, `s` is the explicit version of `this`/`self`.
// You pick the name (convention: short, often first letter of type).
// ============================================================

func (s *Storage) Load() (TaskStore, error) {
	store := TaskStore{NextID: 1, NextUserID: 1}

	// os.ReadFile reads the entire file into memory as a byte slice.
	// A byte slice ([]byte) is Go's way of handling raw binary data.
	// Think of it like a Buffer in Node.js.
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		// errors.Is() checks if an error matches a specific known error.
		// os.ErrNotExist is the "file not found" error — like ENOENT in Node.js.
		if errors.Is(err, os.ErrNotExist) {
			return store, nil
		}
		// %w "wraps" the error, preserving the original cause.
		// This is Go's error chaining — similar to `new Error("...", { cause: err })` in JS.
		return store, fmt.Errorf("reading file: %w", err)
	}

	// ============================================================
	// JSON UNMARSHALING
	// ============================================================
	// json.Unmarshal converts JSON bytes into a Go struct.
	//   TypeScript: const store = JSON.parse(data) as TaskStore;
	//   Go:         json.Unmarshal(data, &store)
	//
	// Key difference: Go doesn't return a new object. Instead, you
	// pass a POINTER (&store) and it fills in the existing struct.
	// This is a common Go pattern — "fill this struct for me".
	// ============================================================
	if err := json.Unmarshal(data, &store); err != nil {
		return store, fmt.Errorf("parsing JSON: %w", err)
	}

	return store, nil
}

func (s *Storage) Save(store TaskStore) error {
	// ============================================================
	// JSON MARSHALING
	// ============================================================
	// json.MarshalIndent converts a Go struct to formatted JSON bytes.
	//   TypeScript: JSON.stringify(store, null, 2)
	//   Go:         json.MarshalIndent(store, "", "  ")
	//
	// The two string args are: prefix (before each line) and indent.
	// ============================================================
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding JSON: %w", err)
	}

	// os.WriteFile writes bytes to a file (creates if needed).
	// 0644 is the Unix file permission:
	//   6 = owner can read+write
	//   4 = group can read
	//   4 = others can read
	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	return nil
}

// FindByID searches for a task by its ID in the store.
// Returns:
//   - *Task:  pointer to the task (so the caller can modify it in place)
//   - int:    the index in the slice (useful for deletion)
//   - error:  non-nil if not found
func (s *Storage) FindByID(store *TaskStore, id int) (*Task, int, error) {
	// ============================================================
	// RANGE LOOPS
	// ============================================================
	// `range` iterates over slices, maps, strings, and channels.
	//   TypeScript: store.tasks.forEach((task, i) => { ... })
	//   Python:     for i, task in enumerate(store.tasks): ...
	//   Go:         for i, task := range store.Tasks { ... }
	//
	// IMPORTANT: `task` in the range loop is a COPY, not a reference.
	// If you want to modify the original, use the index:
	//   store.Tasks[i].Name = "new name"   // modifies original
	//   task.Name = "new name"             // modifies the copy (lost!)
	// ============================================================
	for i, task := range store.Tasks {
		if task.ID == id {
			// Return a pointer to the actual slice element, not the copy.
			return &store.Tasks[i], i, nil
		}
	}

	return nil, -1, fmt.Errorf("task with ID %d not found", id)
}

func (s *Storage) FindUserByUsername(store *TaskStore, username string) (*User, error) {
	for i, user := range store.Users {
		if user.Username == username {
			return &store.Users[i], nil
		}
	}
	return nil, fmt.Errorf("user %q not found", username)
}
