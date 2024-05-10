package issue

import (
	"strings"

	linearClient "github.com/sayedmurtaza24/tinear/linear"
	"github.com/sayedmurtaza24/tinear/pkg/ui/color"
	"github.com/sayedmurtaza24/tinear/pkg/ui/linear/label"
	"github.com/sayedmurtaza24/tinear/pkg/ui/linear/prio"
	"github.com/sayedmurtaza24/tinear/pkg/ui/linear/project"
	"github.com/sayedmurtaza24/tinear/pkg/ui/linear/state"
	"github.com/sayedmurtaza24/tinear/pkg/ui/linear/team"
	"github.com/sayedmurtaza24/tinear/pkg/ui/linear/user"
	"github.com/sayedmurtaza24/tinear/pkg/ui/molecules/table"
	"github.com/sayedmurtaza24/tinear/pkg/ui/text"
)

type Issue struct {
	ID         string
	Identifier string
	Title      string
	Desc       string
	Labels     []label.Label
	Assignee   user.User
	Priority   prio.Prio
	Team       team.Team
	State      state.State
	Project    project.Project
}

func IssuesToRows(issues []Issue, focused bool) []*table.Row {
	rows := make([]*table.Row, 0)

	for _, issue := range issues {
		titleNormal := text.Colored(issue.Title, color.Focusable("#eee", "#888").Darken(0.2))
		titleSelected := text.Colored(issue.Title, color.Focusable("#eee", "#888").Brighten(0.2))

		var projectNormal, projectSelected text.Focusable
		if issue.Project.Name != "" {
			projectNormal = text.Colored(issue.Project.Name, color.Focusable(issue.Project.Color, "#888"))
			projectSelected = text.Colored(issue.Project.Name, color.Focusable(issue.Project.Color, "#888").Brighten(0.2))
		} else {
			projectNormal = text.Colored("", color.Focusable("#555", "#555"))
			projectSelected = text.Colored("", color.Focusable("#555", "#555").Brighten(0.2))
		}

		var assigneeNormal, assigneeSelected text.Focusable
		if issue.Assignee.IsMe {
			assigneeNormal = text.Colored("me", color.Focusable("#76946A", "#888"))
			assigneeSelected = text.Colored("me", color.Focusable("#76946A", "#888").Brighten(0.2))
		} else {
			assigneeNormal = text.Colored(issue.Assignee.DisplayName, color.Focusable("#888", "#888"))
			assigneeSelected = text.Colored(issue.Assignee.DisplayName, color.Focusable("#888", "#888").Brighten(0.2))
		}

		stateNormal := text.Colored(issue.State.Name, color.Focusable(issue.State.Color, "#888"))
		stateSelected := text.Colored(issue.State.Name, color.Focusable(issue.State.Color, "#888").Brighten(0.2))

		priorityNormal := issue.Priority.ToFocusable(0)
		prioritySelected := issue.Priority.ToFocusable(0.2)

		teamNormal := text.Colored(issue.Team.Name, color.Focusable(issue.Team.Color, "#888"))
		teamSelected := text.Colored(issue.Team.Name, color.Focusable(issue.Team.Color, "#888").Brighten(0.2))

		var labelsNormal, labelsSelected []string
		for _, label := range issue.Labels {
			normal := text.Chip(label.Name, color.Focusable("#eee", "#888"), color.Simple(label.Color).Darken(0.5))
			selected := text.Chip(label.Name, color.Focusable("#eee", "#888").Brighten(0.2), color.Simple(label.Color).Darken(0.3))

			if focused {
				labelsNormal = append(labelsNormal, normal.Focused())
				labelsSelected = append(labelsSelected, selected.Focused())
			} else {
				labelsNormal = append(labelsNormal, normal.Blurred())
				labelsSelected = append(labelsSelected, selected.Blurred())
			}
		}

		labelsNormalT := text.Plain(strings.Join(labelsNormal, ""))
		labelsSelectedT := text.Plain(strings.Join(labelsSelected, ""))

		row := &table.Row{
			Identifier: issue.ID,
			Items: []table.RowItem{
				{
					Normal:   projectNormal,
					Selected: projectSelected,
				},
				{
					Normal:   titleNormal,
					Selected: titleSelected,
				},
				{
					Normal:   assigneeNormal,
					Selected: assigneeSelected,
				},
				{
					Normal:   stateNormal,
					Selected: stateSelected,
				},
				{
					Normal:   priorityNormal,
					Selected: prioritySelected,
				},
				{
					Normal:   teamNormal,
					Selected: teamSelected,
				},
				{
					Normal:   labelsNormalT,
					Selected: labelsSelectedT,
				},
			},
		}

		rows = append(rows, row)
	}

	return rows
}

func toString(n *string) string {
	if n == nil {
		return ""
	}

	return *n
}

func FromLinearClientGetIssues(resp linearClient.GetIssues) []Issue {
	var issues []Issue

	for _, iss := range resp.Issues.GetNodes() {
		labels := make([]label.Label, 0, len(iss.Labels.Nodes))

		for _, l := range iss.Labels.Nodes {
			labels = append(labels, label.Label{
				Name:  l.Name,
				Color: l.Color,
			})
		}

		is := Issue{
			ID:         iss.ID,
			Identifier: iss.Identifier,
			Title:      iss.Title,
			Desc:       toString(iss.Description),
			Assignee: user.User{
				ID:          iss.GetAssignee().ID,
				DisplayName: iss.GetAssignee().DisplayName,
				Email:       iss.GetAssignee().Email,
				IsMe:        iss.GetAssignee().IsMe,
			},
			Labels:   labels,
			Priority: prio.Prio(iss.Priority),
			Team: team.Team{
				ID:    iss.Team.ID,
				Name:  iss.Team.Name,
				Color: toString(iss.Team.Color),
			},
			State: state.State{
				Name:     iss.State.Name,
				Color:    iss.State.Color,
				Position: int(iss.State.Position),
			},
			Project: project.Project{
				Name:  iss.GetProject().GetName(),
				Color: iss.GetProject().GetColor(),
			},
		}

		issues = append(issues, is)
	}

	return issues
}
