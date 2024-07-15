package client

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sayedmurtaza24/tinear/pkg/store"
)

type GetTeamsRes Resumable[[]store.Team]

func (c *Client) GetTeams(after *string) tea.Cmd {
	return func() tea.Msg {
		resp, err := c.client.GetTeams(
			context.Background(),
			after,
			first(),
		)
		if err != nil {
			return err
		}

		color := func(n *string) string {
			if n == nil {
				return "#bbb"
			}
			return *n
		}

		var teams []store.Team

		for _, proj := range resp.Teams.GetNodes() {
			teams = append(teams, store.Team{
				ID:    proj.ID,
				Name:  proj.Name,
				Color: color(proj.Color),
			})
		}

		return GetTeamsRes(paginated(teams, &resp.Teams.PageInfo))
	}
}
