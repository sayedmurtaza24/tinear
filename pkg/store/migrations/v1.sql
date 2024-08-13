CREATE TABLE migrations (
    version INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE orgs (
    id TEXT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    url_key TEXT NOT NULL,
    active bool NOT NULL,
    synced_at TIMESTAMP DEFAULT (DATETIME('NOW', '-6 months')),
    sort_order INTEGER DEFAULT 0,
    sort_mode INTEGER DEFAULT 0
);

CREATE UNIQUE INDEX unique_idx_orgs_active
ON orgs (active)
WHERE active = TRUE;

CREATE TABLE projects (
    id TEXT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    color TEXT NOT NULL,
    org_id TEXT NOT NULL,
    FOREIGN KEY (org_id) REFERENCES orgs(id)
);

CREATE TABLE teams (
    id TEXT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    color TEXT NOT NULL,
    org_id TEXT NOT NULL,
    FOREIGN KEY (org_id) REFERENCES orgs(id)
);

CREATE TABLE users (
    id TEXT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    display_name TEXT NOT NULL,
    email TEXT NOT NULL,
    is_me BOOLEAN NOT NULL,
    org_id TEXT NOT NULL,
    FOREIGN KEY (org_id) REFERENCES orgs(id)
);

CREATE UNIQUE INDEX unique_idx_users_is_me_per_org
ON users (org_id)
WHERE is_me = TRUE;

CREATE TABLE states (
    id TEXT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    color TEXT NOT NULL,
    team_id TEXT NOT NULL,
    org_id TEXT NOT NULL
);

CREATE TABLE issues (
    id TEXT PRIMARY KEY NOT NULL,
    identifier TEXT NOT NULL,
    title TEXT NOT NULL,
    priority INTEGER NOT NULL,
    org_id TEXT NOT NULL,
    state_id TEXT NOT NULL,
    team_id TEXT NOT NULL,
    description TEXT,
    labels BLOB,
    assignee_id TEXT,
    project_id TEXT NOT NULL,
    pinned BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP,
    canceled_at TIMESTAMP,
    FOREIGN KEY (org_id) REFERENCES orgs(id),
    FOREIGN KEY (project_id) REFERENCES projects(id),
    FOREIGN KEY (team_id) REFERENCES teams(id),
    FOREIGN KEY (assignee_id) REFERENCES users(id),
    FOREIGN KEY (project_id) REFERENCES projects(id)
);

CREATE VIRTUAL TABLE search USING fts5 (
  tokenize = "trigram",
  id UNINDEXED,
  title,
  description,
  state,
  project,
  team,
  assignee,
  labels
);
