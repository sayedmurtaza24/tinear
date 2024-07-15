package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/glebarez/go-sqlite"
)

type sortOrder int
type SortMode int

const (
	sortOrderAsc sortOrder = iota
	sortOrderDesc
)

const (
	SortModeSmart SortMode = iota
	SortModeProject
	SortModeTitle
	SortModeAssignee
	SortModeState
	SortModePrio
	SortModeAge
	SortModeTeam
)

const currentOrg = "(SELECT id FROM orgs WHERE active = TRUE)"

type StoreState struct {
	FirstTime bool

	SortMode  SortMode
	SortOrder sortOrder

	Search string

	Org      Org
	Me       User
	SyncedAt time.Time
}

type Store struct {
	db      *sql.DB
	current StoreState
}

func New(path string) (*Store, error) {
	db, err := sql.Open("sqlite", fmt.Sprintf("%s?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)", path))
	if err != nil {
		return nil, fmt.Errorf("couldn't instantiate db: %w", err)
	}

	// to avoid locks
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("couldn't ping db: %w", err)
	}

	store := &Store{db: db}

	err = store.migrate()
	if err != nil {
		return nil, fmt.Errorf("couldn't migrate db: %w", err)
	}

	err = store.loadCurrentState()
	if err != nil {
		return nil, fmt.Errorf("couldn't load current current state: %w", err)
	}

	return store, nil
}

func (s *Store) Current() StoreState { return s.current }
func (s *Store) Close() error        { return s.db.Close() }

func (s *Store) Orgs() ([]Org, error) {
	orgs, err := batchSelect(
		s.db,
		"orgs",
		[]string{"id", "name", "url_key"},
		func(o *Org) []any { return []any{&o.ID, &o.Name, &o.URLKey} },
		"",
	)
	if err != nil {
		return nil, fmt.Errorf("couldn't select orgs: %w", err)
	}
	return orgs, nil
}

func (s *Store) StoreOrg(org Org) (changed bool, err error) {
	tx, err := s.db.Begin()
	if err != nil {
		return false, fmt.Errorf("failed to store org: %w", err)
	}
	defer tx.Rollback()

	var id string
	err = tx.QueryRow("SELECT id FROM orgs WHERE active = TRUE;").Scan(&id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, fmt.Errorf("couldn't store org: %w", err)
	}

	if id != org.ID {
		if id != "" {
			_, err := tx.Exec("UPDATE orgs SET active = FALSE WHERE active = TRUE")
			if err != nil {
				return false, fmt.Errorf("couldn't update previous org: %w", err)
			}
		}

		_, err := tx.Exec(`
			INSERT INTO orgs (id, name, url_key, active) 
			VALUES (?, ?, ?, ?)
			ON CONFLICT (id) DO UPDATE
			SET name = EXCLUDED.name, url_key = EXCLUDED.url_key, active = TRUE;
		`, org.ID, org.Name, org.URLKey, true)
		if err != nil {
			return false, fmt.Errorf("couldn't update org: %w", err)
		}

		changed = true
	}

	err = tx.
		QueryRow("SELECT id, name, url_key, synced_at FROM orgs WHERE active = TRUE;").
		Scan(&s.current.Org.ID, &s.current.Org.Name, &s.current.Org.URLKey, &s.current.SyncedAt)
	if err != nil {
		return false, fmt.Errorf("couldn't select active org: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("couldn't commit store org tx: %w", err)
	}

	return changed, nil
}

func (s *Store) Users() ([]User, error) {
	if s.current.Org.ID == "" {
		return make([]User, 0), nil
	}
	users, err := batchSelect(
		s.db,
		"users",
		[]string{"id", "name", "display_name", "email", "is_me"},
		func(o *User) []any { return []any{&o.ID, &o.Name, &o.DisplayName, &o.Email, &o.IsMe} },
		fmt.Sprintf("WHERE org_id = '%s'", currentOrg),
	)
	if err != nil {
		return nil, fmt.Errorf("couldn't select users: %w", err)
	}
	return users, nil
}

func (s *Store) StoreUsers(users []User) error {
	if s.current.Org.ID == "" {
		return nil
	}
	var me *User
	err := batchInsert(
		s.db,
		"users",
		[]string{"id", "name", "display_name", "email", "is_me", "org_id"},
		users,
		func(u *User) []any {
			if u.IsMe {
				me = u
			}
			return []any{u.ID, u.Name, u.DisplayName, u.Email, u.IsMe}
		},
		`ON CONFLICT (id) DO UPDATE 
		 SET name = EXCLUDED.name, 
			display_name = EXCLUDED.display_name, 
			email = EXCLUDED.email,
			is_me = EXCLUDED.is_me,
			org_id = EXCLUDED.org_id;
		`,
	)
	if err != nil {
		return fmt.Errorf("couldn't store users: %w", err)
	}
	if me != nil {
		s.current.Me = *me
	}
	return nil
}

func (s *Store) States() ([]State, error) {
	if s.current.Org.ID == "" {
		return make([]State, 0), nil
	}
	states, err := batchSelect(
		s.db,
		"states",
		[]string{"id", "name", "color"},
		func(t *State) []any { return []any{&t.ID, &t.Name, &t.Color} },
		fmt.Sprintf("WHERE org_id = '%s'", currentOrg),
	)
	if err != nil {
		return nil, fmt.Errorf("couldn't select states: %w", err)
	}
	return states, nil
}

func (s *Store) StoreStates(states []State) error {
	if s.current.Org.ID == "" {
		return nil
	}
	err := batchInsert(
		s.db,
		"states",
		[]string{"id", "name", "color", "org_id"},
		states,
		func(t *State) []any { return []any{t.ID, t.Name, t.Color} },
		"ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, color = EXCLUDED.color",
	)
	if err != nil {
		return fmt.Errorf("couldn't store states: %w", err)
	}
	return nil
}

func (s *Store) Teams() ([]Team, error) {
	if s.current.Org.ID == "" {
		return make([]Team, 0), nil
	}
	teams, err := batchSelect(
		s.db,
		"teams",
		[]string{"id", "name", "color"},
		func(t *Team) []any { return []any{&t.ID, &t.Name, &t.Color} },
		fmt.Sprintf("WHERE org_id = '%s'", currentOrg),
	)
	if err != nil {
		return nil, fmt.Errorf("couldn't select teams: %w", err)
	}
	return teams, nil
}

func (s *Store) StoreTeams(teams []Team) error {
	if s.current.Org.ID == "" {
		return nil
	}
	err := batchInsert(
		s.db,
		"teams",
		[]string{"id", "name", "color", "org_id"},
		teams,
		func(t *Team) []any { return []any{t.ID, t.Name, t.Color} },
		"ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, color = EXCLUDED.color",
	)
	if err != nil {
		return fmt.Errorf("couldn't store teams: %w", err)
	}
	return nil
}

func (s *Store) Projects() ([]Project, error) {
	if s.current.Org.ID == "" {
		return make([]Project, 0), nil
	}
	projects, err := batchSelect(
		s.db,
		"projects",
		[]string{"id", "name", "color"},
		func(p *Project) []any { return []any{&p.ID, &p.Name, &p.Color} },
		fmt.Sprintf("WHERE org_id = '%s'", currentOrg),
	)
	if err != nil {
		return nil, fmt.Errorf("couldn't select projects: %w", err)
	}
	return projects, nil
}

func (s *Store) StoreProjects(projects []Project) error {
	if s.current.Org.ID == "" {
		return nil
	}
	err := batchInsert(
		s.db,
		"projects",
		[]string{"id", "name", "color", "org_id"},
		projects,
		func(p *Project) []any { return []any{p.ID, p.Name, p.Color} },
		"ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, color = EXCLUDED.color",
	)
	if err != nil {
		return fmt.Errorf("couldn't store projects: %w", err)
	}
	return nil
}

func (s *Store) Issue(issueID string) (*Issue, error) {
	if s.current.Org.ID == "" {
		return nil, nil
	}

	var issue Issue
	res := s.db.
		QueryRow(`
			SELECT issues.id, identifier, title,
				priority, description, labels,
				states.id, states.name, states.color,
				projects.id, projects.name, projects.color,
				teams.id, teams.name, teams.color,
				users.id, users.name, users.display_name,
				users.email, users.is_me, pinned,
				created_at, updated_at, canceled_at
			FROM 
				issues, users, projects, teams, states
			WHERE
				issues.assignee_id = users.id AND
				issues.project_id = projects.id AND
				issues.team_id = teams.id AND
				issues.state_id = states.id AND
				issues.id = ?
		`, issueID)
	if res.Err() != nil {
		return nil, fmt.Errorf("failed to query one issue: %w", res.Err())
	}

	err := res.Scan(
		&issue.ID, &issue.Identifier, &issue.Title,
		&issue.Priority, &issue.Desc, &issue.Labels,
		&issue.State.ID, &issue.State.Name, &issue.State.Color,
		&issue.Project.ID, &issue.Project.Name, &issue.Project.Color,
		&issue.Team.ID, &issue.Team.Name, &issue.Team.Color,
		&issue.Assignee.ID, &issue.Assignee.Name,
		&issue.Assignee.DisplayName, &issue.Assignee.Email,
		&issue.Assignee.IsMe, &issue.Pinned,
		&issue.CreatedAt, &issue.UpdatedAt, &issue.CanceledAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("failed to scan one issue: %w", err)
		}
	}

	return &issue, nil
}

func (s *Store) Issues() ([]Issue, error) {
	if s.current.Org.ID == "" {
		return make([]Issue, 0), nil
	}
	issues, err := batchSelect(
		s.db,
		"issues, users, projects, teams, states, orgs",
		[]string{
			"issues.id", "identifier", "title",
			"priority", "description", "labels",
			"states.id", "states.name", "states.color",
			"projects.id", "projects.name", "projects.color",
			"teams.id", "teams.name", "teams.color",
			"users.id", "users.name", "users.display_name",
			"users.email", "users.is_me", "pinned",
			"created_at", "updated_at", "canceled_at",
		},
		func(i *Issue) []any {
			return []any{
				&i.ID, &i.Identifier, &i.Title,
				&i.Priority, &i.Desc, &i.Labels,
				&i.State.ID, &i.State.Name, &i.State.Color,
				&i.Project.ID, &i.Project.Name, &i.Project.Color,
				&i.Team.ID, &i.Team.Name, &i.Team.Color,
				&i.Assignee.ID, &i.Assignee.Name,
				&i.Assignee.DisplayName, &i.Assignee.Email,
				&i.Assignee.IsMe, &i.Pinned,
				&i.CreatedAt, &i.UpdatedAt, &i.CanceledAt,
			}
		},
		fmt.Sprintf(`
			WHERE (states.name NOT IN ('Done', 'Canceled') OR 
				updated_at > DATETIME(CURRENT_TIMESTAMP, '-14 days')) AND
				issues.assignee_id = users.id AND
				issues.project_id = projects.id AND
				issues.team_id = teams.id AND
				issues.state_id = states.id AND
				issues.org_id = orgs.id AND
				orgs.active = TRUE
			ORDER BY pinned = TRUE DESC, 
				%s`,
			s.getSorter(false),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("couldn't select issues: %w", err)
	}
	return issues, nil
}

func (s *Store) SearchIssues(search string) ([]Issue, error) {
	if s.current.Org.ID == "" {
		return make([]Issue, 0), nil
	}

	var searchArg string
	quoteRemoved := strings.ReplaceAll(search, "\"", " ")
	keywordSplit := strings.Split(quoteRemoved, " ")
	for _, keyword := range keywordSplit {
		searchArg += fmt.Sprintf(`"%s" `, keyword)
	}

	issues, err := batchSelect(
		s.db,
		"search, issues, users, projects, teams, states, orgs",
		[]string{
			"issues.id", "identifier", "issues.title",
			"priority", "issues.description", "issues.labels",
			"states.id", "states.name", "states.color",
			"projects.id", "projects.name", "projects.color",
			"teams.id", "teams.name", "teams.color",
			"users.id", "users.name", "users.display_name",
			"users.email", "users.is_me", "pinned",
			"created_at", "updated_at", "canceled_at",
		},
		func(i *Issue) []any {
			return []any{
				&i.ID, &i.Identifier, &i.Title,
				&i.Priority, &i.Desc, &i.Labels,
				&i.State.ID, &i.State.Name, &i.State.Color,
				&i.Project.ID, &i.Project.Name, &i.Project.Color,
				&i.Team.ID, &i.Team.Name, &i.Team.Color,
				&i.Assignee.ID, &i.Assignee.Name,
				&i.Assignee.DisplayName, &i.Assignee.Email,
				&i.Assignee.IsMe, &i.Pinned,
				&i.CreatedAt, &i.UpdatedAt, &i.CanceledAt,
			}
		},
		fmt.Sprintf(` 
			WHERE (states.name NOT IN ('Done', 'Canceled') OR 
				updated_at > DATETIME(CURRENT_TIMESTAMP, '-14 days')) AND
				search MATCH ? AND
				issues.id = search.id AND
				issues.assignee_id = users.id AND
				issues.project_id = projects.id AND
				issues.team_id = teams.id AND
				issues.state_id = states.id AND
				issues.org_id = orgs.id AND
				orgs.active = TRUE
			ORDER BY pinned = TRUE DESC, 
				%s`,
			s.getSorter(true),
		),
		searchArg,
	)
	if err != nil {
		return nil, fmt.Errorf("couldn't select searched issues: %w", err)
	}
	return issues, nil
}

func (s *Store) StoreIssues(issues []Issue) error {
	if s.current.Org.ID == "" {
		return nil
	}

	var teams []Team
	var projects []Project
	var users []User
	var states []State

	for _, issue := range issues {
		teams = append(teams, issue.Team)
		projects = append(projects, issue.Project)
		users = append(users, issue.Assignee)
		states = append(states, issue.State)
	}

	err := s.StoreTeams(teams)
	if err != nil {
		return fmt.Errorf("failed to store teams: %w", err)
	}
	err = s.StoreProjects(projects)
	if err != nil {
		return fmt.Errorf("failed to store teams: %w", err)
	}
	err = s.StoreStates(states)
	if err != nil {
		return fmt.Errorf("failed to store teams: %w", err)
	}
	err = s.StoreUsers(users)
	if err != nil {
		return fmt.Errorf("failed to store teams: %w", err)
	}

	err = batchInsert(
		s.db,
		"issues",
		[]string{
			"id", "identifier", "title",
			"priority", "description", "labels",
			"state_id", "project_id", "team_id",
			"assignee_id", "created_at", "updated_at",
			"canceled_at", "org_id",
		},
		issues,
		func(i *Issue) []any {
			return []any{
				i.ID, i.Identifier, i.Title,
				i.Priority, i.Desc, i.Labels,
				i.State.ID, i.Project.ID, i.Team.ID,
				i.Assignee.ID, i.CreatedAt, i.UpdatedAt,
				i.CanceledAt,
			}
		}, `
		ON CONFLICT (id) DO UPDATE 
		SET identifier = EXCLUDED.identifier, 
			title = EXCLUDED.title,
			priority = EXCLUDED.priority,
			description = EXCLUDED.description,
			labels = EXCLUDED.labels,
			state_id = EXCLUDED.state_id,
			project_id = EXCLUDED.project_id,
			team_id = EXCLUDED.team_id,
			assignee_id = EXCLUDED.assignee_id,
			created_at = EXCLUDED.created_at,
			updated_at = EXCLUDED.updated_at,
			canceled_at = EXCLUDED.canceled_at
		`,
	)
	if err != nil {
		return fmt.Errorf("failed to store teams: %w", err)
	}

	s.current.FirstTime = false

	return nil
}

func (s *Store) ToggleBookmark(issueIDs ...string) error {
	if len(issueIDs) == 0 {
		return nil
	}

	var args []any
	for _, id := range issueIDs {
		args = append(args, id)
	}

	in := strings.Join(strings.Split(strings.Repeat("?", len(issueIDs)), ""), ", ")

	_, err := s.db.Exec(fmt.Sprintf("UPDATE issues SET pinned = NOT pinned WHERE id IN (%s)", in), args...)
	if err != nil {
		return fmt.Errorf("couldn't toggle issue pin: %w", err)
	}

	return nil
}

func (s *Store) Synced() error {
	err := s.updateSearchIndex()
	if err != nil {
		return fmt.Errorf("failed to update search indices: %w", err)
	}

	_, err = s.db.Exec(`
		UPDATE orgs
		SET synced_at = CURRENT_TIMESTAMP
		WHERE active = TRUE`,
	)
	if err != nil {
		return fmt.Errorf("failed to set org: %w", err)
	}
	s.current.SyncedAt = time.Now()

	return nil
}

func (s *Store) SetSortMode(mode SortMode) error {
	if s.current.SortMode == mode {
		if s.current.SortOrder == sortOrderAsc {
			s.current.SortOrder = sortOrderDesc
		} else {
			s.current.SortOrder = sortOrderAsc
		}
	} else {
		s.current.SortMode = mode
	}

	_, err := s.db.Exec(
		"UPDATE orgs SET sort_mode = ?, sort_order = ? WHERE orgs.active = TRUE;",
		s.current.SortMode,
		s.current.SortOrder,
	)
	if err != nil {
		return fmt.Errorf("failed to save sort settings: %w", err)
	}
	return nil
}

func (s *Store) loadCurrentState() error {
	// select org
	err := s.db.
		QueryRow(`SELECT id, name, url_key, synced_at FROM orgs WHERE active = TRUE`).
		Scan(
			&s.current.Org.ID,
			&s.current.Org.Name,
			&s.current.Org.URLKey,
			&s.current.SyncedAt,
		)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("couldn't get org: %w", err)
	}

	// select me
	err = s.db.
		QueryRow(`
			SELECT 
				users.id, 
				users.name,
				display_name,
				email,
				is_me,
				sort_mode,
				sort_order
			FROM users
			JOIN orgs ON orgs.id = users.org_id
			WHERE orgs.active = TRUE AND is_me = TRUE
		`).
		Scan(
			&s.current.Me.ID,
			&s.current.Me.Name,
			&s.current.Me.DisplayName,
			&s.current.Me.Email,
			&s.current.Me.IsMe,
			&s.current.SortMode,
			&s.current.SortOrder,
		)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("couldn't get me: %w", err)
	}

	if s.current.Me.ID == "" || s.current.Org.ID == "" {
		s.current.FirstTime = true
	}

	return nil
}

func (s *Store) getSorter(includeRank bool) string {
	orderStr := "DESC"

	if s.current.SortOrder == sortOrderAsc {
		orderStr = "ASC"
	}

	rank := "rank, "
	if !includeRank {
		rank = ""
	}

	switch s.current.SortMode {
	case SortModeProject:
		return fmt.Sprintf("projects.name %s;", orderStr)
	case SortModeTitle:
		return fmt.Sprintf("issues.title %s;", orderStr)
	case SortModeAssignee:
		return fmt.Sprintf("users.display_name %s;", orderStr)
	case SortModeState:
		return fmt.Sprintf("states.name %s;", orderStr)
	case SortModePrio:
		return fmt.Sprintf("issues.priority != 0 DESC, issues.priority %s;", orderStr)
	case SortModeAge:
		return fmt.Sprintf("created_at %s;", orderStr)
	case SortModeTeam:
		return fmt.Sprintf("teams.name %s;", orderStr)
	default:
		return rank + `
			(states.name = 'Done' OR states.name = 'Canceled') ASC,
			users.is_me DESC,
			states.name = 'In Progress' DESC,
			issues.priority = 0 ASC,
			issues.priority ASC,
			states.name = 'Todo' DESC,
			states.name = 'Backlog' DESC,
			assignee_id = '' DESC,
			created_at DESC;`
	}
}

func (s *Store) updateSearchIndex() error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin updating search indices: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		DELETE FROM search
		WHERE id IN (
			SELECT id
			FROM issues
			WHERE issues.org_id = (SELECT id FROM orgs WHERE orgs.active = TRUE) AND
				updated_at >= (SELECT synced_at FROM orgs WHERE active = TRUE) OR
				canceled_at >= (SELECT synced_at FROM orgs WHERE active = TRUE)
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to delete updated issues from search indices: %w", err)
	}

	_, err = tx.Exec(`
		INSERT INTO search (id, title, description, state, project, team, assignee, labels)
		SELECT issues.id, 
			title, 
			description, 
			states.name,
			projects.name, 
			teams.name, 
			users.name,
			(SELECT group_concat(value->>'$.Name', ' ') FROM json_each(issues.labels))
		FROM issues, projects, teams, users, states, orgs
		WHERE issues.project_id = projects.id AND
			issues.team_id = teams.id AND
			issues.assignee_id = users.id AND
			issues.org_id = orgs.id AND
			issues.state_id = states.id AND
			orgs.active = TRUE AND
			(
				updated_at >= orgs.synced_at OR
				canceled_at >= orgs.synced_at OR
				created_at >= orgs.synced_at
			);
	`)
	if err != nil {
		return fmt.Errorf("failed to insert updated issues into search indices: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit updating search indices: %w", err)
	}

	return nil
}
