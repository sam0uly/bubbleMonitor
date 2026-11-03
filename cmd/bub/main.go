package main

import (
	"fmt"
	// "log"
	// "net/http"
	_ "net/http/pprof"
	"os"

	tea "charm.land/bubbletea/v2"
	"samouly.fun/bubbleMonitor/internal/app"
)

func main() {
	// go func() {
	// 	log.Println("pprof listening on http://localhost:8080")
	// 	log.Println(http.ListenAndServe("localhost:8080", nil))
	// }()

	if len(os.Args) > 1 {
		if err := Execute(buildVersion, buildCommit, buildDate); err != nil {
			os.Exit(1)
		}
		return
	}

	p := tea.NewProgram(app.InitialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
