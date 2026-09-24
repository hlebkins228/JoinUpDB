package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"

	"JoinUp/internal/settings"
	appmigrations "JoinUp/migrations"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	command := "up"
	args := os.Args[1:]
	if len(args) > 0 {
		command = args[0]
		args = args[1:]
	}

	db, err := sql.Open("pgx", settings.ReadDbConnectionConfig())
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("ping database: %v", err)
	}

	if err := appmigrations.Run(ctx, db, command, args...); err != nil {
		log.Fatalf("run migrations command %q: %v", command, err)
	}
}
