package store

import (
	"time"
)

type storeResource interface {
	getID() string
}

type Org struct {
	ID        string
	Name      string
	URLKey    string
	Active    bool
	SyncedAt  time.Time
	SortMode  SortMode
	SortOrder sortOrder
}

type Project struct {
	ID    string
	Name  string
	Color string
	Teams []Team
}

type State struct {
	ID     string
	Name   string
	Color  string
	TeamID string
}

func (state *State) position(name string, pos int) int {
	switch name {
	case "Canceled":
		return 0
	case "Done":
		return 1
	case "Triage":
		return 2
	case "Backlog":
		return 3
	case "QA Ready":
		return 4
	case "In Review":
		return 5
	case "Todo":
		return 6
	case "In Progress":
		return 7
	default:
		return pos
	}
}

type Team struct {
	ID    string
	Name  string
	Color string
}

type User struct {
	ID          string
	Name        string
	DisplayName string
	Email       string
	IsMe        bool
}

type Prio int

type Label struct {
	ID     string
	Name   string
	Color  string
	TeamID string
}

type Issue struct {
	ID          string
	Identifier  string
	Title       string
	Description string
	Labels      []Label
	Priority    Prio
	Team        Team
	State       State
	Assignee    User
	Project     Project
	Pinned      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CanceledAt  *time.Time
}

func (u Org) getID() string     { return u.ID }
func (u User) getID() string    { return u.ID }
func (u Team) getID() string    { return u.ID }
func (u Project) getID() string { return u.ID }
func (u State) getID() string   { return u.ID }
func (u Issue) getID() string   { return u.ID }
func (u Label) getID() string   { return u.ID }
