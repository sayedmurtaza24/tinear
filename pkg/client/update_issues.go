package client

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sayedmurtaza24/tinear/linear/models"
)

type IssueUpdateOpt func(*models.IssueUpdateInput)

func WithSetAssignee(assigneeID string) IssueUpdateOpt {
	return func(i *models.IssueUpdateInput) {
		i.AssigneeID = &assigneeID
	}
}

func WithSetState(stateID string) IssueUpdateOpt {
	return func(i *models.IssueUpdateInput) {
		i.StateID = &stateID
	}
}

func WithSetPrio(priority int64) IssueUpdateOpt {
	return func(i *models.IssueUpdateInput) {
		i.Priority = &priority
	}
}

func WithSetProject(projectID string) IssueUpdateOpt {
	return func(i *models.IssueUpdateInput) {
		i.ProjectID = &projectID
	}
}

func WithSetTeam(teamID string) IssueUpdateOpt {
	return func(i *models.IssueUpdateInput) {
		i.TeamID = &teamID
	}
}

func WithSetTitle(title string) IssueUpdateOpt {
	return func(i *models.IssueUpdateInput) {
		i.Title = &title
	}
}

type UpdateIssuesResponse struct {
	Success       bool
	OnFailCommand tea.Cmd
}

func (c *Client) UpdateIssues(issueIDs []string, onFail tea.Cmd, opts ...IssueUpdateOpt) tea.Cmd {
	return func() tea.Msg {
		var response UpdateIssuesResponse
		var input models.IssueUpdateInput

		for _, opt := range opts {
			opt(&input)
		}

		resp, err := c.client.BatchUpdateIssues(context.Background(), any(input).(models.IssueUpdateInput), issueIDs)
		if err != nil {
			return err
		}

		response.Success = resp.GetIssueBatchUpdate().GetSuccess()

		return response
	}
}
