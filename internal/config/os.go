package config

import (
	"os"

	"gitlab.com/uniget-org/cli/pkg/logging"
	myos "gitlab.com/uniget-org/cli/pkg/os"
)

func (configuration *Config) AssertWritableTarget() {
	myos.AssertWritableDirectory(configuration.Prefix + "/" + configuration.Target)
}

func (configuration *Config) AssertLibDirectory() {
	if !myos.DirectoryExists(configuration.Prefix + "/" + configuration.LibRoot) {
		myos.AssertDirectory(configuration.Prefix + "/" + configuration.LibRoot)
	}
	myos.AssertWritableDirectory(configuration.Prefix + "/" + configuration.LibRoot)
	myos.AssertDirectory(configuration.Prefix + "/" + configuration.GetLibDirectory())
}

func (configuration *Config) AssertCacheDirectory() {
	if !myos.DirectoryExists(configuration.Prefix + "/" + configuration.CacheRoot) {
		myos.AssertDirectory(configuration.Prefix + "/" + configuration.CacheRoot)
	}
	myos.AssertWritableDirectory(configuration.Prefix + "/" + configuration.CacheRoot)
	myos.AssertDirectory(configuration.Prefix + "/" + configuration.GetCacheDirectory())
}

func (configuration *Config) AssertMetadataFileExists() {
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
