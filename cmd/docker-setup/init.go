package main

import (
	"fmt"
	"os"
	"runtime"

	"golang.org/x/sys/unix"

	log "github.com/sirupsen/logrus"

	"github.com/nicholasdille/docker-setup/pkg/tool"
)

var altArch string = runtime.GOARCH
var arch string

var prefix = ""
var target = "usr/local"
var cacheRoot = "var/cache"
var cacheDirectory = cacheRoot + "/docker-setup"
var libRoot = "var/lib"
var libDirectory = libRoot + "/docker-setup"
var user = false
var metadataFileName = "metadata.json"
var metadataFile = cacheDirectory + "/" + metadataFileName
var registry = "ghcr.io"
var repository = "nicholasdille/docker-setup"
var toolSeparator = "/"
var registryImagePrefix = registry + "/" + repository + toolSeparator
var tools tool.Tools

var emojiTool = "\U0001F528"
var emojiRun = "\U0001FE0F"

func directoryExists(directory string) bool {
	log.Tracef("Checking if directory %s exists", directory)
	_, err := os.Stat(directory)
	return err == nil
}

func fileExists(file string) bool {
	log.Tracef("Checking if file %s exists", file)
	_, err := os.Stat(file)
	return err == nil
}

func directoryIsWritable(directory string) bool {
	log.Tracef("Checking if directory %s is writable", directory)
	return unix.Access(directory, unix.W_OK) == nil
}

func assertWritableDirectory(directory string) {
	if !directoryExists(directory) {
		assertDirectory(directory)
	}
	if !directoryIsWritable(directory) {
		log.Errorf("Directory %s is not writable", directory)
		os.Exit(1)
	}
}

func assertWritableTarget() {
	assertWritableDirectory(prefix + "/" + target)
}

func assertDirectory(directory string) {
	log.Tracef("Creating directory %s", directory)
	err := os.MkdirAll(directory, 0755)
	if err != nil {
		fmt.Printf("Error creating directory %s: %s\n", directory, err)
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
		fmt.Printf("Metadata file %s does not exist: %s\n", prefix+"/"+metadataFile, err)
		os.Exit(1)
	}
}

func assertMetadataIsLoaded() {
	if len(tools.Tools) == 0 {
		fmt.Printf("Metadata is not loaded\n")
		os.Exit(1)
	}
}

func initDockerSetup() {
	if altArch == "amd64" {
		arch = "x86_64"

	} else if altArch == "arm64" {
		arch = "aarch64"

	} else {
		log.Errorf("Unsupported architecture: %s", arch)
		os.Exit(1)
	}
}
