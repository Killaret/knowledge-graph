package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/argon2"
)

func main() {
	password := []byte("test")
	
	// Generate salt
	salt := make([]byte, 16)
	rand.Read(salt)
	
	// Hash with Argon2id (same parameters as in config)
	hash := argon2.IDKey(password, salt, 3, 65536, 4, 32)
	
	// Format as Argon2 string
	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)
	
	fmt.Printf("$argon2id$v=19$m=65536,t=3,p=4$%s$%s\n", encodedSalt, encodedHash)
}
