package store

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	_ "github.com/glebarez/go-sqlite"
	"github.com/jmoiron/sqlx"
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

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

var ErrNoOrgSelected = errors.New("no active org")

var spaceRemoveRegEx = regexp.MustCompile("( |\t){2,}")

func removeDuplicatesAndEmpties[T idGetter](list []T) []T {
	m := make(map[string]T)
	for _, item := range list {
		m[item.getID()] = item
	}
	var res []T
	for k, v := range m {
		if k == "" {
			continue
		}
		res = append(res, v)
	}
	return res
}

func getEmptyProjectID(orgID string) string {
	return fmt.Sprintf("empty-project-%s", orgID[:6])
}

type StoreState struct {
	Search    string
	Project   *Project
	Org       Org
	Me        User
	FirstTime bool
}

type Store struct {
	db      *sqlx.DB
	current StoreState
}

func New(path string) (*Store, error) {
	db, err := sqlx.Open("sqlite", fmt.Sprintf("%s?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)", path))
	if err != nil {
		return nil, fmt.Errorf("couldn't instantiate db: %w", err)
	}

	// to avoid locks
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	db.MapperFunc(func(structFieldName string) string {
		snake := matchFirstCap.ReplaceAllString(structFieldName, "${1}_${2}")
		snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
		return strings.ToLower(snake)
	})

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

	store.current.FirstTime = store.current.Org.ID == "" || store.current.Me.ID == ""

	return store, nil
}

func (s *Store) Current() StoreState {
	return s.current
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Orgs() ([]Org, error) {
	var orgs []Org

	err := s.db.Select(&orgs, "SELECT * FROM orgs")
	if err != nil {
		return nil, fmt.Errorf("couldn't select orgs: %w", err)
	}

	return orgs, nil
}

func (s *Store) StoreOrg(org Org) (bool, error) {
	changed := s.current.Org.ID != org.ID
	if changed {
		_, err := s.db.Exec(`
			UPDATE orgs SET active = FALSE WHERE active = TRUE;

			INSERT INTO orgs (id, name, url_key, active) 
			VALUES (?, ?, ?, TRUE)
			ON CONFLICT (id) DO UPDATE
			SET name = EXCLUDED.name, url_key = EXCLUDED.url_key, active = TRUE;
		`, org.ID, org.Name, org.URLKey)
		if err != nil {
			return false, fmt.Errorf("couldn't update org: %w", err)
		}

		err = s.db.Get(&s.current.Org, "SELECT * FROM orgs WHERE active = TRUE;")
		if err != nil {
			return false, fmt.Errorf("couldn't select active org: %w", err)
		}

		_, err = s.db.Exec(`
			INSERT INTO projects (id, name, color, org_id)
			VALUES (?, '(No Project)', '#777', (SELECT id FROM orgs WHERE active = TRUE))
			ON CONFLICT (id) DO NOTHING
		`, getEmptyProjectID(org.ID))
		if err != nil {
			return false, fmt.Errorf("couldn't insert empty project for org: %w", err)
		}
	}
	return changed, nil
}

func (s *Store) Users() ([]User, error) {
	if s.current.Org.ID == "" {
		return nil, ErrNoOrgSelected
	}

	var users []User
	err := s.db.Select(&users, fmt.Sprintf(`
		SELECT id, name, display_name, email, is_me 
		FROM users 
		WHERE org_id = %s`,
		currentOrg,
	))
	if err != nil {
		return nil, fmt.Errorf("couldn't select users: %w", err)
	}

	return users, nil
}

func (s *Store) StoreUsers(users []User) error {
	if s.current.Org.ID == "" {
		return ErrNoOrgSelected
	}

	users = removeDuplicatesAndEmpties(users)
	if len(users) == 0 {
		return nil
	}

	_, err := s.db.NamedExec(fmt.Sprintf(`
		INSERT INTO users (id, name, display_name, email, is_me, org_id) 
		VALUES (:id, :name, :display_name, :email, :is_me, %s)
		ON CONFLICT (id) DO UPDATE 
		SET name = EXCLUDED.name,
			display_name = EXCLUDED.display_name,
			email = EXCLUDED.display_name,
			is_me = EXCLUDED.is_me
		`, currentOrg),
		users,
	)
	if err != nil {
		return fmt.Errorf("couldn't store users: %w", err)
	}

	for _, user := range users {
		if user.IsMe {
			s.current.Me = user
		}
	}

	return nil
}

func (s *Store) States() ([]State, error) {
	if s.current.Org.ID == "" {
		return nil, ErrNoOrgSelected
	}

	var states []State
	err := s.db.Select(&states, fmt.Sprintf(`
		SELECT id, name, color, team_id 
		FROM states 
		WHERE org_id = %s`, currentOrg),
	)
	if err != nil {
		return nil, fmt.Errorf("couldn't select states: %w", err)
	}

	return states, nil
}

func (s *Store) StoreStates(states []State) error {
	if s.current.Org.ID == "" {
		return ErrNoOrgSelected
	}

	states = removeDuplicatesAndEmpties(states)
	if len(states) == 0 {
		return nil
	}

	_, err := s.db.NamedExec(fmt.Sprintf(`
		INSERT INTO states (id, name, color, team_id, org_id)
		VALUES (:id, :name, :color, :team_id, %s)
		ON CONFLICT (id) DO UPDATE 
		SET name = EXCLUDED.name, 
			color = EXCLUDED.color
		`, currentOrg),
		states,
	)
	if err != nil {
		return fmt.Errorf("couldn't store states: %w", err)
	}

	return nil
}

func (s *Store) Teams() ([]Team, error) {
	if s.current.Org.ID == "" {
		return nil, ErrNoOrgSelected
	}

	var teams []Team
	err := s.db.Select(&teams, fmt.Sprintf(`
		SELECT id, name, color 
		FROM teams
		WHERE org_id = %s`, currentOrg),
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

	if len(teams) == 0 {
		return nil
	}

	_, err := s.db.NamedExec(fmt.Sprintf(`
		INSERT INTO teams (id, name, color, org_id)
		VALUES (:id, :name, :color, %s)
		ON CONFLICT (id) DO UPDATE 
		SET name = EXCLUDED.name, 
			color = EXCLUDED.color
		`, currentOrg),
		removeDuplicatesAndEmpties(teams),
	)
	if err != nil {
		return fmt.Errorf("couldn't store teams: %w", err)
	}

	return nil
}

func (s *Store) Projects() ([]Project, error) {
	if s.current.Org.ID == "" {
		return nil, ErrNoOrgSelected
	}

	var projects []Project
	err := s.db.Select(&projects, fmt.Sprintf(`
		SELECT id, name, color 
		FROM projects
		WHERE org_id = %s
		ORDER BY name`, currentOrg),
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

	projects = removeDuplicatesAndEmpties(projects)
	if len(projects) == 0 {
		return nil
	}

	_, err := s.db.NamedExec(fmt.Sprintf(`
		INSERT INTO projects (id, name, color, org_id)
		VALUES (:id, :name, :color, %s)
		ON CONFLICT (id) DO UPDATE 
		SET name = EXCLUDED.name, 
			color = EXCLUDED.color
		`, currentOrg), projects,
	)
	if err != nil {
		return fmt.Errorf("couldn't store projects: %w", err)
	}

	return nil
}

func (s *Store) Issue(issueID string) (*Issue, error) {
	if s.current.Org.ID == "" {
		return nil, ErrNoOrgSelected
	}

	var issue Issue
	err := s.db.
		QueryRowx(fmt.Sprintf(`
			SELECT 
				issues.id, identifier, title,
				priority, description, labels, 
				pinned, created_at, updated_at, canceled_at,
				states.id AS "state.id",
				states.name AS "state.name",
				states.color AS "state.color",
				teams.id AS "team.id",
				teams.name AS "team.name",
				teams.color AS "team.color",
				COALESCE(projects.id, '') AS "project.id",
				COALESCE(projects.name, '') AS "project.name",
				COALESCE(projects.color, '') AS "project.color",
				COALESCE(users.id, '') AS "assignee.id",
				COALESCE(users.name, '') AS "assignee.name",
				COALESCE(users.display_name, '') AS "assignee.display_name",
				COALESCE(users.email, '') AS "assignee.email",
				COALESCE(users.is_me, FALSE) AS "assignee.is_me"
			FROM issues
			LEFT JOIN users ON issues.assignee_id = users.id
			LEFT JOIN projects ON issues.project_id = projects.id
			LEFT JOIN teams ON issues.team_id = teams.id
			LEFT JOIN states ON issues.state_id = states.id
			WHERE issues.id = ? AND issues.org_id = %s
		`, currentOrg), issueID).
		StructScan(&issue)
	if err != nil {
		return nil, fmt.Errorf("failed to scan one issue: %w", err)
	}

	return &issue, nil
}

func (s *Store) Issues(issueIDs ...string) ([]Issue, error) {
	if s.current.Org.ID == "" {
		return nil, ErrNoOrgSelected
	}

	issueFilterQuery, args, err := s.getIssueFilter(issueIDs...)
	if err != nil {
		return nil, fmt.Errorf("failed to generate issue filter query: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT 
			issues.id, identifier, title,
			priority, description, labels, 
			pinned, created_at, updated_at, canceled_at,
			states.id AS "state.id",
			states.name AS "state.name",
			states.color AS "state.color",
			states.team_id AS "state.team_id",
			teams.id AS "team.id",
			teams.name AS "team.name",
			teams.color AS "team.color",
			COALESCE(projects.id, '') AS "project.id",
			COALESCE(projects.name, '') AS "project.name",
			COALESCE(projects.color, '') AS "project.color",
			COALESCE(users.id, '') AS "assignee.id",
			COALESCE(users.name, '') AS "assignee.name",
			COALESCE(users.display_name, '') AS "assignee.display_name",
			COALESCE(users.email, '') AS "assignee.email",
			COALESCE(users.is_me, FALSE) AS "assignee.is_me"
		FROM issues
		INNER JOIN orgs ON issues.org_id = orgs.id
		LEFT JOIN users ON issues.assignee_id = users.id
		LEFT JOIN projects ON issues.project_id = projects.id
		LEFT JOIN teams ON issues.team_id = teams.id
		LEFT JOIN states ON issues.state_id = states.id
		WHERE %s %s orgs.active = TRUE AND (
			states.name NOT IN ('Done', 'Canceled') OR 
			updated_at > DATETIME(CURRENT_TIMESTAMP, '-14 days')
		)
		ORDER BY pinned = TRUE DESC, 
			%s
	`, issueFilterQuery, s.getProjectFilter(), s.getSorter(false))

	var issues []Issue
	rows, err := s.db.Queryx(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query issues: %w", err)
	}

	for rows.Next() {
		var issue Issue
		err := rows.StructScan(&issue)
		if err != nil {
			return nil, fmt.Errorf("failed to scan issue: %w", err)
		}
		issues = append(issues, issue)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("scanning issues had error: %w", rows.Err())
	}

	return issues, nil
}

func (s *Store) SearchIssues(search string) ([]Issue, error) {
	if s.current.Org.ID == "" {
		return nil, ErrNoOrgSelected
	}

	var searchArg string
	quoteRemoved := strings.ReplaceAll(search, "\"", " ")
	keywordSplit := strings.Split(quoteRemoved, " ")
	for _, keyword := range keywordSplit {
		searchArg += fmt.Sprintf(`"%s" `, keyword)
	}

	query := fmt.Sprintf(`
		SELECT 
			issues.id, 
			issues.identifier, 
			issues.title,
			priority, 
			issues.description, 
			issues.labels, 
			pinned, created_at, 
			updated_at, canceled_at,
			states.id AS "state.id",
			states.name AS "state.name",
			states.color AS "state.color",
			teams.id AS "team.id",
			teams.name AS "team.name",
			teams.color AS "team.color",
			COALESCE(projects.id, '') AS "project.id",
			COALESCE(projects.name, '') AS "project.name",
			COALESCE(projects.color, '') AS "project.color",
			COALESCE(users.id, '') AS "assignee.id",
			COALESCE(users.name, '') AS "assignee.name",
			COALESCE(users.display_name, '') AS "assignee.display_name",
			COALESCE(users.email, '') AS "assignee.email",
			COALESCE(users.is_me, FALSE) AS "assignee.is_me"
		FROM issues
		INNER JOIN orgs ON issues.org_id = orgs.id
		INNER JOIN search ON issues.id = search.id
		LEFT JOIN users ON issues.assignee_id = users.id
		LEFT JOIN projects ON issues.project_id = projects.id
		LEFT JOIN teams ON issues.team_id = teams.id
		LEFT JOIN states ON issues.state_id = states.id
		WHERE %s orgs.active = TRUE AND search MATCH ? AND (
			states.name NOT IN ('Done', 'Canceled') OR 
			updated_at > DATETIME(CURRENT_TIMESTAMP, '-14 days')
		)
		ORDER BY pinned = TRUE DESC, 
			%s
	`, s.getProjectFilter(), s.getSorter(true))

	var issues []Issue
	rows, err := s.db.Queryx(query, searchArg)
	if err != nil {
		return nil, fmt.Errorf("failed to query searched issues: %w", err)
	}

	for rows.Next() {
		var issue Issue
		err := rows.StructScan(&issue)
		if err != nil {
			return nil, fmt.Errorf("failed to scan issue: %w", err)
		}
		issues = append(issues, issue)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("scanning issues had error: %w", rows.Err())
	}

	return issues, nil
}

func (s *Store) StoreIssues(issues []Issue) error {
	if s.current.Org.ID == "" {
		return nil
	}

	if len(issues) == 0 {
		return nil
	}

	var teams []Team
	var projects []Project
	var users []User
	var states []State

	for _, issue := range issues {
		teams = append(teams, issue.Team)
		states = append(states, issue.State)
		projects = append(projects, issue.Project)
		users = append(users, issue.Assignee)
	}

	err := s.StoreTeams(teams)
	if err != nil {
		return fmt.Errorf("failed to store teams: %w", err)
	}
	err = s.StoreProjects(projects)
	if err != nil {
		return fmt.Errorf("failed to store projects: %w", err)
	}
	err = s.StoreStates(states)
	if err != nil {
		return fmt.Errorf("failed to store states: %w", err)
	}
	err = s.StoreUsers(users)
	if err != nil {
		return fmt.Errorf("failed to store users: %w", err)
	}

	type issueModel struct {
		ID          string
		Identifier  string
		Title       string
		Description string
		Labels      Label
		Priority    Prio
		TeamID      string
		StateID     string
		AssigneeID  sql.Null[string]
		ProjectID   string
		Pinned      bool
		CreatedAt   time.Time
		UpdatedAt   time.Time
		CanceledAt  *time.Time
	}

	var issueModels []issueModel
	for _, issue := range issues {
		projectID := issue.Project.ID
		if projectID == "" {
			projectID = getEmptyProjectID(s.current.Org.ID)
		}

		issueModels = append(issueModels, issueModel{
			ID:          issue.ID,
			Identifier:  issue.Identifier,
			Title:       issue.Title,
			Description: issue.Description,
			Labels:      issue.Labels,
			Priority:    issue.Priority,
			TeamID:      issue.Team.ID,
			StateID:     issue.State.ID,
			ProjectID:   projectID,
			AssigneeID: sql.Null[string]{
				Valid: issue.Assignee.ID != "",
				V:     issue.Assignee.ID,
			},
			Pinned:     issue.Pinned,
			CreatedAt:  issue.CreatedAt,
			UpdatedAt:  issue.UpdatedAt,
			CanceledAt: issue.CanceledAt,
		})
	}

	_, err = s.db.NamedExec(fmt.Sprintf(`
		INSERT INTO issues (
			id, identifier, title, 
			description, labels, priority, 
			team_id, state_id, assignee_id, 
			project_id, pinned, created_at, 
			updated_at, canceled_at, org_id
		)
		VALUES (
			:id, :identifier, :title, 
			:description, :labels, :priority, 
			:team_id, :state_id, :assignee_id, 
			:project_id, :pinned, :created_at, 
			:updated_at, :canceled_at, %s
		)
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
		`, currentOrg),
		issueModels,
	)
	if err != nil {
		return fmt.Errorf("couldn't store issues: %w", err)
	}

	return nil
}

func (s *Store) SetBookmark(issueIDs ...string) error {
	if len(issueIDs) == 0 {
		return nil
	}

	query, args, err := sqlx.In(`
		UPDATE issues 
		SET pinned = NOT pinned 
		WHERE id IN (?)`,
		issueIDs,
	)
	if err != nil {
		return fmt.Errorf("couldn't generate toggle issue pin: %w", err)
	}

	_, err = s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("couldn't toggle issue pin: %w", err)
	}

	return nil
}

type UpdateIssueField string

const (
	UpdateIssueFieldAssignee UpdateIssueField = "assignee_id"
	UpdateIssueFieldPrio     UpdateIssueField = "priority"
	UpdateIssueFieldProject  UpdateIssueField = "project_id"
	UpdateIssueFieldTeam     UpdateIssueField = "team_id"
	UpdateIssueFieldTitle    UpdateIssueField = "title"
	UpdateIssueFieldState    UpdateIssueField = "state_id"
)

func (s *Store) UpdateIssues(field UpdateIssueField, value any, issueIDs ...string) error {
	if len(issueIDs) == 0 {
		return nil
	}

	if len(field) == 0 {
		return nil
	}

	query, args, err := sqlx.In(
		fmt.Sprintf(`UPDATE issues SET %s = ?, updated_at = ? WHERE id IN (?)`, field),
		value,
		time.Now(),
		issueIDs,
	)
	if err != nil {
		return fmt.Errorf("couldn't generate set assignee query: %w", err)
	}

	_, err = s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("couldn't set assignees: %w", err)
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
	s.current.Org.SyncedAt = time.Now()

	return nil
}

func (s *Store) SetSortMode(mode SortMode) error {
	if s.current.Org.SortMode == mode {
		if s.current.Org.SortOrder == sortOrderAsc {
			s.current.Org.SortOrder = sortOrderDesc
		} else {
			s.current.Org.SortOrder = sortOrderAsc
		}
	} else {
		s.current.Org.SortMode = mode
	}

	_, err := s.db.Exec(`
		UPDATE orgs 
		SET sort_mode = ?, 
			sort_order = ? 
		WHERE orgs.active = TRUE;`,
		s.current.Org.SortMode,
		s.current.Org.SortOrder,
	)
	if err != nil {
		return fmt.Errorf("failed to save sort settings: %w", err)
	}
	return nil
}

func (s *Store) SetProject(project *Project) {
	s.current.Project = project
}

func (s *Store) loadCurrentState() error {
	err := s.db.Get(&s.current.Org, `SELECT * FROM orgs WHERE active = TRUE`)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("couldn't get org: %w", err)
	}

	err = s.db.Get(&s.current.Me, `
			SELECT users.id, users.name, display_name, email, is_me
			FROM users
			JOIN orgs ON orgs.id = users.org_id
			WHERE orgs.active = TRUE AND is_me = TRUE
		`)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("couldn't get me: %w", err)
	}

	return nil
}

func (s *Store) getSorter(includeRank bool) string {
	orderStr := "DESC"

	if s.current.Org.SortOrder == sortOrderAsc {
		orderStr = "ASC"
	}

	rank := "rank, "
	if !includeRank {
		rank = ""
	}

	switch s.current.Org.SortMode {
	case SortModeProject:
		return fmt.Sprintf("projects.name %s", orderStr)
	case SortModeTitle:
		return fmt.Sprintf("issues.title %s", orderStr)
	case SortModeAssignee:
		return fmt.Sprintf("users.display_name %s", orderStr)
	case SortModeState:
		return fmt.Sprintf("states.name %s", orderStr)
	case SortModePrio:
		return fmt.Sprintf("issues.priority != 0 DESC, issues.priority %s", orderStr)
	case SortModeAge:
		return fmt.Sprintf("created_at %s", orderStr)
	case SortModeTeam:
		return fmt.Sprintf("teams.name %s", orderStr)
	default:
		return rank + `
			(states.name = 'Done' OR states.name = 'Canceled') ASC,
			users.is_me DESC,
			states.name = 'In Progress' DESC,
			issues.priority = 0 ASC,
			states.name = 'Todo' DESC,
			states.name = 'Backlog' DESC,
			assignee_id = '' DESC,
			issues.priority ASC,
			created_at DESC
		`
	}
}

func (s *Store) getProjectFilter() string {
	if s.current.Project == nil {
		return ""
	}
	return fmt.Sprintf("issues.project_id = '%s' AND", s.current.Project.ID)
}

func (s *Store) getIssueFilter(issueIDs ...string) (string, []any, error) {
	if len(issueIDs) == 0 {
		return "TRUE AND", nil, nil
	}
	return sqlx.In("issues.id IN (?) AND", issueIDs)
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
		FROM issues
		INNER JOIN orgs ON issues.org_id = orgs.id
		LEFT JOIN users ON issues.assignee_id = users.id
		LEFT JOIN projects ON issues.project_id = projects.id
		LEFT JOIN teams ON issues.team_id = teams.id
		LEFT JOIN states ON issues.state_id = states.id
		WHERE orgs.active = TRUE AND
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
