package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func ApplyDir(ctx context.Context, db *sql.DB, dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		return fmt.Errorf("list migrations in %s: %w", dir, err)
	}

	sort.Strings(files)
	for _, file := range files {
		if err := ApplyFile(ctx, db, file); err != nil {
			return err
		}
	}

	return nil
}

func ApplyFile(ctx context.Context, db *sql.DB, path string) error {
	sqlBytes, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read migration %s: %w", path, err)
	}

	if _, err := db.ExecContext(ctx, string(sqlBytes)); err != nil {
		return fmt.Errorf("apply migration %s: %w", path, err)
	}

	return nil
}
