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

## Install

Download and run `docker-setup`:

```bash
curl -sLO https://github.com/nicholasdille/docker-setup/releases/latest/download/docker-setup.sh
bash docker-setup.sh
```

See [below](#usage) for more options.

`docker-setup` will warn you if some prerequisites are missing.

Releases are tested on the following distributions:
- Alpine 3.15
- CentOS 7 (see note below)
- Debian 11
- Fedora 35
- Ubuntu 20.04
- Ubuntu 21.04

`docker-setup` implements a workaround for CentOS 7 because it does not offer `iptables-legacy`. Therefore, `docker-setup` installs a binary package for `iptables-legacy` from [nicholasdille/centos-iptables-legacy](https://github.com/nicholasdille/centos-iptables-legacy). As long as Docker does not support `nftables`, the daemon requires `iptables-legacy` or can only run with [`--iptables=false` which breaks container networking](https://docs.docker.com/network/iptables/#prevent-docker-from-manipulating-iptables). The test for CentOS 7 is currently hanging (tracked in [#262](https://github.com/nicholasdille/docker-setup/issues/262)) and therefore disabled. CentOS 8 fails to update repository metadata for `appstream` (tracked in [#263](https://github.com/nicholasdille/docker-setup/issues/263)) and is therefore disabled.

## Tools

The following tools are included in `docker-setup`. The exact versions are pinned inside `docker-setup`.

```plaintext
arkade buildah buildkit buildx clusterawsadm clusterctl cni cni-isolation conmon containerd cosign crane crictl crun ctop dasel dive docker docker-compose docker-machine docker-scan docuum dry duffle firecracker firectl footloose fuse-overlayfs fuse-overlayfs-snapshotter glow gvisor helm helmfile hub-tool ignite img imgcrypt ipfs jp jq jwt k3d k3s k9s kapp kind kompose krew kubectl kubectl-build kubectl-free kubectl-resources kubefire kubeletctl kubeswitch kustomize lazydocker lazygit manifest-tool minikube nerdctl oras patat portainer porter podman qemu regclient rootlesskit runc skopeo slirp4netns sops stargz-snapshotter trivy umoci yq ytt
```

## Usage

All tools will be installed in parallel. Many tools only require a simple download so that most tools will be installed really quickly.

`docker-setup` displays a progress bar unless suppressed by the command line switch:

[![Using docker-setup](https://asciinema.org/a/6rptGICcjvJZR4F5OjMmRqG7L.svg)](https://asciinema.org/a/6rptGICcjvJZR4F5OjMmRqG7L)

Download and run `docker-setup` as a one-liner:

```bash
curl -sL https://github.com/nicholasdille/docker-setup/releases/latest/download/docker-setup.sh | bash
```

You can tweak the behaviour of `docker-setup` by passing parameters or environment variables:

| Parameter          | Variable                 | Meaning |
| ------------------ | ------------------------ | ------- |
| `--help`           | n/a                      | Display help for parameters and environment variables |
| `--version`        | n/a                      | Display version and exit |
| `--check`          | `CHECK`                  | Only check if tools need to be installed or updated |
| `--no-wait`        | `NO_WAIT`                | Do not wait before installing |
| `--reinstall`      | `REINSTALL`              | Install all tools again |
| `--only`           | `ONLY`                   | Only install specified tools |
| `--no-progressbar` | `NO_PROGRESSBAR`         | Do not display progress bar |
| `--no-color`       | `NO_COLOR`               | Do not display colored output |
| `--plan`           | `PLAN`                   | Show planned installations |
| `--skip-docs`      | `SKIP_DOCS`              | Do not install documentation for faster installation |
|                    | `PREFIX`                 | Install into a subdirectory (see notes below) |
|                    | `TARGET`                 | Specifies the target directory for binaries. Defaults to /usr |
|                    | `CGROUP_VERSION`         | Specifies which version of cgroup to use. Defaults to v2 |
|                    | `DOCKER_ADDRESS_BASE`    | Specifies the address pool for networks, e.g. 192.168.0.0/16 |
|                    | `DOCKER_ADDRESS_SIZE`    | Specifies the size of each network, e.g. 24 |
|                    | `DOCKER_REGISTRY_MIRROR` | Specifies a host to be used as registry mirror, e.g. https://proxy.my-domain.tld |
|                    | `DOCKER_COMPOSE`         | Specifies which major version of docker-compose to use. Defaults to v2 |

## Internals

`docker-setup` contains a list of all tools with pinned versions. These versions are automatically updated using [RenovateBot](https://www.whitesourcesoftware.com/free-developer-tools/renovate/).

Installation logs are placed in `/var/log/docker-setup`.

The installation progress is cached in `/var/cache/docker-setup/progress`. The context will be removed after the installation completed.

Depending on the tool additional files are placed outside of `${TARGET}`:

- Systemd units in `/etc/systemd/system/`
- Init scripts in `/etc/init.d/` with defaults in `/etc/default/`

When `PREFIX` is specified, these files will also be placed in a subdirectory. But `docker-setup` will not handle daemon starts and restarts because it is assumed that the installation directory contains a different root file system.

## Air-gapped installation

`docker-setup` downloads several file during the installation. Some of them are coming from this repository. These files can now be placed in `/var/lib/docker-setup/contrib` to reduce the dependency on an internet connection. A tarball is published in the release (`contrib.tar.gz`) and included in the container image.

Air-gapped installations are not possible because not all files are included in the contrib tarball.

## Container Image

The [`docker-setup` container image](https://hub.docker.com/r/nicholasdille/docker-setup) helps installing all tools without otherweise touching the system:

```bash
docker container run --interactive --tty --rm \
    --mount type=bind,src=/,dst=/opt/docker-setup \
    --mount type=bind,src=/var/run/docker.sock,dst=/var/run/docker.sock \
    --env PREFIX=/opt/docker-setup \
    nicholasdille/docker-setup
```

## Docker

The Docker daemon will use the executables installed to `${TARGET}/libexec/docker/bin/` which are installed from the [official binary package](https://download.docker.com/linux/static/stable/x86_64/). The systemd unit as well as the init script have been modified to ensure this.

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
