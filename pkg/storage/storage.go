package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/sayedmurtaza24/tinear/pkg/linear/issue"
)

type storage struct {
	Issues   []issue.Issue `json:"issues"`
	LastSync time.Time     `json:"last_sync"`
}

type IssueStore struct {
	path     string
	read     bool
	issues   issue.IssueList
	lastSync time.Time
}

func New() IssueStore {
	return IssueStore{}
}

func (s *IssueStore) Get() []issue.Issue {
	if s.read {
		return s.issues
	}

	f, err := os.Open(s.getPath())
	if err != nil {
		return make([]issue.Issue, 0)
	}
	defer f.Close()

	var strg storage

	err = json.NewDecoder(f).Decode(&strg)
	if err != nil {
		return nil
	}

	s.issues = strg.Issues
	s.lastSync = strg.LastSync
	s.read = true

	return strg.Issues
}

func (s *IssueStore) Put(upsertedIssues ...issue.Issue) error {
	if !s.read {
		s.Get()
	}

	for _, upsertedIssue := range upsertedIssues {
		issueFindFunc := func(issue issue.Issue) bool {
			return issue.ID == upsertedIssue.ID
		}

		foundIssue := slices.IndexFunc(s.issues, issueFindFunc)

		if foundIssue == -1 {
			s.issues = append(s.issues, upsertedIssue)
		} else {
			s.issues[foundIssue] = upsertedIssue
		}
	}

	return s.put(s.issues)
}

func (s *IssueStore) LastReset() time.Time {
	return s.lastSync.UTC()
}

func (s *IssueStore) put(issues []issue.Issue) error {
	s.read = true
	s.issues = issues

	f, err := os.OpenFile(s.getPath(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer f.Close()

	strg := storage{
		Issues:   issues,
		LastSync: time.Now().UTC(),
	}

	return json.NewEncoder(f).Encode(strg)
}

func (s *IssueStore) getPath() string {
	if s.path != "" {
		return s.path
	}

	return "/tmp/tinear-issues.json"
}
