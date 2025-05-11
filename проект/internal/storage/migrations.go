package storage

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const migrationsTable = "schema_migrations"

type Migration struct {
	Version int
	Query   string
}

func (s *Storage) Migrate() error {
	if err := s.createMigrationsTable(); err != nil {
		return fmt.Errorf("create migrations table failed: %w", err)
	}

	applied, err := s.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("get applied migrations failed: %w", err)
	}

	migrations, err := s.loadMigrations()
	if err != nil {
		return fmt.Errorf("load migrations failed: %w", err)
	}

	for _, m := range migrations {
		if _, ok := applied[m.Version]; !ok {
			if err := s.applyMigration(m); err != nil {
				return fmt.Errorf("apply migration %d failed: %w", m.Version, err)
			}
		}
	}

	return nil
}

func (s *Storage) createMigrationsTable() error {
	_, err := s.db.Exec(fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s (
            version INTEGER PRIMARY KEY,
            applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )
    `, migrationsTable))
	return err
}

func (s *Storage) getAppliedMigrations() (map[int]struct{}, error) {
	rows, err := s.db.Query(fmt.Sprintf(
		"SELECT version FROM %s ORDER BY version ASC",
		migrationsTable,
	))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	migrations := make(map[int]struct{})
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		migrations[version] = struct{}{}
	}
	return migrations, nil
}

func (s *Storage) loadMigrations() ([]Migration, error) {
	files, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return nil, err
	}

	var migrations []Migration
	for _, f := range files {
		if filepath.Ext(f.Name()) != ".sql" {
			continue
		}

		base := filepath.Base(f.Name())
		parts := strings.SplitN(base, "_", 2)
		version, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid migration name %s: %w", f.Name(), err)
		}

		content, err := fs.ReadFile(migrationsFS, filepath.Join("migrations", f.Name()))
		if err != nil {
			return nil, err
		}

		migrations = append(migrations, Migration{
			Version: version,
			Query:   string(content),
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func (s *Storage) applyMigration(m Migration) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(m.Query); err != nil {
		return fmt.Errorf("execute query failed: %w", err)
	}

	if _, err := tx.Exec(
		fmt.Sprintf("INSERT INTO %s (version) VALUES (?)", migrationsTable),
		m.Version,
	); err != nil {
		return fmt.Errorf("record migration failed: %w", err)
	}

	return tx.Commit()
}

var migrationsFS embed.FS
