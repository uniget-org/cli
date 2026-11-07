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
	HooksRootDirectory          = "hooks"
	HooksPreInstallDirectory    = HooksRootDirectory + "/pre-install.d"
	HooksPostInstallDirectory   = HooksRootDirectory + "/post-install.d"
	HooksPreUninstallDirectory  = HooksRootDirectory + "/pre-uninstall.d"
	HooksPostUninstallDirectory = HooksRootDirectory + "/post-uninstall.d"
	Registry                    = "ghcr.io"
	Organization                = "uniget-org"
	ImageRepository             = Organization + "/tools"
	ToolSeparator               = "/"
	RegistryImagePrefix         = Registry + "/" + ImageRepository + ToolSeparator
	SigstoreIssuer              = "https://token.actions.githubusercontent.com"
	SigstoreSubjectRegexp       = "https://github\\.com/uniget-org/tools/\\.github/workflows/[^.]+\\.yml@.+"
)
