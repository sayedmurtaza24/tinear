package command

import (
	"context"
	"log/slog"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	linearClient "github.com/sayedmurtaza24/tinear/linear"
	"github.com/sayedmurtaza24/tinear/linear/models"
	"github.com/sayedmurtaza24/tinear/pkg/linear/issue"
	"github.com/sayedmurtaza24/tinear/pkg/storage"
)

const sixMonths = 6 * 30 * 24 * time.Hour

type GetIssuesRes Resumable[[]issue.Issue]

func GetIssues(client linearClient.LinearClient, store storage.IssueStore, after *string) tea.Cmd {
	return func() tea.Msg {
		lastReset := store.LastReset().Format(time.RFC3339)

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

		resp, err := client.GetIssues(
			context.Background(),
			&filter,
			after,
			first(),
		)
		if err != nil {
			return err
		}

		issues := issue.FromLinearClientGetIssues(resp)

		slog.Info("[command.GetIssues]", "len", len(issues))

		err = store.Put(issues...)
		if err != nil {
			panic(err)
		}

		return GetIssuesRes(paginated(issues, &resp.Issues.PageInfo))
	}
}
