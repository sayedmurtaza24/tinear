package client

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sayedmurtaza24/tinear/pkg/store"
)

type Me struct {
	Me     store.User
	Org    store.Org
	Teams  []store.Team
	States []store.State
	Labels []store.Label
}

type GetMeRes Command[Me]

func (c *Client) GetMe() tea.Cmd {
	return func() tea.Msg {
		resp, err := c.client.GetMe(context.Background())
		if err != nil {
			return err
		}

		me := Me{
			Me: store.User{
				ID:          resp.Viewer.ID,
				Name:        resp.Viewer.Name,
				DisplayName: resp.Viewer.DisplayName,
				Email:       resp.Viewer.Email,
				IsMe:        true,
			},
			Org: store.Org{
				ID:     resp.Viewer.Organization.ID,
				Name:   resp.Viewer.Organization.Name,
				URLKey: resp.Viewer.Organization.URLKey,
			},
		}

		coalesce := func(s *string) string {
			if s == nil {
				return "#bbb"
			}
			return *s
		}

		for _, team := range resp.Viewer.Teams.GetNodes() {
			me.Teams = append(me.Teams, store.Team{
				ID:    team.ID,
				Name:  team.Name,
				Color: coalesce(team.Color),
			})

			for _, state := range team.States.GetNodes() {
				me.States = append(me.States, store.State{
					ID:     state.ID,
					Name:   state.Name,
					Color:  state.Color,
					TeamID: team.ID,
				})
			}

			for _, label := range team.Labels.GetNodes() {
				if label.IsGroup {
					continue
				}
				me.Labels = append(me.Labels, store.Label{
					ID:     label.ID,
					Name:   label.Name,
					Color:  label.Color,
					TeamID: label.GetTeam().GetID(),
				})
			}
		}

		return GetMeRes(response(me))
	}
}
