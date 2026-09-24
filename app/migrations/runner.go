package migrations

import (
	"context"
	"database/sql"
	"embed"

	"github.com/pressly/goose/v3"
)

const embeddedDir = "."

//go:embed *.sql
var embeddedMigrations embed.FS

func Up(ctx context.Context, db *sql.DB) error {
	return run(ctx, db, "up")
}

func Down(ctx context.Context, db *sql.DB) error {
	return run(ctx, db, "down")
}

func Status(ctx context.Context, db *sql.DB) error {
	return run(ctx, db, "status")
}

func Run(ctx context.Context, db *sql.DB, command string, args ...string) error {
	return run(ctx, db, command, args...)
}

func run(ctx context.Context, db *sql.DB, command string, args ...string) error {
	goose.SetBaseFS(embeddedMigrations)
	defer goose.SetBaseFS(nil)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	return goose.RunContext(ctx, command, db, embeddedDir, args...)
}
