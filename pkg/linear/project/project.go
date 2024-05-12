package project

import linearClient "github.com/sayedmurtaza24/tinear/linear"

type Project struct {
	Name  string
	Color string
}

func FromLinearClientGetProjects(resp linearClient.GetProjects) []Project {
	var projects []Project

	for _, proj := range resp.Projects.GetNodes() {
		projects = append(projects, Project{
			Name:  proj.Name,
			Color: proj.Color,
		})
	}

	return projects
}
