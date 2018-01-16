# Rootless mode (WIP)


## Terminal 1:

```
$ mkdir -p ~/.config/containerd
$ vi ~/.config/containerd/config.toml
(you can use `dotconfig_containetd_config.toml` as a template)
```

```
$ unshare -U -m
unshared$ echo $$
3539
```

Unsharing mountns (and userns) is required for mounting overlayfs.

## Terminal 2:

```
$ id -u
1001
$ grep $(whoami) /etc/subuid
suda:231072:65536
$ grep $(whoami) /etc/subgid
suda:231072:65536
$ newuidmap 3539 0 1001 1 1 231072 65536
$ newgidmap 3539 0 1001 1 1 231072 65536
```

## Terminal 1:

```
unshared# echo "root:1:65536" > ~/.config/containerd/subuidgid
unshared# mount --bind --ro ~/.config/containerd/subuidgid /etc/subuid
unshared# mount --bind --ro ~/.config/containerd/subuidgid /etc/subgid
unshared# mount -t tmpfs none /run/containerd
```

- Mounting `/etc/sub{u,g}id` is needed for enabling `newuidmap(1)` and `newgidmap(1)` to be called from runc to enable `setgroups(2)` within containers. (needed for e.g. apt)
- Mounting `/run/containerd` is needed for issue XXXXX. (`runtimes.linux.runtime_root` is not hornored)


```
unshared# containerd -c ~/.config/containerd/config.toml
```

## Terminal 2:

```
$ nsenter -U -m -t 3539
unshared# ctr -a /run/user/1001/containerd/containerd.sock images pull docker.io/library/alpine:latest
unshared# ctr -a /run/user/1001/containerd/containerd.sock run -t --rm  --rootless --fifo-dir /run/user/1001/containerd/fifo docker.io/library/alpine:latest foo
container# echo "nameserver 8.8.8.8" > /etc/resolv.conf
container# apk update
container# apk add fortune
```

Make sure you can also execute apt and yum,

- The container works in host netns, but `--host-net` fails for some bind-mount error.
