package tui

import "github.com/pterm/pterm"

type PtermProgressReader struct {
	ProgressReader
	progressPrinter *pterm.ProgressbarPrinter
}

func NewPtermProgressReader(title string) (PtermProgressReader, error) {
	progressPrinter, err := pterm.DefaultProgressbar.WithTitle(title).WithTotal(0).Start()
	if err != nil {
		return PtermProgressReader{}, err
	}

	ptermProgressReader := PtermProgressReader{
		progressPrinter: progressPrinter,
	}
	ptermProgressReader.onTotalUpdate = func(n int64) {
		ptermProgressReader.SetTotal(n)
	}
	ptermProgressReader.onProgress = func(n int64) {
		progressPrinter.Add(int(n))
	}

	return ptermProgressReader, nil
}
