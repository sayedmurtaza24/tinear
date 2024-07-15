package client

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sayedmurtaza24/tinear/pkg/store"
)

type GetUsersRes Resumable[[]store.User]

func (c *Client) GetUsers(after *string) tea.Cmd {
	return func() tea.Msg {
		resp, err := c.client.GetAllUsers(context.Background(), after, first())
		if err != nil {
			return err
		}

		var users []store.User

		for _, user := range resp.Users.Nodes {
			users = append(users, store.User{
				ID:          user.ID,
				Name:        user.Name,
				DisplayName: user.DisplayName,
				Email:       user.Email,
				IsMe:        user.IsMe,
			})
		}

		return GetUsersRes(paginated(users, &resp.Users.PageInfo))
	}
}
