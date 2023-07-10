package main

import (
	"os"
	"runtime"

	"golang.org/x/sys/unix"

	"github.com/pterm/pterm"

	"github.com/uniget-org/cli/pkg/tool"
)

var altArch string = runtime.GOARCH
var arch string

var prefix = ""
var target = "usr/local"
var cacheRoot = "var/cache"
var cacheDirectory = cacheRoot + "/uniget"
var libRoot = "var/lib"
var libDirectory = libRoot + "/uniget"
var user = false
var metadataFileName = "metadata.json"
var metadataFile = cacheDirectory + "/" + metadataFileName
var registry = "ghcr.io"
var projectRepository = "uniget-org/uniget"
var repository = "nicholasdille/docker-setup"
var toolSeparator = "/"
var registryImagePrefix = registry + "/" + repository + toolSeparator
var tools tool.Tools
var noInteractive bool

func directoryExists(directory string) bool {
	pterm.Debug.Printfln("Checking if directory %s exists", directory)
	_, err := os.Stat(directory)
	return err == nil
}

func fileExists(file string) bool {
	pterm.Debug.Printfln("Checking if file %s exists", file)
	_, err := os.Stat(file)
	return err == nil
}

func directoryIsWritable(directory string) bool {
	pterm.Debug.Printfln("Checking if directory %s is writable", directory)
	return unix.Access(directory, unix.W_OK) == nil
}

func assertWritableDirectory(directory string) {
	if !directoryExists(directory) {
		assertDirectory(directory)
	}
	if !directoryIsWritable(directory) {
		pterm.Error.Printfln("Directory %s is not writable", directory)
		os.Exit(1)
	}
}

func assertWritableTarget() {
	assertWritableDirectory(prefix + "/" + target)
}

func assertDirectory(directory string) {
	pterm.Debug.Printfln("Creating directory %s", directory)
	err := os.MkdirAll(directory, 0755)
	if err != nil {
		pterm.Error.Printfln("Error creating directory %s: %s", directory, err)
		os.Exit(1)
	}
}

func assertLibDirectory() {
	if !directoryExists(prefix + "/" + libRoot) {
		assertDirectory(prefix + "/" + libRoot)
	}
	assertWritableDirectory(prefix + "/" + libRoot)
	assertDirectory(prefix + "/" + libDirectory)
}

func assertCacheDirectory() {
	if !directoryExists(prefix + "/" + cacheRoot) {
		assertDirectory(prefix + "/" + cacheRoot)
	}
	assertWritableDirectory(prefix + "/" + cacheRoot)
	assertDirectory(prefix + "/" + cacheDirectory)
}

func assertMetadataFileExists() {
	_, err := os.Stat(prefix + "/" + metadataFile)
	if err != nil {
		pterm.Error.Printfln("Metadata file %s does not exist: %s", prefix+"/"+metadataFile, err)
		os.Exit(1)
	}
}

func assertMetadataIsLoaded() {
	if len(tools.Tools) == 0 {
		pterm.Error.Printfln("Metadata is not loaded")
		os.Exit(1)
	}
}

func initDockerSetup() {
	if altArch == "amd64" {
		arch = "x86_64"

	} else if altArch == "arm64" {
		arch = "aarch64"

	} else {
		pterm.Error.Printfln("Unsupported architecture: %s", arch)
		os.Exit(1)
	}
}
