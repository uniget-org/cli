# docker-setup

XXX

## cgroup v2

XXX

```bash
sed -i 's/GRUB_CMDLINE_LINUX=""/GRUB_CMDLINE_LINUX="systemd.unified_cgroup_hierarchy=1"/' /etc/default/grub
update-grub
reboot
```

## iptables-legacy

XXX

```
apt-get update
apt-get -y install iptables
```

## uidmap

XXX

```bash
apt-get update
apt-get -y install uidmap
```
