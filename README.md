# docker-setup

```plaintext
     _            _                           _
  __| | ___   ___| | _____ _ __      ___  ___| |_ _   _ _ __
 / _` |/ _ \ / __| |/ / _ \ '__|____/ __|/ _ \ __| | | | '_ \
| (_| | (_) | (__|   <  __/ | |_____\__ \  __/ |_| |_| | |_) |
 \__,_|\___/ \___|_|\_\___|_|       |___/\___|\__|\__,_| .__/
                                                       |_|
```

The container tools installer and updater

## Purpose

`docker-setup` is inspired by the [convenience script](https://docs.docker.com/engine/install/ubuntu/#install-using-the-convenience-script) to install the Docker daemon. But the scope is much larger.

`docker-setup` is meant to bootstrap a new box with Docker as well as install useful tools from the container ecosystem and beyond. It can also be used to update these tools. It aims to be distribution-agnostic and provide reasonable defaults. Personally, I am using it to prepare virtual machines for my own experiments as well as training environments.

Tools are downloaded, installed and updated automatically.

## Version 2

The `main` branch now represents the code for version 2 of `docker-setup`. Please refer to the milestone [v2.0.0](https://github.com/nicholasdille/docker-setup/milestone/10) to track the progress.

While version 1 was a huge bash script to install tools defined in `tools.yaml`, version 2 stores tools in container images. From there they can be added to a container image or installed locally. See the [documentation](docs) for details about the new concept.

For details about version 1 of `docker-setup` please refer to the [last stable release v1.7](https://github.com/nicholasdille/docker-setup/tree/v1.7).

## Quickstart

Download and run `docker-setup`:

```bash
curl --silent --location --remote-name https://github.com/nicholasdille/docker-setup/releases/latest/download/docker-setup
bash docker-setup
```

Download and run `docker-setup` as a one-liner:

```bash
curl -sL https://github.com/nicholasdille/docker-setup/releases/latest/download/docker-setup | bash
```

Install to a well-known location:

```bash
curl --silent --location --output /usr/local/bin/docker-setup https://github.com/nicholasdille/docker-setup/releases/latest/download/docker-setup
chmod +x /usr/local/bin/docker-setup
```

`docker-setup` will warn you if some prerequisites are missing.

See [docs](docs) for the complete documentation.
