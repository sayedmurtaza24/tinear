package hover

import (
	"log"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/sayedmurtaza24/tinear/pkg/store"
	"github.com/sayedmurtaza24/tinear/pkg/ui/color"
	"github.com/sayedmurtaza24/tinear/pkg/ui/text"
)

func HoverIssue(issue store.Issue, width, maxHeight int, focus bool) string {
	const (
		labelProject   = "project:      "
		labelTeam      = "team:         "
		labelAssignee  = "assignee:     "
		labelCreatedAt = "created at:   "
		labelUpdatedAt = "updated at:   "
	)

	label := func(s string) string {
		t := text.Colored(s, color.Focusable("#aaa", "#444"))

		if focus {
			return t.Focused()
		} else {
			return t.Blurred()
		}
	}

	colored := func(s string, c string, placeholder string, opts ...text.Opt) string {
		t := text.Colored(s, color.Simple(c), opts...)
		if s == "" {
			t = text.Colored(placeholder, color.Simple("#888"), opts...)
		}
		if focus {
			return t.Focused()
		} else {
			return t.Blurred()
		}
	}

	chip := func(s string, fg string, bg string) string {
		t := text.Chip(s, color.Simple(fg), color.Simple(bg))

		if focus {
			return t.Focused()
		} else {
			return t.Blurred()
		}
	}

	assigneeColor := "#888"
	if issue.Assignee.IsMe {
		assigneeColor = "#76946A"
	}

	parsed, err := issue.Labels.Parse()
	if err != nil {
		log.Println(err)
	}

	var labels []text.Focusable
	for _, label := range parsed {
		labels = append(labels, text.Chip(
			label.Name,
			color.Focusable("#eee", "#888"),
			color.Focusable(label.Color, "#444").Darken(0.5),
		))
	}

	var labelsStr string
	if len(labels) > 0 {
		labelsStr = lipgloss.JoinVertical(
			lipgloss.Left,
			"",
			text.Joined(" ", labels...).Focused(),
			"",
		)
	}

	topBar := lipgloss.JoinVertical(
		lipgloss.Left,
		chip(issue.State.Name, "#222", issue.State.Color)+" "+colored(issue.Title, "#eee", "No state", text.B),
		labelsStr,
		label(labelProject)+colored(issue.Project.Name, issue.Project.Color, "No project"),
		label(labelTeam)+colored(issue.Team.Name, issue.Team.Color, "No team"),
		label(labelAssignee)+colored(issue.Assignee.DisplayName, assigneeColor, "No assignee"),
		label(labelCreatedAt)+colored(issue.CreatedAt.Format(time.RFC822), "#ddd", ""),
		label(labelUpdatedAt)+colored(issue.UpdatedAt.Format(time.RFC822), "#ddd", ""),
	)

	topBar = lipgloss.NewStyle().
		Padding(0, 2).Render(topBar)

	var description string

	if issue.Desc != "" {
		r, _ := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(width-8),
		)

		description, _ = r.Render(issue.Desc)
	}

	maxH := lipgloss.NewStyle().MaxHeight(maxHeight - 5).Render

	s := lipgloss.
		NewStyle().
		Width(width).
		Padding(1, 1, 0).
		Border(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color("#444")).
		Render

	return s(maxH(lipgloss.JoinVertical(
		lipgloss.Left,
		topBar,
		description,
	)))
}
