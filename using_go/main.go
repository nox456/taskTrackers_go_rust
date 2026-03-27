package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// ============================================================
// ENTRY POINT
// ============================================================
// In Go, the program starts at func main() in package "main".
// Every Go file must declare its package at the top.
// All files in the same directory that say `package main`
// are compiled together — they can use each other's exported
// AND unexported names without importing anything.
//
// This is different from TypeScript where you need:
//   import { Task } from './task'
//
// In Go, if task.go and main.go both say `package main`,
// main.go can freely use Task, NewTask, Storage, etc.
// ============================================================

const dataFile = "tasks.json"

func main() {
	// ============================================================
	// OS.ARGS — POSITIONAL ARGUMENTS
	// ============================================================
	// os.Args is a slice of strings containing CLI arguments.
	//   os.Args[0] = program name (like process.argv[0] in Node.js)
	//   os.Args[1:] = actual arguments
	//
	// The [1:] syntax is a "slice expression" — it creates a new
	// slice from index 1 to the end. Works like array.slice(1) in JS.
	// More examples:
	//   args[1:3]  = from index 1 to 2 (exclusive end, like JS)
	//   args[:3]   = from start to index 2
	//   args[2:]   = from index 2 to end
	// ============================================================
	args := os.Args[1:]

	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	storage := NewStorage(dataFile)

	// ============================================================
	// SWITCH (auto-break)
	// ============================================================
	// Unlike TypeScript/Java/C, Go's switch breaks automatically
	// after each case. No need to write `break`.
	// If you actually WANT fall-through, use the `fallthrough` keyword.
	//
	// Commands are split into two groups:
	//   1. Auth commands (register, login, logout) — no session needed
	//   2. Task commands (add, list, update, delete, status) — require
	//      an active session so we know which user is operating
	// ============================================================
	command := args[0]

	switch command {
	case "register":
		handleRegister(storage, args[1:])
		return
	case "login":
		handleLogin(storage, args[1:])
		return
	case "logout":
		handleLogout()
		return
	}

	// All task commands require an active session
	store, err := storage.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading data: %v\n", err)
		os.Exit(1)
	}

	user, err := LoadSession(store.Users)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	switch command {
	case "add":
		handleAdd(storage, args[1:], user)
	case "list":
		handleList(storage, args[1:], user)
	case "update":
		handleUpdate(storage, args[1:], user)
	case "delete":
		handleDelete(storage, args[1:], user)
	case "status":
		handleStatus(storage, args[1:], user)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Task Tracker CLI")
	fmt.Println()
	fmt.Println("Auth:")
	fmt.Println("  task-tracker register <username> <password>")
	fmt.Println("  task-tracker login <username> <password>")
	fmt.Println("  task-tracker logout")
	fmt.Println()
	fmt.Println("Tasks (requires login):")
	fmt.Println("  task-tracker add <name> <description>")
	fmt.Println("  task-tracker list [state]")
	fmt.Println("  task-tracker update <id> <name> <description>")
	fmt.Println("  task-tracker delete <id>")
	fmt.Println("  task-tracker status <id> <state>")
	fmt.Println()
	fmt.Println("States: PENDING, IN_PROGRESS, DONE")
}

// ============================================================
// AUTH HANDLERS
// ============================================================

func handleRegister(s *Storage, args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Error: 'register' requires <username> and <password>")
		fmt.Fprintln(os.Stderr, "Usage: task-tracker register <username> <password>")
		os.Exit(1)
	}

	store, err := s.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading data: %v\n", err)
		os.Exit(1)
	}

	// Check if username is already taken
	_, findErr := s.FindUserByUsername(&store, args[0])
	if findErr == nil {
		fmt.Fprintf(os.Stderr, "Error: username %q is already taken\n", args[0])
		os.Exit(1)
	}

	user := NewUser(store.NextUserID, args[0], args[1])
	store.Users = append(store.Users, user)
	store.NextUserID++

	if err := s.Save(store); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving data: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("User %q registered successfully (ID: %d)\n", user.Username, user.ID)
}

func handleLogin(s *Storage, args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Error: 'login' requires <username> and <password>")
		fmt.Fprintln(os.Stderr, "Usage: task-tracker login <username> <password>")
		os.Exit(1)
	}

	store, err := s.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading data: %v\n", err)
		os.Exit(1)
	}

	user, err := s.FindUserByUsername(&store, args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: invalid username or password")
		os.Exit(1)
	}

	if user.Password != HashSHA256(args[1]) {
		fmt.Fprintln(os.Stderr, "Error: invalid username or password")
		os.Exit(1)
	}

	if err := SaveSession(user.ID); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating session: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Welcome back, %s!\n", user.Username)
}

func handleLogout() {
	if err := ClearSession(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Logged out successfully.")
}

// ============================================================
// TASK HANDLERS
// ============================================================

func handleAdd(s *Storage, args []string, user *User) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Error: 'add' requires <name> and <description>")
		fmt.Fprintln(os.Stderr, "Usage: task-tracker add <name> <description>")
		os.Exit(1)
	}

	store, err := s.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading tasks: %v\n", err)
		os.Exit(1)
	}

	task := NewTask(store.NextID, args[0], args[1], user.ID)

	// ============================================================
	// APPEND (growing slices)
	// ============================================================
	// Slices are Go's dynamic arrays (like ArrayList in Java or
	// Array in TypeScript). Unlike those, append() returns a NEW
	// slice rather than modifying in place:
	//
	//   TypeScript: tasks.push(task)        // modifies array
	//   Go:         tasks = append(tasks, task)  // returns new slice
	//
	// Why? Slices are backed by fixed-size arrays. When the array
	// is full, Go allocates a bigger one and copies the data.
	// The returned slice may point to a different array.
	// ============================================================
	store.Tasks = append(store.Tasks, task)
	store.NextID++

	if err := s.Save(store); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving tasks: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Task added successfully (ID: %d)\n", task.ID)
}

func handleList(s *Storage, args []string, user *User) {
	store, err := s.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading tasks: %v\n", err)
		os.Exit(1)
	}

	var filterState State
	if len(args) > 0 {
		filterState, err = ParseState(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}

	// ============================================================
	// FMT.PRINTF — FORMATTED OUTPUT
	// ============================================================
	// Printf works like printf in C, sprintf in Python, or
	// template literals in TypeScript — but with format verbs:
	//
	//   %s  = string       (like ${str} in JS template literals)
	//   %d  = integer      (like ${num})
	//   %v  = any value    (Go figures out the format)
	//   %q  = quoted string (adds quotes around it)
	//   %-4d = left-aligned integer, 4 chars wide (for tables)
	//   %-20s = left-aligned string, 20 chars wide
	//   \n  = newline (Printf doesn't add one automatically)
	// ============================================================
	fmt.Printf("%-4s %-20s %-30s %-12s %s\n", "ID", "Name", "Description", "State", "Updated")
	fmt.Println("---- -------------------- ------------------------------ ------------ -------------------")

	found := false
	for _, task := range store.Tasks {
		if task.Owner != user.ID {
			continue
		}
		if filterState != "" && task.State != filterState {
			continue
		}
		found = true
		fmt.Printf("%-4d %-20s %-30s %-12s %s\n",
			task.ID,
			truncate(task.Name, 20),
			truncate(task.Description, 30),
			task.State,
			task.UpdatedAt,
		)
	}

	if !found {
		if filterState != "" {
			fmt.Printf("No tasks with state %s.\n", filterState)
		} else {
			fmt.Println("No tasks found.")
		}
	}
}

func handleUpdate(s *Storage, args []string, user *User) {
	if len(args) < 3 {
		fmt.Fprintln(os.Stderr, "Error: 'update' requires <id> <name> <description>")
		fmt.Fprintln(os.Stderr, "Usage: task-tracker update <id> <name> <description>")
		os.Exit(1)
	}

	// ============================================================
	// STRCONV.ATOI — STRING TO INTEGER
	// ============================================================
	// "Atoi" = "ASCII to integer" (name from C's standard library).
	//   TypeScript: parseInt(args[0])  — returns NaN on failure
	//   Python:     int(args[0])       — raises ValueError on failure
	//   Go:         strconv.Atoi(args[0]) — returns (int, error)
	//
	// Go's approach: instead of NaN or exceptions, you get an explicit
	// error that you MUST handle. No silent failures.
	// ============================================================
	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid ID %q - must be a number\n", args[0])
		os.Exit(1)
	}

	store, err := s.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading tasks: %v\n", err)
		os.Exit(1)
	}

	task, _, err := s.FindByID(&store, id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if task.Owner != user.ID {
		fmt.Fprintln(os.Stderr, "Error: you can only update your own tasks")
		os.Exit(1)
	}

	// Because FindByID returns a *Task (pointer), modifying `task`
	// here modifies the actual task inside store.Tasks.
	// If it returned Task (value), this would modify a copy
	// and the original would stay unchanged — a common Go gotcha!
	task.Name = args[1]
	task.Description = args[2]
	task.UpdatedAt = time.Now().Format(time.RFC3339)

	if err := s.Save(store); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving tasks: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Task %d updated successfully.\n", id)
}

func handleDelete(s *Storage, args []string, user *User) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Error: 'delete' requires <id>")
		fmt.Fprintln(os.Stderr, "Usage: task-tracker delete <id>")
		os.Exit(1)
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid ID %q - must be a number\n", args[0])
		os.Exit(1)
	}

	store, err := s.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading tasks: %v\n", err)
		os.Exit(1)
	}

	task, index, err := s.FindByID(&store, id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if task.Owner != user.ID {
		fmt.Fprintln(os.Stderr, "Error: you can only delete your own tasks")
		os.Exit(1)
	}

	// ============================================================
	// DELETING FROM A SLICE
	// ============================================================
	// Go has no built-in "remove at index" for slices. The pattern:
	//   slice = append(slice[:i], slice[i+1:]...)
	//
	// Breaking it down:
	//   slice[:i]     = everything BEFORE index i
	//   slice[i+1:]   = everything AFTER index i
	//   append(a, b...) = concatenate them
	//
	// The `...` operator unpacks a slice into individual elements,
	// similar to the spread operator (...) in TypeScript:
	//   TypeScript: [...before, ...after]
	//   Go:         append(before, after...)
	// ============================================================
	store.Tasks = append(store.Tasks[:index], store.Tasks[index+1:]...)

	if err := s.Save(store); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving tasks: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Task %d deleted successfully.\n", id)
}

func handleStatus(s *Storage, args []string, user *User) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Error: 'status' requires <id> <state>")
		fmt.Fprintln(os.Stderr, "Usage: task-tracker status <id> <state>")
		fmt.Fprintln(os.Stderr, "States: PENDING, IN_PROGRESS, DONE")
		os.Exit(1)
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid ID %q - must be a number\n", args[0])
		os.Exit(1)
	}

	newState, err := ParseState(args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	store, err := s.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading tasks: %v\n", err)
		os.Exit(1)
	}

	task, _, err := s.FindByID(&store, id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if task.Owner != user.ID {
		fmt.Fprintln(os.Stderr, "Error: you can only change the status of your own tasks")
		os.Exit(1)
	}

	task.State = newState
	task.UpdatedAt = time.Now().Format(time.RFC3339)

	if err := s.Save(store); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving tasks: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Task %d state changed to %s.\n", id, newState)
}

// truncate shortens a string to maxLen characters.
func truncate(s string, maxLen int) string {
	// ============================================================
	// RUNES vs BYTES
	// ============================================================
	// Go strings are sequences of BYTES, not characters.
	// For ASCII this doesn't matter, but for characters like
	// Spanish "n" or emojis, one character can be multiple bytes.
	//
	// A "rune" is Go's name for a Unicode code point (like Java's char).
	// Converting string -> []rune gives you actual characters:
	//   "cafe" = 4 runes, 4 bytes
	//   "hola" = 4 runes, 4 bytes (all ASCII)
	// ============================================================
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-3]) + "..."
}
