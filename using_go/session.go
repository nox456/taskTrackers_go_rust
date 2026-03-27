package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// ============================================================
// SESSION MANAGEMENT
// ============================================================
// The session is stored as a plain .session text file at the
// project root. It contains the SHA256 hash of the logged-in
// user's ID. This is intentionally simple — no tokens, no
// expiration, no encryption beyond the hash.
//
// On login:  write SHA256(userID) to .session
// On logout: delete .session
// On any command: read .session, find matching user
// ============================================================

const sessionFile = ".session"

func SaveSession(userID int) error {
	hash := HashSHA256(fmt.Sprintf("%d", userID))
	return os.WriteFile(sessionFile, []byte(hash), 0644)
}

func LoadSession(users []User) (*User, error) {
	data, err := os.ReadFile(sessionFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("no active session. Please login first")
		}
		return nil, fmt.Errorf("reading session: %w", err)
	}

	sessionHash := strings.TrimSpace(string(data))

	for i, user := range users {
		userHash := HashSHA256(fmt.Sprintf("%d", user.ID))
		if userHash == sessionHash {
			return &users[i], nil
		}
	}

	return nil, fmt.Errorf("session expired or invalid. Please login again")
}

func ClearSession() error {
	err := os.Remove(sessionFile)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("removing session: %w", err)
	}
	return nil
}
