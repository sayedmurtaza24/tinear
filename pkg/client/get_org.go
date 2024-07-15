package client

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sayedmurtaza24/tinear/pkg/store"
)

type GetOrgRes Command[store.Org]

func (c *Client) GetOrg() tea.Cmd {
	return func() tea.Msg {
		resp, err := c.client.GetOrg(context.Background())
		if err != nil {
			return err
		}
		org := store.Org{
			ID:     resp.Organization.ID,
			Name:   resp.Organization.Name,
			URLKey: resp.Organization.URLKey,
		}
		return GetOrgRes(response(org))
	}
}
