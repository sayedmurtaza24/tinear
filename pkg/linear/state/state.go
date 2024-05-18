package state

type State struct {
	Name     string
	Color    string
	Position int
}

func position(name string, pos int) int {
	switch name {
	case "Canceled":
		return 0
	case "Done":
		return 1
	case "Triage":
		return 2
	case "Backlog":
		return 3
	case "QA Ready":
		return 4
	case "In Review":
		return 5
	case "Todo":
		return 6
	case "In Progress":
		return 7
	default:
		return pos
	}
}

func New(name, color string, pos int) State {
	return State{
		Name:     name,
		Color:    color,
		Position: position(name, pos),
	}
}
