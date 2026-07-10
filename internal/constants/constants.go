package constants

import _ "embed"

var (
	//go:embed logo.txt
	Header string
	Slogan string = "The universal installer and updater for (container) tools" + "\n" +
		"                                       https://uniget.dev"
)

const (
	ProjectName                 = "uniget"
	MetadataFileName            = "metadata.json"
	MetadataImageTag            = "main"
	HooksPreInstallDirectory    = "hooks/pre-install.d"
	HooksPostInstallDirectory   = "hooks/post-install.d"
	HooksPreUninstallDirectory  = "hooks/pre-uninstall.d"
	HooksPostUninstallDirectory = "hooks/post-uninstall.d"
	Registry                    = "ghcr.io"
	Organization                = "uniget-org"
	ImageRepository             = Organization + "/tools"
	ToolSeparator               = "/"
	RegistryImagePrefix         = Registry + "/" + ImageRepository + ToolSeparator
)
