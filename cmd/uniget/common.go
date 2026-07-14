package main

import (
	"os"

	"github.com/pterm/pterm"
	"gitlab.com/uniget-org/cli/pkg/logging"
	myos "gitlab.com/uniget-org/cli/pkg/os"
	"gitlab.com/uniget-org/cli/pkg/tui"
)

func assertMetadataIsLoaded() {
	if len(tools.Tools) == 0 {
		logging.Error.Printfln("Metadata is not loaded")
		os.Exit(1)
	}
}

func createProgressReader(title string) tui.ProgressReader {
	progressReader := tui.NewProgressReader(nil, nil)

	if myos.IsTty() && !configuration.Debug && !configuration.Trace {
		progressPrinter, err := pterm.DefaultProgressbar.
			WithTitle(title).
			WithTotal(0).
			WithRemoveWhenDone().
			WithShowElapsedTime(false).
			Start()
		if err == nil {
			progressReader = tui.NewProgressReader(
				func(n int64) {
					progressPrinter.Total = int(n)
				},
				func(n int64) {
					progressPrinter.Add(int(n))
				},
			)
		}
	}

	return progressReader
}
