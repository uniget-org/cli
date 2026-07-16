package common

import (
	"github.com/pterm/pterm"
	myos "gitlab.com/uniget-org/cli/pkg/os"
	"gitlab.com/uniget-org/cli/pkg/tui"
)

func CreateProgressReader(title string, suppress bool) tui.ProgressReader {
	progressReader := tui.NewProgressReader(nil, nil)

	if myos.IsTty() && !suppress {
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
