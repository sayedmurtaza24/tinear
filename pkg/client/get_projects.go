package client

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sayedmurtaza24/tinear/pkg/store"
)

type GetProjectsRes Resumable[[]store.Project]

func (c *Client) GetProjects(after *string) tea.Cmd {
	return func() tea.Msg {
		resp, err := c.client.GetProjects(
			context.Background(),
			after,
			first(),
		)
		if err != nil {
			return err
		}

		var projects []store.Project

		for _, proj := range resp.Projects.GetNodes() {
			projects = append(projects, store.Project{
				ID:    proj.ID,
				Name:  proj.Name,
				Color: proj.Color,
			})
		}

		return GetProjectsRes(paginated(projects, &resp.Projects.PageInfo))
	}
}
