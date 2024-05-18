package command

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	linearClient "github.com/sayedmurtaza24/tinear/linear"
	"github.com/sayedmurtaza24/tinear/linear/models"
	"github.com/sayedmurtaza24/tinear/pkg/linear/issue"
)

type GetIssuesRes Resumable[[]issue.Issue]

func GetIssues(client linearClient.LinearClient, all bool, after *string) tea.Cmd {
	return func() tea.Msg {
		now := time.Now().Format(time.RFC3339)

		if all {
			now = time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
		}

		filter := models.IssueFilter{
			Or: []*models.IssueFilter{
				{
					UpdatedAt: &models.DateComparator{
						Gt: &now,
					},
				},
				{
					CreatedAt: &models.DateComparator{
						Gt: &now,
					},
				},
			},
		}

		resp, err := client.GetIssues(
			context.Background(),
			&filter,
			after,
			first(),
		)
		if err != nil {
			return err
		}

		issues := issue.FromLinearClientGetIssues(*resp)

		return paginated(issues, &resp.Issues.PageInfo)
	}
}
