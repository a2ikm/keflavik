package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	migrate "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("invalid command")
	}

	switch os.Args[1] {
	case "create":
		return create()
	case "up":
		return up()
	case "down":
		return down()
	case "new":
		return new()
	}

	fmt.Printf("Usage: migrate <create|up|down|new>\n")
	return nil
}

func wrongUsage() error {
	return fmt.Errorf("usage: migrate <create|up>")
}

func create() error {
	db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5432?sslmode=disable")
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	query := `CREATE DATABASE keflavik`
	_, err = db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	return nil
}

func up() error {
	db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5432/keflavik?sslmode=disable")
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to initialize postgres: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://db/migrations",
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to initialize golang-migrate: %w", err)
	}

	m.Up()

	return nil
}

func down() error {
	db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5432/keflavik?sslmode=disable")
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to initialize postgres: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://db/migrations",
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to initialize golang-migrate: %w", err)
	}

	m.Down()

	return nil
}

func new() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("missing migration name")
	}
	name := os.Args[2]

	dirname := "db/migrations"
	if err := os.MkdirAll(dirname, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	version := time.Now().UTC().Format("20060102150405")

	for _, direction := range []string{"up", "down"} {
		basename := fmt.Sprintf("%s_%s.%s.sql", version, name, direction)

		filename := filepath.Join(dirname, basename)
		f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}

		err = f.Close()
		if err != nil {
			return fmt.Errorf("failed to close file: %w", err)
		}
	}

	return nil
}
