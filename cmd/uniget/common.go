package main

import (
	"os"

	goversion "github.com/hashicorp/go-version"
	"github.com/spf13/viper"
	"gitlab.com/uniget-org/cli/pkg/logging"
	"gitlab.com/uniget-org/cli/pkg/tool"
	"golang.org/x/sys/unix"
)

func checkClientVersionRequirement(tool *tool.Tool) {
	if version == "main" {
		logging.Warning.Printfln("You are running an unreleased version of uniget. Cannot check client version requirement for %s", tool.Name)
		return
	}

	var requiredCliVersion = "0.0.0"
	for schemaVersion, cliVersion := range minimumCliVersionForSchemaVersion {
		if tool.SchemaVersion > schemaVersion {
			requiredCliVersion = cliVersion
		}
	}

	logging.Debugf("Checking if client version %s is at least %s", version, requiredCliVersion)

	v1, err := goversion.NewVersion(requiredCliVersion)
	if err != nil {
		panic(err)
	}
	v2, err := goversion.NewVersion(version)
	if err != nil {
		panic(err)
	}

	if v1.GreaterThan(v2) {
		logging.Error.Printfln("The tool %s requires at least version %s but you have %s", tool.Name, requiredCliVersion, version)
		os.Exit(1)
	}
}

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
	assertWritableDirectory(viper.GetString("prefix") + "/" + viper.GetString("target"))
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
	if !directoryExists(viper.GetString("prefix") + "/" + libRoot) {
		assertDirectory(viper.GetString("prefix") + "/" + libRoot)
	}
	assertWritableDirectory(viper.GetString("prefix") + "/" + libRoot)
	assertDirectory(viper.GetString("prefix") + "/" + libDirectory)
}

func assertCacheDirectory() {
	if !directoryExists(viper.GetString("prefix") + "/" + cacheRoot) {
		assertDirectory(viper.GetString("prefix") + "/" + cacheRoot)
	}
	assertWritableDirectory(viper.GetString("prefix") + "/" + cacheRoot)
	assertDirectory(viper.GetString("prefix") + "/" + cacheDirectory)
}

func assertMetadataFileExists() {
	_, err := os.Stat(viper.GetString("prefix") + "/" + metadataFile)
	if err != nil {
		logging.Error.Printfln("Metadata file %s does not exist: %s", viper.GetString("prefix")+"/"+metadataFile, err)
		os.Exit(1)
	}
}

func assertMetadataIsLoaded() {
	if len(tools.Tools) == 0 {
		logging.Error.Printfln("Metadata is not loaded")
		os.Exit(1)
	}
}
