package main

import (
	"fmt"
	"log"

	"knowledge-graph/internal/auth"
)

func main() {
	password := "test"
	config := &auth.PasswordConfig{
		Time:    3,
		Memory:  65536,
		Threads: 4,
		KeyLen:  32,
	}
	
	hash, err := auth.HashPassword(password, config)
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("Hash: %s\n", hash)
}
