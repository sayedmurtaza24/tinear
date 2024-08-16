package client

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sayedmurtaza24/tinear/pkg/store"
)

type GetStatesRes Resumable[[]store.State]

func (c *Client) GetStates(after *string) tea.Cmd {
	return func() tea.Msg {
		resp, err := c.client.GetStates(context.Background(), after, first())
		if err != nil {
			return err
		}

		var states []store.State

		for _, state := range resp.WorkflowStates.GetNodes() {
			states = append(states, store.State{
				ID:     state.ID,
				Name:   state.Name,
				Color:  state.Color,
				TeamID: state.Team.ID,
			})
		}

		return GetStatesRes(paginated(states, &resp.WorkflowStates.PageInfo))
	}
}
