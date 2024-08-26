package client

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sayedmurtaza24/tinear/linear/models"
)

type issueLabelMutation string

const (
	issueLabelAdd    issueLabelMutation = "issueAddLabel"
	issueLabelRemove issueLabelMutation = "issueRemoveLabel"
)

func buildUpdateLabelQuery(mut issueLabelMutation, labelID string, issueIDs ...string) (string, map[string]interface{}) {
	if len(issueIDs) == 0 {
		return "", nil
	}

	var issueIDVars []string
	var issueMutations []string
	issueMutationArgs := map[string]interface{}{
		"labelId": labelID,
	}

	for i, issueID := range issueIDs {
		issueIDvar := fmt.Sprintf("$issueId%d", i+1)
		issueIDVars = append(issueIDVars, fmt.Sprintf("%s: String!", issueIDvar))
		issueMutations = append(issueMutations, fmt.Sprintf(`
			update%d: %s(labelId: $labelId, id: %s) {
				success
			}
		`, i+1, mut, issueIDvar))
		issueMutationArgs[strings.TrimPrefix(issueIDvar, "$")] = issueID
	}

	query := fmt.Sprintf(`
		mutation IssueBatchUpdate($labelId: String!, %s) {
			%s
		}
	`, strings.Join(issueIDVars, ", "), strings.Join(issueMutations, ",\n"))

	return query, issueMutationArgs
}

type issueUpdateOpt struct {
	hasOpt    bool
	opt       models.IssueUpdateInput
	label     string
	labelsMut issueLabelMutation
}

type IssueUpdateOpt func(*issueUpdateOpt)

func WithAddLabels(labelID string) IssueUpdateOpt {
	return func(i *issueUpdateOpt) {
		i.label = labelID
		i.labelsMut = issueLabelAdd
	}
}

func WithRemoveLabels(labelID string) IssueUpdateOpt {
	return func(i *issueUpdateOpt) {
		i.label = labelID
		i.labelsMut = issueLabelRemove
	}
}

func WithSetAssignee(assigneeID string) IssueUpdateOpt {
	return func(i *issueUpdateOpt) {
		i.hasOpt = true
		i.opt.AssigneeID = &assigneeID
	}
}

func WithSetState(stateID string) IssueUpdateOpt {
	return func(i *issueUpdateOpt) {
		i.hasOpt = true
		i.opt.StateID = &stateID
	}
}

func WithSetPrio(priority int64) IssueUpdateOpt {
	return func(i *issueUpdateOpt) {
		i.hasOpt = true
		i.opt.Priority = &priority
	}
}

func WithSetProject(projectID string) IssueUpdateOpt {
	return func(i *issueUpdateOpt) {
		i.hasOpt = true
		i.opt.ProjectID = &projectID
	}
}

func WithSetTeam(teamID string) IssueUpdateOpt {
	return func(i *issueUpdateOpt) {
		i.hasOpt = true
		i.opt.TeamID = &teamID
	}
}

func WithSetTitle(title string) IssueUpdateOpt {
	return func(i *issueUpdateOpt) {
		i.hasOpt = true
		i.opt.Title = &title
	}
}

type UpdateIssuesResponse struct {
	Success       bool
	OnFailCommand tea.Cmd
}

func (c *Client) UpdateIssues(issueIDs []string, onFail tea.Cmd, opts ...IssueUpdateOpt) tea.Cmd {
	return func() tea.Msg {
		var response UpdateIssuesResponse
		var input issueUpdateOpt

		for _, opt := range opts {
			opt(&input)
		}

		if input.hasOpt {
			resp, err := c.client.BatchUpdateIssues(context.Background(), input.opt, issueIDs)
			if err != nil {
				return err
			}
			response.Success = resp.GetIssueBatchUpdate().GetSuccess()
		}

		if input.label != "" {
			resp := make(map[string]struct {
				Data map[string]struct{ Success bool }
			})

			query, args := buildUpdateLabelQuery(input.labelsMut, input.label, issueIDs...)
			err := c.rawClient.Post(context.Background(), "", query, &resp, args)
			if err != nil {
				return err
			}

			// bad way, but works for now
			for _, v := range resp["data"].Data {
				if !v.Success {
					response.Success = false
					break
				}
			}

			response.Success = true
		}

		return response
	}
}
