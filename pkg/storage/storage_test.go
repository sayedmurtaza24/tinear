package storage

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/sayedmurtaza24/tinear/pkg/linear/issue"
	"github.com/sayedmurtaza24/tinear/pkg/linear/label"
	"github.com/sayedmurtaza24/tinear/pkg/linear/prio"
	"github.com/sayedmurtaza24/tinear/pkg/linear/project"
	"github.com/sayedmurtaza24/tinear/pkg/linear/state"
	"github.com/sayedmurtaza24/tinear/pkg/linear/team"
	"github.com/sayedmurtaza24/tinear/pkg/linear/user"
)

func prepareTestIssues() []issue.Issue {
	var testIssues []issue.Issue

	for i := 0; i < 10; i++ {
		testIssues = append(testIssues, issue.Issue{
			ID:         "test-id-" + fmt.Sprint(i),
			Identifier: "test-identifier",
			Title:      "test-title",
			Desc:       "test-desc",
			Labels:     []label.Label{{Name: "test", Color: "test-color"}},
			Assignee:   user.User{ID: "user-id", DisplayName: "user-test", Email: "usertestemail@test.com", IsMe: i%2 == 0},
			Priority:   prio.Prio(i),
			Team:       team.Team{ID: "team-id", Name: "team-test", Color: "team-test-color"},
			State:      state.State{Name: "state-test", Color: "state-test-color"},
			Project:    project.Project{Name: "project-test", Color: "project-test-color"},
		})
	}

	return testIssues
}

func TestDiffPut(t *testing.T) {
	testIssues := prepareTestIssues()

	updatedIssue := issue.Issue{
		ID:         "test-id-2",
		Identifier: "test-identifier updated",
		Title:      "test-title updated",
		Desc:       "test-desc",
		Labels:     []label.Label{{Name: "test", Color: "test-color"}, {Name: "test2", Color: "test-color2"}},
		Assignee:   user.User{ID: "user-id", DisplayName: "user-test", Email: "usertestemail@test.com", IsMe: true},
		Priority:   prio.Prio(5),
		Team:       team.Team{ID: "team-id", Name: "team-test", Color: "team-test-color"},
		State:      state.State{Name: "state-test", Color: "state-test-color"},
		Project:    project.Project{Name: "project-test", Color: "project-test-color"},
	}

	var store IssueStore

	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	store.read = true
	store.issues = testIssues
	store.path = path.Join(dir, "tinear-issues-diff.json")

	err = store.PutDiff(updatedIssue)
	if err != nil {
		t.Fatal(err)
	}

	if !store.issues[2].Equal(updatedIssue) {
		t.Fatalf("expected %v, got %v", updatedIssue, store.issues[2])
	}

	if len(store.issues) != 10 {
		t.Fatalf("expected 10 issues, got %d", len(store.issues))
	}
}

func TestReset(t *testing.T) {
	testIssues := prepareTestIssues()

	var store IssueStore

	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	store.path = path.Join(dir, "tinear-issues-w.json")

	err = store.PutAll(testIssues...)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGet(t *testing.T) {
	testIssues := prepareTestIssues()

	var store IssueStore

	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	store.path = path.Join(dir, "tinear-issues-r.json")
	issues := store.Get()

	if issues == nil {
		t.Fatal("issues is nil")
	}

	if len(issues) != 10 {
		t.Fatalf("expected 10 issues, got %d", len(issues))
	}

	for i, issue := range issues {
		if !issue.Equal(testIssues[i]) {
			t.Fatalf("expected %v, got %v", testIssues[i], issue)
		}
	}
}
