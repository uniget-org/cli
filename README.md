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

[![forthebadge](https://forthebadge.com/images/badges/uses-badges.svg)](https://forthebadge.com) [![forthebadge](https://forthebadge.com/images/badges/check-it-out.svg)](https://forthebadge.com) [![forthebadge](https://forthebadge.com/images/badges/for-you.svg)](https://forthebadge.com) [![forthebadge](https://forthebadge.com/images/badges/kinda-sfw.svg)](https://forthebadge.com)

[![forthebadge](https://forthebadge.com/images/badges/works-on-my-machine.svg)](https://forthebadge.com)

![](https://badgen.net/github/tags/nicholasdille/docker-setup?icon=github)

## Install

Install prerequisites on Ubuntu:

```bash
apt-get update
apt-get -y install --no-install-recommends \
    curl \
    ca-certificates \
    bash \
    git \
    iptables
```

Download and run `docker-setup.sh`:

```bash
curl -sLO https://github.com/nicholasdille/docker-setup/releases/latest/download/docker-setup.sh
bash docker-setup.sh
```

## cloud-init

If using `cloud-init` you can use the included [`cloud-init.yaml`](contrib/cloud-init.yaml).

How to use this on Hetzner Cloud:

```bash
hcloud server create \
   --name foo \
   --location fsn1 \
   --type cx21 \
   --image ubuntu-20.04 \
   --ssh-key 4662975 \
   --user-data-from-file contrib/cloud-init.yaml
```
