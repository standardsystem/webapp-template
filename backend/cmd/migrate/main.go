// Package main is the DB migration CLI for the webapp-template backend.
//
// Subcommands:
//
//	up                — apply all pending migrations
//	down              — roll back the latest migration
//	version           — print the current schema version
//	force <version>   — set the schema version without running migrations (recovery)
//
// DATABASE_URL environment variable is required.
//
// Migrations are embedded via migrations.FS and executed with golang-migrate,
// which uses pg_advisory_lock for safe concurrent execution and wraps each
// migration in a transaction.
package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/joho/godotenv"

	"github.com/your-org/webapp-template/migrations"
)

func main() {
	_ = godotenv.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		slog.Error("DATABASE_URL is required")
		os.Exit(1)
	}

	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		slog.Error("failed to open embedded migrations", "err", err)
		os.Exit(1)
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, dbURL)
	if err != nil {
		slog.Error("failed to initialize migrate", "err", err)
		os.Exit(1)
	}
	defer func() {
		if srcErr, dbErr := m.Close(); srcErr != nil || dbErr != nil {
			slog.Warn("migrate close errors", "source", srcErr, "database", dbErr)
		}
	}()

	switch cmd := os.Args[1]; cmd {
	case "up":
		runResult("up", m.Up())
	case "down":
		runResult("down", m.Steps(-1))
	case "version":
		v, dirty, err := m.Version()
		if err != nil {
			if errors.Is(err, migrate.ErrNilVersion) {
				slog.Info("no migrations applied yet")
				return
			}
			slog.Error("failed to read version", "err", err)
			os.Exit(1)
		}
		slog.Info("current version", "version", v, "dirty", dirty)
	case "force":
		if len(os.Args) < 3 {
			slog.Error("force requires a version argument")
			os.Exit(2)
		}
		v, err := strconv.Atoi(os.Args[2])
		if err != nil {
			slog.Error("invalid version", "value", os.Args[2], "err", err)
			os.Exit(2)
		}
		if err := m.Force(v); err != nil {
			slog.Error("force failed", "err", err)
			os.Exit(1)
		}
		slog.Info("forced version", "version", v)
	default:
		slog.Error("unknown subcommand", "cmd", cmd)
		usage()
		os.Exit(2)
	}
}

func runResult(name string, err error) {
	switch {
	case err == nil:
		slog.Info("migration succeeded", "command", name)
	case errors.Is(err, migrate.ErrNoChange):
		slog.Info("no migrations to apply", "command", name)
	default:
		slog.Error("migration failed", "command", name, "err", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: migrate <up|down|version|force <v>>")
}
