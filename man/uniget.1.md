% uniget "1"

# NAME
**uniget** - The universal installer and updater for (container) tools

# SYNOPSIS

**uniget** [_global-option_ ...] _command_ [_command-option_ ...] [_argument_ ...]

# DESCRIPTION

uniget is a command line tool for installing and updating tools. It is the last
binary you need to install. It included a reasonable default configuration where
applicable.

By default, uniget requires root permissions to install tools. By using --user,
uniget installs in the user home directory.

# COMMANDS

**completion**
: Generate command completion script.

**cron**
: MAnage cron job for updating tool definitions and upgrading tools.

**describe**
: Display information about a specific tool.

**generate**
: Display a Dockerfile for a configurable base image including the specified
tools.

**healthcheck**
: Check installed tool.

**help**
: Display help.

**inspect**
: Display files shipped with a specific tool.

**install**
: Install a tool.

**list**
: List available or installed tools.

**message**
: Display messages for a specific tool.

**postinstall**
: Run post installation script for a specific tool.

**search**
: Search for a tool in name, description and tags.

**self-upgrade**
: Upgrade uniget to the latest version.

**tags**
: Display tags used by tools.

**uninstall**
: Uninstall a tool.

**update**
: Update tool definitions.

**upgrade**
: Upgrade all tools.

**version**
: Display the version of an installed tool.

# GLOBAL OPTIONS

These options can be used with any command, and must precede the **command**.

**--debug**|**-d**
: Enable debug logging.

**--help**|**-h**
: Display help.

**--no-interactive**
: Disable interactive menus after some commands.

**--prefix**|**-p** _path_
: Set base directory to prepare chroot environments. Defaults to /.

**--target**|**-t** _path_
: Set target directory relative to prefix. Defaults to usr/local.

**--trace**
: Enable trace logging.

**--user**
: Enable installation in user context.

**--version**|**-v**
: Show version of uniget.

# SEE ALSO

Homepage: https://uniget.dev

Documentation: https://docs.uniget.dev

Index of included tools: https://tools.uniget.dev

Repository for CLI: https://github.com/uniget-org/cli

Repository for tools: https://github.com/uniget-org/tools
