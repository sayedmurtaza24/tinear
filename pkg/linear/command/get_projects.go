package command

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	linearClient "github.com/sayedmurtaza24/tinear/linear"
	"github.com/sayedmurtaza24/tinear/pkg/linear/project"
)

type GetProjectsRes Command[[]project.Project]

func GetProjects(client linearClient.LinearClient, after *string) tea.Cmd {
	return func() tea.Msg {
		resp, err := client.GetProjects(
			context.Background(),
			after,
			first(),
		)
		if err != nil {
			return err
		}

		projects := project.FromLinearClientGetProjects(*resp)

		return paginated(projects, &resp.Projects.PageInfo)
	}
}
