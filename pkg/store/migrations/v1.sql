BEGIN;

CREATE TABLE migrations (
    version INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE orgs (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE TABLE projects (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    color TEXT NOT NULL,
    org_id TEXT NOT NULL,
    FOREIGN KEY (org_id) REFERENCES org(id)
);

CREATE TABLE teams (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    color TEXT NOT NULL,
    org_id TEXT NOT NULL,
    FOREIGN KEY (org_id) REFERENCES org(id)
);

CREATE TABLE users (
    id TEXT PRIMARY KEY,
    display_name TEXT NOT NULL,
    email TEXT NOT NULL,
    is_me BOOLEAN NOT NULL,
    org_id TEXT NOT NULL,
    FOREIGN KEY (org_id) REFERENCES org(id)
);

CREATE TABLE issues (
    id TEXT PRIMARY KEY,
    identifier TEXT NOT NULL,
    title TEXT NOT NULL,
    state INTEGER NOT NULL,
    description TEXT,
    labels BLOB,
    priority INTEGER,
    org_id TEXT NOT NULL,
    user_id TEXT,
    team_id TEXT,
    project_id TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (org_id) REFERENCES org(id),
    FOREIGN KEY (project_id) REFERENCES projects(id),
    FOREIGN KEY (team_id) REFERENCES teams(id),
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (project) REFERENCES projects(id)
);

CREATE TABLE config (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  key TEXT NOT NULL,
  value TEXT NOT NULL
);

CREATE TRIGGER update_timestamp
AFTER UPDATE ON issues
FOR EACH ROW
BEGIN
    UPDATE tasks SET updated_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;

COMMIT;
