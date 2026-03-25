package main

import (
	"fmt"
	"strings"
	"time"
)

// ============================================================
// CUSTOM TYPES
// ============================================================
// In Go, you can create new types based on existing ones.
// This is STRONGER than TypeScript's `type State = string` —
// Go treats State as a completely distinct type from string.
// You can't accidentally pass a random string where a State
// is expected without an explicit conversion.
// ============================================================

type State string

// ============================================================
// CONSTANTS (Go's version of enums)
// ============================================================
// Go doesn't have enums like TypeScript (`enum State { ... }`)
// or Java (`enum State { ... }`). Instead, we group related
// constants in a `const (...)` block. This is idiomatic Go.
//
// Since State is a custom type, these constants are type-safe:
//   var s State = "random"  // won't compile — "random" isn't a State constant
//   var s State = Pending   // works
// ============================================================

const (
	Pending    State = "PENDING"
	InProgress State = "IN_PROGRESS"
	Done       State = "DONE"
)

// ============================================================
// STRUCTS (Go's version of classes)
// ============================================================
// In Go, structs replace classes. Key differences from TS/Java:
//   - No constructors — use factory functions (see NewTask below)
//   - No inheritance — use composition (embedding structs)
//   - No access modifiers (public/private/protected) — instead,
//     Uppercase = exported (public), lowercase = unexported (private)
//     This applies to ALL identifiers: types, functions, fields.
//
// STRUCT TAGS (`json:"..."`)
// The backtick strings after each field are "struct tags".
// They're metadata that other packages can read at runtime.
// The `json` tag tells Go's JSON encoder/decoder how to name
// each field in JSON. Without tags, it would use the Go field
// names (ID, Name, etc.) which might not match your JSON format.
//
// TypeScript equivalent:
//   interface Task {
//     id: number;          // json:"id"
//     name: string;        // json:"name"
//     description: string; // json:"description"
//     state: State;        // json:"state"
//   }
//
// Java equivalent: @JsonProperty("id") annotations on fields.
// ============================================================

type Task struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	State       State  `json:"state"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// ============================================================
// MULTIPLE RETURN VALUES & ERROR HANDLING
// ============================================================
// Go functions can return multiple values. This is how Go
// handles errors — instead of throwing exceptions (try/catch),
// functions return an error as the last value.
//
// Comparison:
//   TypeScript: function parseState(s: string): State { throw new Error(...) }
//   Python:     def parse_state(s: str) -> State: raise ValueError(...)
//   Java:       State parseState(String s) throws IllegalArgumentException { ... }
//   Go:         func ParseState(s string) (State, error) { return "", fmt.Errorf(...) }
//
// The caller MUST handle the error:
//   state, err := ParseState(input)
//   if err != nil { /* handle it */ }
//
// This is one of Go's most distinctive features: errors are values,
// not exceptions. You'll see this `if err != nil` pattern everywhere.
// ============================================================

func ParseState(s string) (State, error) {
	// strings.ToUpper works like s.toUpperCase() in TypeScript
	switch strings.ToUpper(s) {
	case string(Pending):
		return Pending, nil // nil is Go's null/undefined — means "no error"
	case string(InProgress), "IN PROGRESS", "INPROGRESS":
		return InProgress, nil
	case string(Done):
		return Done, nil
	default:
		// fmt.Errorf creates a new error with a formatted message.
		// %q prints the string in quotes (useful for showing user input).
		return "", fmt.Errorf("invalid state %q: must be PENDING, IN_PROGRESS, or DONE", s)
	}
}

// ============================================================
// FACTORY FUNCTIONS (Go's version of constructors)
// ============================================================
// Since Go has no constructors, the convention is to create a
// function named New<Type> that returns an initialized struct.
//
// TypeScript equivalent:
//   class Task { constructor(id: number, name: string, desc: string) { ... } }
//
// Note: this returns `Task` (a value), not `*Task` (a pointer).
// Structs in Go are value types — when you assign or return them,
// they get copied. For small structs like this, that's efficient.
// For large structs, you'd return a pointer (*Task) to avoid copying.
// ============================================================

func NewTask(id int, name, description string) Task {
	// time.Now() returns the current time.
	// .Format(time.RFC3339) converts it to a standard string format
	// like "2026-03-21T14:30:00Z". RFC3339 is Go's equivalent of
	// new Date().toISOString() in TypeScript.
	now := time.Now().Format(time.RFC3339)

	// This is a "struct literal" — like an object literal in TypeScript:
	//   { id: id, name: name, ... }
	// In Go, you MUST use field names in multi-line struct literals.
	return Task{
		ID:          id,
		Name:        name,
		Description: description,
		State:       Pending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}
