package main

import (
	"log"
	"log/slog"

	// _ "net/http/pprof"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sayedmurtaza24/tinear/cmd/tinear/show"
	"github.com/sayedmurtaza24/tinear/pkg/common"
	"github.com/sayedmurtaza24/tinear/pkg/keymap"
	"github.com/sayedmurtaza24/tinear/pkg/screen"
)

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
	model := show.New(common)

	_, err = tea.NewProgram(model, tea.WithAltScreen()).Run()
	if err != nil {
		slog.Error("failed to start tinear", slog.Any("error", err))
	}
}
