package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	var dsn string
	flag.StringVar(&dsn, "dsn", os.Getenv("DATABASE_URL"), "Postgres DSN (or set DATABASE_URL env)")
	var file string
	flag.StringVar(&file, "file", "scripts/seed_achievements.sql", "SQL file to run")
	flag.Parse()

	if dsn == "" {
		log.Fatal("Database DSN is required: set DATABASE_URL or pass -dsn")
	}

	sqlBytes, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalf("failed to read sql file: %v", err)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(string(sqlBytes)); err != nil {
		log.Fatalf("failed to execute seed sql: %v", err)
	}

	fmt.Println("Seed applied successfully")
}
