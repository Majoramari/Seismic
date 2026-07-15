package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RunMigrations applies any .sql files in db/migrations/ that haven't been applied yet.
// It tracks which ones have already been applied in a table called migrations
func RunMigrations(pool *pgxpool.Pool) error {
	ctx := context.Background()

	// Create the tracking table if it doesn't exist yet
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS migrations (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Find all .sql files in the migrations folder
	files, err := filepath.Glob("db/migrations/*.sql")
	if err != nil {
		return fmt.Errorf("failed to find migration files: %w", err)
	}
	sort.Strings(files) // ensures they're applied in order (001_ runs before 002_)

	for _, file := range files {
		filename := filepath.Base(file)

		// Check if this migration already ran
		var alreadyApplied bool
		err := pool.QueryRow(ctx,
			"SELECT EXISTS (SELECT 1 FROM migrations WHERE name = $1)",
			filename,
		).Scan(&alreadyApplied)
		if err != nil {
			return fmt.Errorf("failed to check migration status: %w", err)
		}

		if alreadyApplied {
			continue // skip already applied migrations
		}

		// Apply the migration
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", filename, err)
		}

		log.Printf("Applying migration %s", filename)
		_, err = pool.Exec(ctx, string(content))
		if err != nil {
			return fmt.Errorf("failed to apply %s: %w", filename, err)
		}

		// Record that this migration has now been applied,
		// so it won't run again on the next restart
		_, err = pool.Exec(ctx,
			"INSERT INTO migrations (name) VALUES ($1)",
			filename,
		)
		if err != nil {
			return fmt.Errorf("failed to record migration %s: %w", filename, err)
		}

		log.Printf("Applied migration %s", filename)
	}

	return nil
}
