package main

import (
	"fmt"
	"os"
	"runtime"

	log "github.com/sirupsen/logrus"

	"github.com/nicholasdille/docker-setup/pkg/tool"
)

var alt_arch string = runtime.GOARCH
var arch string

var prefix = "/"
var target = "usr/local"
var cacheDirectory = "/var/cache/docker-setup"
var libDirectory = "/var/lib/docker-setup"
var metadataFileName = "metadata.json"
var metadataFile = cacheDirectory + "/" + metadataFileName
var registry = "ghcr.io"
var repository = "nicholasdille/docker-setup"
var toolSeparator = "/"
var registryImagePrefix = registry + "/" + repository + toolSeparator
var tools tool.Tools

var emoji_tool = "\U0001F528"

// TODO: variables for registry

func initDockerSetup() {
	var err error

	if os.Geteuid() != 0 {
		fmt.Printf("Use must use sudo\n")
		os.Exit(1)
	}

	if alt_arch == "amd64" {
		arch = "x86_64"

	} else if alt_arch == "arm64" {
		arch = "aarch64"

	} else {
		log.Errorf("Unsupported architecture: %s", arch)
		os.Exit(1)
	}

	os.MkdirAll(cacheDirectory, 0755)
	os.MkdirAll(libDirectory, 0755)

	_, err = os.Stat(metadataFile)
	if err == nil {
		tools, err = tool.LoadFromFile(metadataFile)
		if err != nil {
			fmt.Printf("Error loading metadata from file %s: %s\n", metadataFile, err)
			os.Exit(1)
		}
	}
}