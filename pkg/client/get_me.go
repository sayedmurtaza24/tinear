package client

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sayedmurtaza24/tinear/pkg/store"
)

type GetMeRes Command[store.User]

func (c *Client) GetMe() tea.Cmd {
	return func() tea.Msg {
		resp, err := c.client.GetCurrentUser(context.Background())
		if err != nil {
			return err
		}

		return GetMeRes(response(store.User{
			ID:          resp.Viewer.ID,
			DisplayName: resp.Viewer.DisplayName,
			Email:       resp.Viewer.Email,
			IsMe:        resp.Viewer.IsMe,
		}))
	}
}
