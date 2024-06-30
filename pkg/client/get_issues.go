package client

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	linearClient "github.com/sayedmurtaza24/tinear/linear"
	"github.com/sayedmurtaza24/tinear/linear/models"
	"github.com/sayedmurtaza24/tinear/pkg/store"
)

const sixMonths = 6 * 30 * 24 * time.Hour

type GetIssuesRes Resumable[[]store.Issue]

func (c *Client) GetIssues(after *string) tea.Cmd {
	return func() tea.Msg {
		lastReset := time.Now().Format(time.RFC3339)

		filter := models.IssueFilter{
			Or: []*models.IssueFilter{
				{
					UpdatedAt: &models.DateComparator{Gte: &lastReset},
				},
				{
					CreatedAt: &models.DateComparator{Gte: &lastReset},
				},
				{
					CanceledAt: &models.NullableDateComparator{Gte: &lastReset},
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

		toString := func(n *string) string {
			if n == nil {
				return ""
			}

			return *n
		}

		if resp == nil {
			return []store.Issue{}
		}

		var issues []store.Issue

		for _, iss := range resp.Issues.GetNodes() {
			// var labels []parsedLabel
			// for _, l := range iss.GetLabels().GetNodes() {
			// 	labels = append(labels, parsedLabel{
			// 		Name:  l.Name,
			// 		Color: l.Color,
			// 	})
			// }
			// encodedLabelsB, err := json.Marshal(labels)
			// if err != nil {
			// 	panic(err)
			// }

			createdAt, err := time.Parse(time.RFC3339, iss.CreatedAt)
			if err != nil {
				panic(err)
			}

			updatedAt, err := time.Parse(time.RFC3339, iss.UpdatedAt)
			if err != nil {
				panic(err)
			}

			assignee := iss.Assignee

			if assignee == nil {
				assignee = &linearClient.GetIssues_Issues_Nodes_Assignee{}
			}

			is := store.Issue{
				ID:         iss.ID,
				Identifier: iss.Identifier,
				Title:      iss.Title,
				Desc:       toString(iss.Description),
				Assignee: store.User{
					ID:          assignee.ID,
					DisplayName: assignee.DisplayName,
					Email:       assignee.Email,
					IsMe:        assignee.IsMe,
				},
				// Labels:   encodedLabelsB,
				Priority: store.Prio(iss.Priority),
				Team: store.Team{
					ID:    iss.Team.ID,
					Name:  iss.Team.Name,
					Color: toString(iss.Team.Color),
				},
				State: store.State{
					Name:     iss.State.Name,
					Color:    iss.State.Color,
					Position: int(iss.State.Position),
				},
				Project: store.Project{
					Name:  iss.GetProject().GetName(),
					Color: iss.GetProject().GetColor(),
				},
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			}

			issues = append(issues, is)
		}

		return GetIssuesRes(paginated(issues, &resp.Issues.PageInfo))
	}
}
