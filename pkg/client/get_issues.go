package client

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sayedmurtaza24/tinear/linear/models"
	"github.com/sayedmurtaza24/tinear/pkg/store"
)

const syncThreshold = -6 * 30 * 24 * time.Hour

type GetIssuesRes Resumable[[]store.Issue]

func (c *Client) GetIssues(lastSync time.Time, after *string) tea.Cmd {
	return func() tea.Msg {
		syncedAt := lastSync.Format(time.RFC3339)

		if lastSync.IsZero() {
			syncedAt = time.Now().Add(syncThreshold).Format(time.RFC3339)
		}

		filter := models.IssueFilter{
			Or: []*models.IssueFilter{
				{
					UpdatedAt: &models.DateComparator{Gte: &syncedAt},
				},
				{
					CreatedAt: &models.DateComparator{Gte: &syncedAt},
				},
				{
					CanceledAt: &models.NullableDateComparator{Gte: &syncedAt},
				},
			},
		}

		resp, err := c.client.GetIssues(
			context.Background(),
			&filter,
			after,
			first(),
		)
		if err != nil {
			return err
		}

		coalece := func(n *string, c string) string {
			if n == nil {
				return c
			}
			return *n
		}

		if resp == nil {
			return []store.Issue{}
		}

		var issues []store.Issue

		for _, iss := range resp.Issues.GetNodes() {
			createdAt, err := time.Parse(time.RFC3339, iss.CreatedAt)
			if err != nil {
				return fmt.Errorf("error parsing created_at")
			}

			updatedAt, err := time.Parse(time.RFC3339, iss.UpdatedAt)
			if err != nil {
				return fmt.Errorf("error parsing updated_at")
			}

			var canceledAt *time.Time
			if iss.CanceledAt != nil {
				t, err := time.Parse(time.RFC3339, *iss.CanceledAt)
				if err != nil {
					return fmt.Errorf("error parsing canceled_at")
				}
				canceledAt = &t
			}

			labels := make([]store.Label, len(iss.Labels.GetNodes()))
			for i, label := range iss.Labels.GetNodes() {
				labels[i] = store.Label{
					ID:     label.ID,
					Name:   label.Name,
					Color:  label.Color,
					TeamID: label.GetTeam().GetID(),
				}
			}

			is := store.Issue{
				ID:          iss.GetID(),
				Identifier:  iss.GetIdentifier(),
				Title:       iss.GetTitle(),
				Description: coalece(iss.Description, ""),
				Assignee: store.User{
					ID:          iss.GetAssignee().GetID(),
					Name:        iss.GetAssignee().GetName(),
					DisplayName: iss.GetAssignee().GetDisplayName(),
					Email:       iss.GetAssignee().GetEmail(),
					IsMe:        iss.GetAssignee().GetIsMe(),
				},
				Labels:   labels,
				Priority: store.Prio(iss.GetPriority()),
				Team: store.Team{
					ID:    iss.GetTeam().GetID(),
					Name:  iss.GetTeam().GetName(),
					Color: coalece(iss.GetTeam().GetColor(), "#bbb"),
				},
				State: store.State{
					ID:     iss.GetState().GetID(),
					Name:   iss.GetState().GetName(),
					Color:  iss.GetState().GetColor(),
					TeamID: iss.GetState().GetTeam().GetID(),
				},
				Project: store.Project{
					ID:    iss.GetProject().GetID(),
					Name:  iss.GetProject().GetName(),
					Color: iss.GetProject().GetColor(),
				},
				CreatedAt:  createdAt,
				UpdatedAt:  updatedAt,
				CanceledAt: canceledAt,
			}

			issues = append(issues, is)
		}

		return GetIssuesRes(paginated(issues, &resp.Issues.PageInfo))
	}
}
