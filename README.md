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

## Quickstart

Download and run `docker-setup`:

```bash
curl -sLO https://github.com/nicholasdille/docker-setup/releases/latest/download/docker-setup.sh
bash docker-setup.sh
```

Download and run `docker-setup` as a one-liner:

```bash
curl -sL https://github.com/nicholasdille/docker-setup/releases/latest/download/docker-setup.sh | bash
```

See [below](#usage) for more options.

`docker-setup` will warn you if some prerequisites are missing.

## Purpose

`docker-setup` is meant to bootstrap a new box with Docker as well as install useful tools from the container ecosystem.It aims to be distribution-agnostic and provide reasonable defaults. Personally, I am using it to prepare virtual machines for my own experiments as well as training environments.

`docker-setup` is not meant to be a competitor to Docker Desktop. It is lacking important features required for developing using Docker.

## Included tools

The following tools are included in `docker-setup`. The exact versions are pinned inside `docker-setup`.

```plaintext
arkade buildah buildkit buildx clusterawsadm clusterctl cni cni-isolation conmon containerd containerssh cosign crane crictl crun ctop dasel dive docker docker-compose docker-machine docker-scan docuum dry duffle dyff firecracker firectl footloose fuse-overlayfs fuse-overlayfs-snapshotter glow gvisor hcloud helm helmfile hub-tool ignite img imgcrypt ipfs jp jq jwt k3d k3s k9s kapp kind kompose krew kubectl kubectl-build kubectl-free kubectl-resources kubefire kubeletctl kubeswitch kustomize lazydocker lazygit manifest-tool minikube nerdctl oras patat portainer porter podman qemu regclient rootlesskit runc skopeo slirp4netns sops sshocker stargz-snapshotter trivy umoci yq ytt
```

## Supported distributons

Releases are tested on the following distributions:
- Alpine 3.15
- CentOS 7 (see note below)
- Debian 11
- Fedora 35
- Ubuntu 20.04
- Ubuntu 21.04

`docker-setup` implements a workaround for CentOS 7 because it does not offer `iptables-legacy`. Therefore, `docker-setup` installs a binary package for `iptables-legacy` from [nicholasdille/centos-iptables-legacy](https://github.com/nicholasdille/centos-iptables-legacy). As long as Docker does not support `nftables`, the daemon requires `iptables-legacy` or can only run with [`--iptables=false` which breaks container networking](https://docs.docker.com/network/iptables/#prevent-docker-from-manipulating-iptables). The test for CentOS 7 is currently hanging (tracked in [#262](https://github.com/nicholasdille/docker-setup/issues/262)) and therefore disabled. CentOS 8 fails to update repository metadata for `appstream` (tracked in [#263](https://github.com/nicholasdille/docker-setup/issues/263)) and is therefore disabled.

## Usage

All tools will be installed in parallel. Many tools only require a simple download so that most tools will be installed really quickly.

You can tweak the behaviour of `docker-setup` by passing parameters or environment variables:

| Parameter           | Variable                 | Meaning |
| ------------------- | ------------------------ | ------- |
| `--help`            | n/a                      | Display help for parameters and environment variables |
| `--version`         | n/a                      | Display version and exit |
| `--bash-completion` | n/a                      | Output completion script for bash |
| `--check`           | `CHECK`                  | Only check if tools need to be installed or updated |
| `--no-wait`         | `NO_WAIT`                | Do not wait before installing |
| `--reinstall`       | `REINSTALL`              | Install all tools again |
| `--only`            | `ONLY`                   | Only install specified tools |
| `--only-installed`  | `ONLY_INSTALLED`         | Only process installed tools |
| `--no-progressbar`  | `NO_PROGRESSBAR`         | Do not display progress bar |
| `--no-color`        | `NO_COLOR`               | Do not display colored output |
| `--plan`            | `PLAN`                   | Show planned installations |
| `--skip-docs`       | `SKIP_DOCS`              | Do not install documentation for faster installation |
|                     | `PREFIX`                 | Install into a subdirectory (see notes below) |
|                     | `TARGET`                 | Specifies the target directory for binaries. Defaults to /usr |
|                     | `CGROUP_VERSION`         | Specifies which version of cgroup to use. Defaults to v2 |
|                     | `DOCKER_ADDRESS_BASE`    | Specifies the address pool for networks, e.g. 192.168.0.0/16 |
|                     | `DOCKER_ADDRESS_SIZE`    | Specifies the size of each network, e.g. 24 |
|                     | `DOCKER_REGISTRY_MIRROR` | Specifies a host to be used as registry mirror, e.g. https://proxy.my-domain.tld |
|                     | `DOCKER_COMPOSE`         | Specifies which major version of docker-compose to use. Defaults to v2 |

Before installing any tools, `docker-setup` displays a list of all supported tools to visualize the current status and what will happen. All tools show the following indicators:

- Suffix with either a green check mark or a red cross mark to indicate whether it is up-to-date or outdated
- Colored in green to indicate that the tool is already installed and will not be re-installed
- Colored in yellow to indicate that the tool will be installed or updated
- Colored in grey to indicate that the tool will be skipped because you specified `--only`/`ONLY`

## Scenario 1: You want all tools

Install tools for the first time:

```bash
bash docker-setup.sh
```

The same command updates outdated tools.

Check if tools are outdated. `docker-setup` will return with exit code 1 if one or more tools are outdated:

```bash
bash docker-setup.sh --check
```

## Scenario 2: You want some tools

Install or update selected tools, e.g. `docker`:

```bash
bash docker-setup.sh --only docker
```

Check if tools are outdated:

```bash
bash docker-setup.sh --only docker --check
```

## Scenario 3: Reinstall all or some tools

By adding the `--reinstall` parameter, all tools can be reinstalled regardless if they are outdated:

```bash
bash docker-setup.sh --reinstall
```

The same applies when combining `--reinstall` with `--only`:

```bash
bash docker-setup.sh --only docker --reinstall
```

## Scenario 4: You only want to process installed tools

If you have previously installed tools using `docker-setup`, you can choose to update only installed tools:

```bash
bash docker-setup.sh --only-installed
```

You cannot combine this with `--only`/`ONLY`.

## Scenario 5: You don't want to wait

If you are used to running `docker-setup` with `--check` before installing or updating, you can also skip the delay by adding `--no-wait`:

```bash
bash docker-setup.sh --no-wait
```

This can also be used when installing or updating some tools:

```bash
bash docker-setup.sh --only docker --no-wait
```

## Scenario 6: Plan installation of all or some tools

Specifying `--check` will display outdated tools and return with exit code 1 if any tools are outdated. `--plan` will do neither and stop execution before any installation takes place:

```bash
bash docker-setup.sh --only docker --plan
```

## Scenario 7: Provide parameters using the onliner

All parameters are mapped to environment variables internally. Therefore you can supply environment variables instead of parameters. For a reference, see [usage](#usage) above.

```bash
curl -sL https://github.com/nicholasdille/docker-setup/releases/latest/download/docker-setup.sh | NO_WAIT=true bash
```

If you prefer parameters, `bash` requires the parameter `-s` before any parameters for `docker-setup` can be supplied:

```bash
curl -sL https://github.com/nicholasdille/docker-setup/releases/latest/download/docker-setup.sh | bash -s --no-wait
```

## Scenario 8: Container Image

The [`docker-setup` container image](https://hub.docker.com/r/nicholasdille/docker-setup) helps installing all tools without otherweise touching the system:

```bash
docker container run --interactive --tty --rm \
    --mount type=bind,src=/,dst=/opt/docker-setup \
    --mount type=bind,src=/var/run/docker.sock,dst=/var/run/docker.sock \
    --env PREFIX=/opt/docker-setup \
    nicholasdille/docker-setup
```

The Docker socket is necessary to install some tools or complete the installation of some tools.

## Internals

`docker-setup` contains a list of all tools with pinned versions. These versions are automatically updated using [RenovateBot](https://www.whitesourcesoftware.com/free-developer-tools/renovate/).

Installation logs are placed in `/var/log/docker-setup`.

The installation progress is cached in `/var/cache/docker-setup/progress`. The context will be removed after the installation completed.

Depending on the tool additional files are placed outside of `${TARGET}`:

- Systemd units in `/etc/systemd/system/`
- Init scripts in `/etc/init.d/` with defaults in `/etc/default/`

When `PREFIX` is specified, these files will also be placed in a subdirectory. But `docker-setup` will not handle daemon starts and restarts because it is assumed that the installation directory contains a different root file system.

## Docker

The Docker daemon will use the executables installed to `${TARGET}/libexec/docker/bin/` which are installed from the [official binary package](https://download.docker.com/linux/static/stable/x86_64/). The systemd unit as well as the init script have been modified to ensure this.

## Air-gapped installation

`docker-setup` downloads several file during the installation. Some of them are coming from this repository. These files can now be placed in `/var/lib/docker-setup/contrib` to reduce the dependency on an internet connection. A tarball is published in the release (`contrib.tar.gz`) and included in the container image.

Air-gapped installations are not possible because not all files are included in the contrib tarball.

## cloud-init

When used together with `cloud-init` you can apply the included [`cloud-init.yaml`](contrib/cloud-init.yaml). It automatically prepares your VM for `docker-setup`:

1. Configures `apt` to skip recommended as well as suggested packages
1. Install prerequisites
1. Enable cgroup v2
1. Reboot

The following example if for [Hetzner Cloud](https://www.hetzner.com/cloud):

```bash
hcloud server create \
   --name foo \
   --location fsn1 \
   --type cx21 \
   --image ubuntu-20.04 \
   --ssh-key 12345678 \
   --user-data-from-file contrib/cloud-init.yaml
```

## Windows

`docker-setup.ps1` is still very much work in progress and is mostly untested.
