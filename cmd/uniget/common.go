package main

import (
	"os"

	"github.com/pterm/pterm"
	"gitlab.com/uniget-org/cli/pkg/logging"
	myos "gitlab.com/uniget-org/cli/pkg/os"
	"gitlab.com/uniget-org/cli/pkg/tui"
)

func AssertWritableTarget() {
	myos.AssertWritableDirectory(configuration.Prefix + "/" + configuration.Target)
}

func AssertLibDirectory() {
	if !myos.DirectoryExists(configuration.Prefix + "/" + configuration.LibRoot) {
		myos.AssertDirectory(configuration.Prefix + "/" + configuration.LibRoot)
	}
	myos.AssertWritableDirectory(configuration.Prefix + "/" + configuration.LibRoot)
	myos.AssertDirectory(configuration.Prefix + "/" + configuration.GetLibDirectory())
}

func AssertCacheDirectory() {
	if !myos.DirectoryExists(configuration.Prefix + "/" + configuration.CacheRoot) {
		myos.AssertDirectory(configuration.Prefix + "/" + configuration.CacheRoot)
	}
	myos.AssertWritableDirectory(configuration.Prefix + "/" + configuration.CacheRoot)
	myos.AssertDirectory(configuration.Prefix + "/" + configuration.GetCacheDirectory())
}

func AssertMetadataFileExists() {
	_, err := os.Stat(configuration.Prefix + "/" + configuration.GetMetadataFile())
	if err != nil {
		logging.Error.Printfln("Metadata file %s does not exist: %s",
			configuration.Prefix+"/"+configuration.GetMetadataFile(),
			err)
		os.Exit(1)
	}

	_, err = os.Stat(configuration.Prefix + "/" + configuration.GetMetadataFile() + ".sigstore.json")
	if err != nil {
		logging.Error.Printfln("Metadata signature %s does not exist: %s",
			configuration.Prefix+"/"+configuration.GetMetadataFile()+".sigstore.json",
			err)
		os.Exit(1)
	}
}

func AssertMetadataIsLoaded() {
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
