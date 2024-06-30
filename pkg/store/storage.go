package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/glebarez/go-sqlite"
)

type configKey string

const (
	configLastOrgID        configKey = "last_org_id"
	configIssuesSyncedAt   configKey = "issues_synced_at"
	configTeamsSyncedAt    configKey = "teams_synced_at"
	configProjectsSyncedAt configKey = "projects_synced_at"
	configUsersSyncedAt    configKey = "users_synced_at"
)

var ErrNoOrgSelected = errors.New("no org is selected")

type configRow struct {
	key   configKey
	value string
}

type config struct {
	orgID    string
	issues   time.Time
	projects time.Time
	teams    time.Time
	users    time.Time
}

type Store struct {
	db     *sql.DB
	config config
}

func New(path string) (*Store, error) {
	db, err := sql.Open("sqlite", fmt.Sprintf("%s?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)", path))
	if err != nil {
		return nil, fmt.Errorf("couldn't instantiate db: %w", err)
	}

	// to avoid locks
	db.SetMaxOpenConns(1)

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("couldn't ping db: %w", err)
	}

	store := &Store{db: db}

	err = store.migrate()
	if err != nil {
		return nil, fmt.Errorf("couldn't migrate db: %w", err)
	}

	err = store.loadConfig()
	if err != nil {
		return nil, fmt.Errorf("couldn't load config: %w", err)
	}

	return store, nil
}

func (s *Store) Orgs(ctx context.Context) ([]Organization, error) {
	orgs, err := batchSelect(
		s.db,
		"orgs",
		[]string{"id", "name"},
		func(o *Organization) []any { return []any{&o.ID, &o.Name} },
		"",
	)
	if err != nil {
		return nil, fmt.Errorf("couldn't select orgs: %w", err)
	}
	return orgs, nil
}

func (s *Store) StoreOrgs(ctx context.Context, orgs []Organization) error {
	err := batchInsert(
		s.db,
		"orgs",
		[]string{"id", "name"},
		orgs,
		func(o *Organization) []any { return []any{o.ID, o.Name} },
		"ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name",
	)
	if err != nil {
		return fmt.Errorf("couldn't store orgs: %w", err)
	}
	return nil
}

func (s *Store) Teams(ctx context.Context) ([]Team, error) {
	if s.config.orgID == "" {
		return nil, ErrNoOrgSelected
	}
	teams, err := batchSelect(
		s.db,
		"teams",
		[]string{"id", "name", "color"},
		func(t *Team) []any { return []any{&t.ID, &t.Name, &t.Color} },
		fmt.Sprintf("WHERE org_id = '%s'", s.config.orgID),
	)
	if err != nil {
		return nil, fmt.Errorf("couldn't select teams: %w", err)
	}
	return teams, nil
}

func (s *Store) StoreTeams(ctx context.Context, teams []Team) error {
	if s.config.orgID == "" {
		return ErrNoOrgSelected
	}
	err := batchInsert(
		s.db,
		"teams",
		[]string{"id", "name", "color", "org_id"},
		teams,
		func(t *Team) []any { return []any{t.ID, t.Name, t.Color, s.config.orgID} },
		"ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, color = EXCLUDED.color",
	)
	if err != nil {
		return fmt.Errorf("couldn't store teams: %w", err)
	}
	return nil
}

func (s *Store) Projects(ctx context.Context) ([]Project, error) {
	if s.config.orgID == "" {
		return nil, ErrNoOrgSelected
	}
	projects, err := batchSelect(
		s.db,
		"projects",
		[]string{"id", "name", "color"},
		func(p *Project) []any { return []any{&p.ID, &p.Name, &p.Color} },
		fmt.Sprintf("WHERE org_id = '%s'", s.config.orgID),
	)
	if err != nil {
		return nil, fmt.Errorf("couldn't select projects: %w", err)
	}
	return projects, nil
}

func (s *Store) StoreProjects(ctx context.Context, projects []Project) error {
	if s.config.orgID == "" {
		return ErrNoOrgSelected
	}
	err := batchInsert(
		s.db,
		"projects",
		[]string{"id", "name", "color", "org_id"},
		projects,
		func(p *Project) []any { return []any{p.ID, p.Name, p.Color, s.config.orgID} },
		"ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, color = EXCLUDED.color",
	)
	if err != nil {
		return fmt.Errorf("couldn't store projects: %w", err)
	}
	return nil
}

func (s *Store) Issues(ctx context.Context) ([]Issue, error) {
	if s.config.orgID == "" {
		return nil, ErrNoOrgSelected
	}
	issues, err := batchSelect(
		s.db,
		"issues",
		[]string{
			"id", "identifier", "title", "priority",
			"state", "description", "labels",
			"user_id", "team_id", "project_id",
		},
		func(i *Issue) []any {
			return []any{
				&i.ID, &i.Identifier, &i.Title, &i.Priority,
				&i.State, &i.Desc, &i.Labels,
			}
		},
		fmt.Sprintf("WHERE org_id = '%s'", s.config.orgID),
	)
	if err != nil {
		return nil, fmt.Errorf("couldn't select projects: %w", err)
	}
	return issues, nil
}

func (s *Store) StoreIssues(ctx context.Context, issues []Issue) error {
	if s.config.orgID == "" {
		return ErrNoOrgSelected
	}
	err := batchInsert(
		s.db,
		"issues",
		[]string{"id", "name", "color", "org_id"},
		issues,
		func(t *Issue) []any { return []any{} },
		"ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, color = EXCLUDED.color",
	)
	if err != nil {
		return fmt.Errorf("couldn't store projects: %w", err)
	}
	return nil
}

func (s *Store) SwitchOrg(ctx context.Context, orgID string) error {
	if orgID == "" {
		return errors.New("orgID is empty")
	}

	s.config.orgID = orgID

	return s.writeConfig(configLastOrgID, orgID)
}

func (s *Store) Close() error {
	return s.db.Close()
}
