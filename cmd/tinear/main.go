package main

import (
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sayedmurtaza24/tinear/cmd/tinear/show"
	"github.com/sayedmurtaza24/tinear/pkg/client"
	"github.com/sayedmurtaza24/tinear/pkg/store"
)

func main() {
	f, err := tea.LogToFile("/tmp/tinear.log", "DEBUG")
	if err != nil {
		slog.Error("failed to setup logger", slog.Any("error", err))
		return
	}
	defer f.Close()

	store, err := store.New("/tmp/tinear")
	if err != nil {
		slog.Error("failed to setup store", slog.Any("error", err))
		return
	}

	client := client.New()
	model := show.New(store, client)

	_, err = tea.NewProgram(model, tea.WithAltScreen()).Run()
	if err != nil {
		slog.Error("failed to start tinear", slog.Any("error", err))
	}
}
