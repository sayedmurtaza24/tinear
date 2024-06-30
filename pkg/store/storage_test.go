package store

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"
)

func prepareTestIssues() []Issue {
	var testIssues []Issue

	for i := 0; i < 10; i++ {
		testIssues = append(testIssues, Issue{
			ID:         "test-id-" + fmt.Sprint(i),
			Identifier: "test-identifier",
			Title:      "test-title",
			Desc:       "test-desc",
			// Labels:     []Label{{Name: "test", Color: "test-color"}},
			Assignee: User{ID: "user-id", DisplayName: "user-test", Email: "usertestemail@test.com", IsMe: i%2 == 0},
			Priority: Prio(i),
			Team:     Team{ID: "team-id", Name: "team-test", Color: "team-test-color"},
			State:    State{Name: "state-test", Color: "state-test-color"},
			Project:  Project{Name: "project-test", Color: "project-test-color"},
		})
	}

	return testIssues
}

func TestStore(t *testing.T) {
	store, err := New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	if store.config.orgID != "" {
		t.Fatal("expected empty org_id")
	}

	times := []*time.Time{
		&store.config.issues,
		&store.config.teams,
		&store.config.projects,
		&store.config.users,
	}

	for _, tim := range times {
		if !tim.Before(time.Now().UTC().Add(sixMonthsAgo + 2*time.Minute)) {
			t.Fatalf("expected 6 months ago, got %s", tim)
		}
	}

	var orgs []Organization

	for i := 0; i < 10; i++ {
		orgs = append(orgs, Organization{
			ID:   fmt.Sprintf("id-%d", i),
			Name: fmt.Sprintf("org-%d", i),
		})
	}

	err = store.StoreOrgs(context.Background(), orgs)
	if err != nil {
		t.Fatal(err)
	}

	result, err := batchSelect(
		store.db,
		"orgs",
		[]string{"id", "name"},
		func(org *Organization) []any {
			return []any{&org.ID, &org.Name}
		},
		"WHERE name LIKE '%org-%'",
	)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(orgs, result) {
		t.Fatalf("expected deep equal: wanted %v, got %v", orgs, result)
	}

	err = batchDelete(store.db, "orgs", "WHERE id = 'id-0'")
	if err != nil {
		t.Fatal(err)
	}

	result, err = batchSelect(
		store.db,
		"orgs",
		[]string{"id", "name"},
		func(org *Organization) []any {
			return []any{&org.ID, &org.Name}
		},
		"WHERE name LIKE '%org-%'",
	)
	if err != nil {
		t.Fatal(err)
	}

	for _, res := range result {
		if res.ID == "id-0" {
			t.Fatal("org id-0 expected to be deleted")
		}
	}
}
