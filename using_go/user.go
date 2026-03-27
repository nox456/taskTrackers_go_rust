package main

import (
	"crypto/sha256"
	"fmt"
)

// ============================================================
// USER STRUCT
// ============================================================
// The User struct represents an authenticated user in the system.
//
// The `Tasks` field uses the struct tag `json:"-"` which tells
// the JSON encoder to SKIP this field entirely. It won't appear
// in the JSON file and won't be read from it either.
//
// This is useful for "computed" or "derived" fields — data that
// exists in memory but shouldn't be persisted. Here, we populate
// Tasks by filtering all tasks where task.Owner == user.ID.
//
// TypeScript equivalent:
//   interface User {
//     id: number;
//     username: string;
//     password: string;       // stored as SHA256 hash
//     tasks?: Task[];         // not serialized, populated at runtime
//   }
// ============================================================

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Tasks    []Task `json:"-"`
}

func NewUser(id int, username, password string) User {
	return User{
		ID:       id,
		Username: username,
		Password: HashSHA256(password),
	}
}

// ============================================================
// CRYPTO/SHA256 — HASHING
// ============================================================
// sha256.Sum256 takes a byte slice and returns a fixed-size
// array [32]byte (256 bits = 32 bytes). We then format it as
// a hexadecimal string using %x.
//
// TypeScript equivalent (using Node.js crypto):
//   import { createHash } from 'crypto';
//   createHash('sha256').update(s).digest('hex');
//
// Note: SHA256 is a one-way hash — you can't reverse it.
// For passwords in production, you'd use bcrypt or argon2
// which add salting and are deliberately slow.
// ============================================================

func HashSHA256(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h)
}
