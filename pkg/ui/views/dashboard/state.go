package dashboard

import (
	"github.com/sayedmurtaza24/tinear/pkg/ui/linear/issue"
	"github.com/sayedmurtaza24/tinear/pkg/ui/linear/user"
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
	Projects         []issue.Issue
}
