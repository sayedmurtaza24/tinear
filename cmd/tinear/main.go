package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sayedmurtaza24/tinear/cmd/tinear/show"
	linearClient "github.com/sayedmurtaza24/tinear/linear"
	"github.com/sayedmurtaza24/tinear/pkg/client"
	"github.com/sayedmurtaza24/tinear/pkg/store"
)

func initLinearClient() linearClient.LinearClient {
	const linearBaseUrL = "https://api.linear.app/graphql"

	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		log.Fatal("LINEAR_API_KEY is not set")
	}

	md := linearClient.GetAuthMiddleware(apiKey)

	client := linearClient.NewClient(http.DefaultClient, linearBaseUrL, nil, md)

	return client
}

func main() {
	f, err := tea.LogToFile("tinear.log", "DEBUG")
	if err != nil {
		slog.Error("failed to setup logger", slog.Any("error", err))
	}
	defer f.Close()

	store, err := store.New("db")
	if err != nil {
		slog.Error("failed to setup store", slog.Any("error", err))
	}

	client := client.New(initLinearClient())
	model := show.New(store, client)

	_, err = tea.NewProgram(model, tea.WithAltScreen()).Run()
	if err != nil {
		slog.Error("failed to start tinear", slog.Any("error", err))
	}
}
