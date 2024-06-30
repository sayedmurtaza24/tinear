package store

import (
	"encoding/json"
	"fmt"
	"time"
)

type Organization struct {
	ID   string
	Name string
}

type Project struct {
	ID    string
	Name  string
	Color string
}

type State struct {
	Name     string
	Color    string
	Position int
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
	DisplayName string
	Email       string
	IsMe        bool
}

type Prio int

type Label []byte

type ParsedLabel struct {
	Name  string
	Color string
}

func (l Label) Parse() ([]ParsedLabel, error) {
	var parsed []ParsedLabel
	err := json.Unmarshal(l, &parsed)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse labels: %w", err)
	}
	return parsed, nil
}

func ToLabel(p []ParsedLabel) ([]byte, error) {
	b, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("couldn't marshal parsed labels: %w", err)
	}
	return b, nil
}

type Issue struct {
	ID         string
	Identifier string
	Title      string
	Desc       string
	Labels     Label
	Assignee   User
	Priority   Prio
	Team       Team
	State      State
	Project    Project
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
