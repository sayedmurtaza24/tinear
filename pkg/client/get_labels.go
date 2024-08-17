package client

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sayedmurtaza24/tinear/pkg/store"
)

type GetLabelsRes Resumable[[]store.Label]

func (c *Client) GetLabels(after *string) tea.Cmd {
	return func() tea.Msg {
		resp, err := c.client.GetLabels(
			context.Background(),
			after,
			first(),
		)
		if err != nil {
			return err
		}

		var labels []store.Label

		for _, label := range resp.IssueLabels.GetNodes() {
			labels = append(labels, store.Label{
				ID:     label.ID,
				Name:   label.Name,
				Color:  label.Color,
				TeamID: label.GetTeam().GetID(),
			})
		}

		return GetLabelsRes(paginated(labels, &resp.IssueLabels.PageInfo))
	}
}
