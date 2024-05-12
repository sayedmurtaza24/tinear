package dashboard

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	linearClient "github.com/sayedmurtaza24/tinear/linear"
	"github.com/sayedmurtaza24/tinear/linear/models"
	"github.com/sayedmurtaza24/tinear/pkg/linear/issue"
	"github.com/sayedmurtaza24/tinear/pkg/linear/project"
	"github.com/sayedmurtaza24/tinear/pkg/linear/resumable"
	"github.com/sayedmurtaza24/tinear/pkg/linear/sort"
	"github.com/sayedmurtaza24/tinear/pkg/linear/user"
)

var first int64 = 50

type GetMeResponse struct {
	user.User
	OrganizationName string
}

// returns GetMeResponse or ErroredCommand
func GetMe(client linearClient.LinearClient) tea.Cmd {
	return func() tea.Msg {
		response, err := client.GetCurrentUser(context.Background())
		if err != nil {
			return err
		}

		return GetMeResponse{
			User: user.User{
				ID:          response.Viewer.ID,
				DisplayName: response.Viewer.DisplayName,
				Email:       response.Viewer.Email,
				IsMe:        response.Viewer.IsMe,
			},
			OrganizationName: response.Viewer.Organization.Name,
		}
	}
}

type GetMyIssuesResponse []issue.Issue

func GetMyIssues(client linearClient.LinearClient, sort sort.SortOption, after *string) tea.Cmd {
	return func() tea.Msg {
		isMe := true

		filter := models.IssueFilter{
			Assignee: &models.NullableUserFilter{
				IsMe: &models.BooleanComparator{
					Eq: &isMe,
				},
			},
		}

		response, err := client.GetIssues(
			context.Background(),
			sort.ToIssueSortInput(),
			&filter,
			after,
			&first,
		)
		if err != nil {
			return err
		}

		issues := issue.FromLinearClientGetIssues(*response)

		return resumable.FromLinearClientResponse(issues, &response.Issues.PageInfo)
	}
}

type GetProjectsResponse []project.Project

func GetProjects(client linearClient.LinearClient, after *string) tea.Cmd {
	return func() tea.Msg {
		response, err := client.GetProjects(
			context.Background(),
			after,
			&first,
		)
		if err != nil {
			return err
		}

		projects := project.FromLinearClientGetProjects(*response)

		return resumable.FromLinearClientResponse(projects, &response.Projects.PageInfo)
	}
}
