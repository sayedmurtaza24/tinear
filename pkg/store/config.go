package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

const sixMonthsAgo = -6 * 30 * 24 * time.Hour

func (s *Store) loadConfig() error {
	rows, err := s.db.Query("SELECT key, value FROM config;")
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("couldn't get config: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var key configKey
		var value string
		if err := rows.Scan(&key, &value); err != nil {
			return fmt.Errorf("couldn't scan config: %w", err)
		}
		switch key {
		case configLastOrgID:
			s.config.orgID = value
		case configIssuesSyncedAt:
			s.config.issues, _ = time.Parse(time.RFC3339, value)
		case configProjectsSyncedAt:
			s.config.projects, _ = time.Parse(time.RFC3339, value)
		case configTeamsSyncedAt:
			s.config.teams, _ = time.Parse(time.RFC3339, value)
		case configUsersSyncedAt:
			s.config.users, _ = time.Parse(time.RFC3339, value)
		}
	}

	if s.config.issues.IsZero() {
		s.config.issues = time.Now().UTC().Add(sixMonthsAgo)
	}
	if s.config.teams.IsZero() {
		s.config.teams = time.Now().UTC().Add(sixMonthsAgo)
	}
	if s.config.projects.IsZero() {
		s.config.projects = time.Now().UTC().Add(sixMonthsAgo)
	}
	if s.config.users.IsZero() {
		s.config.users = time.Now().UTC().Add(sixMonthsAgo)
	}

	return nil
}

func (s *Store) writeConfig(key configKey, v any) error {
	var value string

	switch v := v.(type) {
	case string:
		value = v
	case time.Time:
		value = v.Format(time.RFC3339)
	default:
		panic("unsupported")
	}

	_, err := s.db.Exec(`
		INSERT INTO config (key, value)
		VALUES (?, ?)
		ON CONFLICT (key) 
		DO UPDATE
		SET value = EXCLUDED.value;
	`, key, value)
	if err != nil {
		return fmt.Errorf("couldn't write config %s=%s: %w", key, value, err)
	}

	return nil
}
