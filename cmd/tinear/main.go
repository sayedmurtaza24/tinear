package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	// _ "net/http/pprof"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sayedmurtaza24/tinear/cmd/tinear/show"
	linearClient "github.com/sayedmurtaza24/tinear/linear"
	"github.com/sayedmurtaza24/tinear/pkg/common"
	"github.com/sayedmurtaza24/tinear/pkg/keymap"
	"github.com/sayedmurtaza24/tinear/pkg/screen"
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
	// go func() {
	// 	log.Fatal(http.ListenAndServe(":3000", nil))
	// }()
	//
	f, err := tea.LogToFile("tinear.log", "DEBUG")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	keymap := keymap.NewDefault()
	size := screen.NewSize(0, 0)
	common := common.New(keymap, size)
	client := initLinearClient()
	model := show.New(common, client)

	_, err = tea.NewProgram(model, tea.WithAltScreen()).Run()
	if err != nil {
		slog.Error("failed to start tinear", slog.Any("error", err))
	}
}
