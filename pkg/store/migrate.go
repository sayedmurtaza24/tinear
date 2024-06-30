package store

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
)

//go:embed migrations/*.sql
var migrationsDir embed.FS

func (s *Store) migrate() error {
	var version int

	_ = s.db.QueryRow(`SELECT version FROM migrations ORDER BY version DESC LIMIT 1;`).Scan(&version)

	slog.Info("[store.migrate] latest migration", "version", version)

	i := version + 1
	for {
		migration, err := migrationsDir.ReadFile(fmt.Sprintf("migrations/v%d.sql", i))
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				slog.Info("[store.migrate] migrations ended")
				break
			}
			return fmt.Errorf("unexpected error: %w", err)
		}

		_, err = s.db.Exec(string(migration))
		if err != nil {
			return fmt.Errorf("error running migration %d: %w", version+i, err)
		}

		_, _ = s.db.Exec(`INSERT INTO migrations DEFAULT VALUES;`)

		slog.Info("[store.migrate] ran migration", "version", i)

		i++
	}

	return nil
}
