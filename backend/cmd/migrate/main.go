package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
)

const migrationsTable = "schema_migrations"

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	command := "up"
	if len(os.Args) > 1 {
		command = strings.TrimSpace(os.Args[1])
	}

	databaseURL, err := loadDatabaseURL()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	database, err := db.New(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer database.Close()

	migrationsDir, err := findMigrationsDir()
	if err != nil {
		return err
	}

	switch command {
	case "up":
		return migrateUp(ctx, database, migrationsDir)
	case "status":
		return printStatus(ctx, database, migrationsDir)
	case "down":
		return errors.New("migrate down is not supported yet because current SQL migrations do not include down sections")
	default:
		return fmt.Errorf("unknown migrate command %q; use: up, status", command)
	}
}

func loadDatabaseURL() (string, error) {
	_ = godotenv.Load(".env")
	_ = godotenv.Load(filepath.Join("backend", ".env"))

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return "", errors.New("DATABASE_URL is required for migration")
	}

	return databaseURL, nil
}

func findMigrationsDir() (string, error) {
	for _, candidate := range []string{"migrations", filepath.Join("backend", "migrations")} {
		info, err := os.Stat(candidate)
		if err == nil && info.IsDir() {
			return candidate, nil
		}
	}

	return "", errors.New("migrations directory not found; run from backend/ or repository root")
}

func migrateUp(ctx context.Context, database *db.DB, migrationsDir string) error {
	if err := ensureMigrationsTable(ctx, database); err != nil {
		return err
	}

	files, err := listMigrationFiles(migrationsDir)
	if err != nil {
		return err
	}

	applied, err := loadAppliedMigrations(ctx, database)
	if err != nil {
		return err
	}

	appliedCount := 0
	for _, file := range files {
		if applied[file.Name] {
			continue
		}

		sqlBytes, err := os.ReadFile(file.Path)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", file.Name, err)
		}

		log.Printf("applying migration %s", file.Name)
		if err := database.WithTx(ctx, func(tx db.Tx) error {
			if _, err := tx.Exec(ctx, string(sqlBytes)); err != nil {
				return fmt.Errorf("execute migration %s: %w", file.Name, err)
			}

			const insertQuery = `
				INSERT INTO schema_migrations (version, name)
				VALUES ($1, $2)
			`
			if _, err := tx.Exec(ctx, insertQuery, file.Name, file.Name); err != nil {
				return fmt.Errorf("record migration %s: %w", file.Name, err)
			}

			return nil
		}); err != nil {
			return err
		}

		appliedCount++
		log.Printf("applied migration %s", file.Name)
	}

	if appliedCount == 0 {
		log.Println("database already up to date")
		return nil
	}

	log.Printf("applied %d migration(s)", appliedCount)
	return nil
}

func printStatus(ctx context.Context, database *db.DB, migrationsDir string) error {
	if err := ensureMigrationsTable(ctx, database); err != nil {
		return err
	}

	files, err := listMigrationFiles(migrationsDir)
	if err != nil {
		return err
	}

	applied, err := loadAppliedMigrations(ctx, database)
	if err != nil {
		return err
	}

	for _, file := range files {
		status := "pending"
		if applied[file.Name] {
			status = "applied"
		}
		log.Printf("%s %s", status, file.Name)
	}

	return nil
}

func ensureMigrationsTable(ctx context.Context, database *db.DB) error {
	const query = `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)
	`

	if _, err := database.Exec(ctx, query); err != nil {
		return fmt.Errorf("ensure %s table: %w", migrationsTable, err)
	}

	return nil
}

type migrationFile struct {
	Name string
	Path string
}

func listMigrationFiles(migrationsDir string) ([]migrationFile, error) {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("read migrations directory: %w", err)
	}

	files := make([]migrationFile, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		files = append(files, migrationFile{
			Name: entry.Name(),
			Path: filepath.Join(migrationsDir, entry.Name()),
		})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name < files[j].Name
	})

	return files, nil
}

func loadAppliedMigrations(ctx context.Context, database *db.DB) (map[string]bool, error) {
	const query = `
		SELECT version
		FROM schema_migrations
	`

	rows, err := database.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("load applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("scan applied migration: %w", err)
		}
		applied[version] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate applied migrations: %w", err)
	}

	return applied, nil
}
