# Rootless mode (WIP)

Terminal 1:
```
$ unshare -U -m
unshared$ echo $$
3539
```

Unsharing mountns (and userns) is required for mounting overlayfs.

Terminal 2:
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

Terminal 1:
```
unshared# mount -t tmpfs none /run/containerd
unshared# containerd -C ~/.config/containerd/config.toml
```

Mounting `/run/containerd` is needed for issue XXXXX.

Terminal 2:
```
$ nsenter -U -m -C  3539
unshared# ctr -a /run/users/1001/containerd/containerd.sock images pull docker.io/library/alpine:latest
unshared# ctr -a /run/users/1001/containerd/containerd.sock run -t --rm  --rootless --fifo-dir /run/users/1001/containerd/fifo docker.io/library/alpine:latest foo
container# echo "nameserver 8.8.8.8" > /etc/resolv.conf
container# apk update
container# apk add fortune
```
