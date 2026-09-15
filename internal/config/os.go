package config

import (
	"os"

	"gitlab.com/uniget-org/cli/pkg/logging"
	myos "gitlab.com/uniget-org/cli/pkg/os"
)

func (configuration *Config) AssertWritableTarget() {
	myos.AssertWritableDirectory(configuration.GetTarget())
}

func (configuration *Config) AssertLibDirectory() {
	if !myos.DirectoryExists(configuration.GetLibRoot()) {
		myos.AssertDirectory(configuration.GetLibRoot())
	}
	myos.AssertWritableDirectory(configuration.GetLibRoot())
	myos.AssertDirectory(configuration.GetLibDirectory())
}

func (configuration *Config) AssertCacheDirectory() {
	if !myos.DirectoryExists(configuration.GetCacheRoot()) {
		myos.AssertDirectory(configuration.GetCacheRoot())
	}
	myos.AssertWritableDirectory(configuration.GetCacheRoot())
	myos.AssertDirectory(configuration.GetCacheDirectory())
}

func (configuration *Config) AssertMetadataFileExists() {
	_, err := os.Stat(configuration.GetMetadataFile())
	if err != nil {
		logging.Error.Printfln("Metadata file %s does not exist: %s",
			configuration.Prefix+"/"+configuration.GetMetadataFile(),
			err)
		os.Exit(1)
	}

	_, err = os.Stat(configuration.GetMetadataFile() + ".sigstore.json")
	if err != nil {
		logging.Error.Printfln("Metadata signature %s does not exist: %s",
			configuration.Prefix+"/"+configuration.GetMetadataFile()+".sigstore.json",
			err)
		os.Exit(1)
	}
}
