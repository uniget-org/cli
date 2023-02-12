package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func initGenerateCmd() {
	rootCmd.AddCommand(generateCmd)
}

var generateCmd = &cobra.Command{
	Use:     "generate",
	Short:   "Generate new tool definition",
	Long:    header + "\nGenerate new tool definition",
	Args:    cobra.NoArgs,
	Run:     func(cmd *cobra.Command, args []string) {
		fmt.Println(`

tools:

# Predefined variables:
# ${name}       Name of the tool
# ${version}    Version of the tool
# ${binary}     Path and name of binary
# ${arch}       x86_64 or aarch64
# ${alt_arch}   amd64 or arm64

- name: foo
  version: 1.2.3

  # Optional:
  # Name of binary if it differs from the name
  # Relative paths will be prepended with ${target}/bin
  #binary: bar

  # Optional:
  # Version check output must match version field
  #check: ${binary} --version | cut -d' ' -f2

  # Optional:
  # Specified flags must set on the command line like --flag-foo
  # For every flag, a matching not-foo is created to define conflicts
  # See docker-compose and docker-compose-v1
  #flags:
  #- foo

  # Optional:
  # Other tools which must be installed before this one
  #needs:
  #- docker

  # Tags to catagorize tool into
  # Tools tagged with "default" will be installed if nothing else is specified
  tags:
  - docker

  # Specify which resources to download and where to place them
  # Resources require an URL which can be:
  # - A template if the URL can be used for both platforms
  # - Selarate URLs for amd64 and optionally arm64
  download:
  - url:

      # Use template if possible
      template: https://someserver.domain.com/foo/${version}/file-${alt_arch}.tar.gz
	  # Alternative to template
	  #x86_64: https://someserver.domain.com/foo/${version}/file.tar.gz
	  # Optional:
	  # Specify URL for arm64
	  #aarch64: https://someserver.domain.com/foo/${version}/file-arm64.tar.gz

	  # Type of resource
	  # - tarball
	  # - executable
	  # - zip
	  type: tarball

	  # Optional:
	  # Where to install files to
	  #path: ${target}/bin

	  # Optional:
	  # How many components to strip from path
	  #strip: 1

	  # Optional:
	  # Which files to extract
	  #files:
	  #- foo

  - url:
      template: https://someserver.domain.com/bar/${version}/file-${alt_arch}
	  type: executable
	  # Optional:
	  # Where to install files to
	  #path: ${target}/bin/blarg

  - url:
      template: https://someserver.domain.com/baz/${version}/file-${alt_arch}.zip
	  type: zip
	  # Mandatory: Specify which files to extract
	  files:
	  - baz
	  # Optional:
	  # Where to install files to
	  #path: ${target}/bin/blubb
	  # strip is only supported for tarball

  # Alternative to "download" if an installation script is required
  #install: |
  #  printenv | sort

  # Optional:
  # Commands to execute after "download" or "install"
  #post_install: |
  #  printenv | sort
		`)
	},
}
