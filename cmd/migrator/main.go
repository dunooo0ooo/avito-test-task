package main

import (
	"database/sql"
	"fmt"
	"github.com/dunooo0ooo/avito-test-task/pkg/config"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	cfg := config.Load()

	dsn := cfg.Postgres.DSN()
	dir := cfg.Postgres.MigrationsDir

	log.Printf("running migrations: dir=%s\n", dir)

	if err := Up(dsn, dir); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	log.Println("migrations applied successfully")
}

func Up(dsn, dir string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}

	if err := goose.Up(db, dir); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}

	return nil
}
