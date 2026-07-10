package main

import (
	"os"

	"github.com/pterm/pterm"
	"gitlab.com/uniget-org/cli/pkg/logging"
	myos "gitlab.com/uniget-org/cli/pkg/os"
	"gitlab.com/uniget-org/cli/pkg/tui"
	"golang.org/x/sys/unix"
)

func directoryExists(directory string) bool {
	logging.Debugf("Checking if directory %s exists", directory)
	_, err := os.Stat(directory)
	return err == nil
}

func fileExists(file string) bool {
	logging.Debugf("Checking if file %s exists", file)
	_, err := os.Stat(file)
	return err == nil
}

func directoryIsWritable(directory string) bool {
	logging.Debugf("Checking if directory %s is writable", directory)
	return unix.Access(directory, unix.W_OK) == nil
}

func assertWritableDirectory(directory string) {
	if !directoryExists(directory) {
		assertDirectory(directory)
	}
	if !directoryIsWritable(directory) {
		logging.Error.Printfln("Directory %s is not writable", directory)
		os.Exit(1)
	}
}

func assertWritableTarget() {
	assertWritableDirectory(configuration.Prefix + "/" + configuration.Target)
}

func assertDirectory(directory string) {
	logging.Debugf("Creating directory %s", directory)
	err := os.MkdirAll(directory, 0755) // #nosec G301 -- Directories will contain public information
	if err != nil {
		logging.Error.Printfln("Error creating directory %s: %s", directory, err)
		os.Exit(1)
	}
}

func assertLibDirectory() {
	if !directoryExists(configuration.Prefix + "/" + configuration.LibRoot) {
		assertDirectory(configuration.Prefix + "/" + configuration.LibRoot)
	}
	assertWritableDirectory(configuration.Prefix + "/" + configuration.LibRoot)
	assertDirectory(configuration.Prefix + "/" + configuration.GetLibDirectory())
}

func assertCacheDirectory() {
	if !directoryExists(configuration.Prefix + "/" + configuration.CacheRoot) {
		assertDirectory(configuration.Prefix + "/" + configuration.CacheRoot)
	}
	assertWritableDirectory(configuration.Prefix + "/" + configuration.CacheRoot)
	assertDirectory(configuration.Prefix + "/" + configuration.GetCacheDirectory())
}

func assertMetadataFileExists() {
	_, err := os.Stat(configuration.Prefix + "/" + configuration.GetMetadataFile())
	if err != nil {
		logging.Error.Printfln("Metadata file %s does not exist: %s", configuration.Prefix+"/"+configuration.GetMetadataFile(), err)
		os.Exit(1)
	}

	_, err = os.Stat(configuration.Prefix + "/" + configuration.GetMetadataFile() + ".sigstore.json")
	if err != nil {
		logging.Error.Printfln("Metadata signature %s does not exist: %s", configuration.Prefix+"/"+configuration.GetMetadataFile()+".sigstore.json", err)
		os.Exit(1)
	}
}

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
