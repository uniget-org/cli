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

## Purpose

`uniget` is inspired by the [convenience script](https://docs.docker.com/engine/install/ubuntu/#install-using-the-convenience-script) to install the Docker daemon. But the scope is much larger.

`uniget` is meant to bootstrap a new box with Docker as well as install useful tools from the container ecosystem and beyond. It can also be used to update these tools. It aims to be distribution-agnostic and provide reasonable default configurations. Personally, I am using it to prepare virtual machines for my own experiments as well as training environments.

Tools are downloaded, installed and updated automatically.

## Quickstart

Download and run `uniget`:

```bash
curl -sLf https://github.com/uniget-org/uniget/releases/latest/download/uniget_linux_$(uname -m).tar.gz | \
sudo tar -xzC /usr/local/bin uniget
```

## Documentation

See [docs](docs) for the complete documentation.
