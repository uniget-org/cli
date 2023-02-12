package main

import (
	"fmt"
	"net/http"
	"io"
	"os"
	"regexp"
	"runtime"
	
	log "github.com/sirupsen/logrus"

	"github.com/nicholasdille/docker-setup/pkg/tool"
)

var alt_arch string = runtime.GOARCH
var arch string
var cacheDirectory = "/var/cache/docker-setup"
var downloadDirectory = cacheDirectory + "/downloads"
var toolsFileName = cacheDirectory + "/metadata.json"
var tools tool.Tools

func get_file(filepath string, url string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func initDockerSetup() {
	if alt_arch == "amd64" {
		arch = "x86_64"

	} else if alt_arch == "arm64" {
		arch = "aarch64"

	} else {
		log.Errorf("Unsupported architecture: %s", arch)
		os.Exit(1)
	}

	os.MkdirAll(cacheDirectory, 0755)
	os.MkdirAll(downloadDirectory, 0755)

	versionPath := "releases/download/v" + version + "/metadata.json"
	match, err := regexp.MatchString("^v[0-9]+", version)
	if err == nil && ! match {
		versionPath = "raw/" + version + "/metadata.json"
	}

	err = get_file(toolsFileName, "https://github.com/nicholasdille/docker-setup/" + versionPath)
	if err != nil {
		fmt.Printf("Error downloading metadata.json from %s: %s", versionPath, err)
		os.Exit(1)
	}

	tools, err = tool.LoadFromFile(toolsFileName)
	if err != nil {
		fmt.Printf("Error loading from file %s: %s\n", toolsFileName, err)
		os.Exit(1)
	}
}