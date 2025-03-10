[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/7946/badge)](https://www.bestpractices.dev/projects/7946) [![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/uniget-org/cli/badge)](https://securityscorecards.dev/viewer/?uri=github.com/uniget-org/cli)

# uniget

```plaintext
             _            _
 _   _ _ __ (_) __ _  ___| |_
| | | | '_ \| |/ _` |/ _ \ __|
| |_| | | | | | (_| |  __/ |_
 \__,_|_| |_|_|\__, |\___|\__|
               |___/
```

The universal installer and updater to (container) tools

⚠️ uniget is in the process of being migrated [to GitLab](https://gitlab.com/uniget-org). The migration is being tracked [here](https://gitlab.com/uniget-org/migration-gh2gl). There is only one thing, you need to do: regularly update the uniget CLI using `uniget self-update`. ⚠️

## Purpose

`uniget` is inspired by the [convenience script](https://docs.docker.com/engine/install/ubuntu/#install-using-the-convenience-script) to install the Docker daemon. But the scope is much larger.

`uniget` is meant to bootstrap a new box with Docker as well as install useful tools from the container ecosystem and beyond. It can also be used to update these tools. It aims to be distribution-agnostic and provide reasonable default configurations. Personally, I am using it to prepare virtual machines for my own experiments as well as training environments.

Tools are downloaded, installed and updated automatically.

## Quickstart

Download and run `uniget`:

```bash
curl -sLf https://github.com/uniget-org/cli/releases/latest/download/uniget_linux_$(uname -m).tar.gz \
| sudo tar -xzC /usr/local/bin uniget
```

## Docs

See the [documentation site](https://docs.uniget.dev).

## Quickstart

The `uniget` CLI comes with help included. The following scenarios are meant as quickstart tutorials.

## You want the default set of tools

By default, `uniget` will only install a small set of tools.

```bash
uniget install --default
```

### You want to investigate which tools are available

List which tools are available in `uniget`:

```bash
uniget list
```

### You want to install a specific tool

It is possible to install individual tools:

```bash
uniget install gojq
uniget install kubectl helm
```

### You want to search for tools

You can search for the specified term in names, tags and dependencies:

```bash
uniget search jq
```

### You want to update installed tools

Updated tools which are already installed:

```bash
uniget update
uniget upgrade
```

### You want to see what will happen

Show which tools will be processed and updated:

```bash
uniget install containerd --plan
uniget upgrade --plan
```

### Reinstall tool(s)

By adding the `--reinstall` parameter, the selected tools can be reinstalled regardless if they are outdated:

```bash
uniget install gojq --reinstall
```
