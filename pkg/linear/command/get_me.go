package command

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	linearClient "github.com/sayedmurtaza24/tinear/linear"
	"github.com/sayedmurtaza24/tinear/pkg/linear/user"
)

type GetMeRes Command[user.User]

func GetMe(client linearClient.LinearClient) tea.Cmd {
	return func() tea.Msg {
		resp, err := client.GetCurrentUser(context.Background())
		if err != nil {
			return err
		}

		return response(user.User{
			ID:          resp.Viewer.ID,
			DisplayName: resp.Viewer.DisplayName,
			Email:       resp.Viewer.Email,
			IsMe:        resp.Viewer.IsMe,
			OrgName:     resp.Viewer.Organization.Name,
		})
	}
}
