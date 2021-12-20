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

## Tools

The following tools are included in `docker-setup`. The exact versions are pinned inside `docker-setup`.

```plaintext
arkade buildah buildkit buildx clusterawsadm clusterctl cni cni-isolation conmon containerd cosign crictl crun dive docker docker-compose docker-machine docker-scan fuse-overlayfs fuse-overlayfs-shapshotter helm hub-tool img imgcrypt jq k3d k3s kapp kind kompose krew kubectl kubeswitch kustomize manifest-tool minikube nerdctl oras portainer porter podman regclient rootlesskit runc skopeo slirp4netns stargz-snapshotter trivy yq ytt
```

## Usage

All tools will be installed in parallel. Many tools only require a simple download so that most tools will be installed really quickly.

Progress is displayed in a fullscreen status page where each line shows the status of one tool:

[![asciicast](https://asciinema.org/a/457481.svg)](https://asciinema.org/a/457481?autoplay=1)

If the terminal window has fewer lines than the number of tools, the output automatically switches to simple output. The progress is only shown by the spinner:

[![asciicast](https://asciinema.org/a/457484.svg)](https://asciinema.org/a/457484)

Download and run `docker-setup` as a one-liner:

```bash
curl -sL https://github.com/nicholasdille/docker-setup/releases/latest/download/docker-setup.sh | bash
```

You can tweak the behaviour of `docker-setup` by passing parameters or environment variables:

| Parameter         | Variable                 | Meaning |
| ----------------- | ------------------------ | ------- |
| `--help`          | n/a                      | Display help for parameters and environment variables |
| `--version`       | n/a                      | Display version and exit |
| `--check-only`    | `CHECK_ONLY`             | Only check if tools need to be installed or updated |
| `--no-wait`       | `NO_WAIT`                | Do not wait before installing |
| `--reinstall`     | `REINSTALL`              | Install all tools again |
| `--simple-output` | `SIMPLE_OUTPUT`          | Do not display  |
| `--no-spinner`    | `NO_SPINNER`             | Do not display spinner in simple output |
|                   | `TARGET`                 | Specifies the target directory for binaries. Defaults to /usr |
|                   | `CGROUP_VERSION`         | Specifies which version of cgroup to use. Defaults to v2 |
|                   | `DOCKER_ADDRESS_BASE`    | Specifies the address pool for networks, e.g. 192.168.0.0/16 |
|                   | `DOCKER_ADDRESS_SIZE`    | Specifies the size of each network, e.g. 24 |
|                   | `DOCKER_REGISTRY_MIRROR` | Specifies a host to be used as registry mirror, e.g. https://proxy.my-domain.tld |
|                   | `DOCKER_COMPOSE`         | Specifies which major version of docker-compose to use. Defaults to v2 |

When one or more tool names are supplied as parameters, they will be re-installed:

```bash
bash docker-setup.sh docker trivy
```

[![asciicast](https://asciinema.org/a/457486.svg)](https://asciinema.org/a/457486)

## Internals

`docker-setup` contains a list of all tools with pinned versions. These versions are automatically updated using [RenovateBot](https://www.whitesourcesoftware.com/free-developer-tools/renovate/).

Installation logs are placed in `/var/log/docker-setup`.

The installation progress is cached in `/var/cache/docker-setup/progress`. The context will be removed after the installation completed.

Depending on the tool additional files are placed outside of `${TARGET}`:

- Systemd units in `/etc/systemd/system/`
- Init scripts in `/etc/init.d/` with defaults in `/etc/default/`

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
