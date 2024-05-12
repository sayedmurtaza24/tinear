package dashboard

import (
	"github.com/sayedmurtaza24/tinear/pkg/linear/issue"
	"github.com/sayedmurtaza24/tinear/pkg/linear/project"
	"github.com/sayedmurtaza24/tinear/pkg/linear/user"
)

type ViewMode int

const (
	ViewModeMyIssues ViewMode = iota
	ViewModeProjects
)

type DashboardState struct {
	Me               user.User
	OrganizationName string
	MyIssues         []issue.Issue
	Projects         map[project.Project]issue.Issue
}
