# Rootless mode (Experimental)

Requirements:
- runc (May 30, 2018) or later
- Some distros such as Debian and Arch Linux require `echo 1 > /proc/sys/kernel/unprivileged_userns_clone`
- `newuidmap` and `newgidmap` need to be installed on the host. These commands are provided by the `uidmap` package.
- `/etc/subuid` and `/etc/subgid` should contain >= 65536 sub-IDs. e.g. `penguin:231072:65536`.
- To run in a Docker container with non-root `USER`, `docker run --privileged` is still required. See also Jessie's blog: https://blog.jessfraz.com/post/building-container-images-securely-on-kubernetes/

Remarks:

* The data dir will be set to `/home/$USER/.local/share/containerd` by default.
* The address will be set to `/run/user/$UID/containerd/containerd.sock` by default.
* CRI plugin is not supported yet.
* `overlayfs` snapshotter is not supported except on [Ubuntu-flavored kernel](http://kernel.ubuntu.com/git/ubuntu/ubuntu-artful.git/commit/fs/overlayfs?h=Ubuntu-4.13.0-25.29&id=0a414bdc3d01f3b61ed86cfe3ce8b63a9240eba7). `native` snapshotter should work on non-Ubuntu kernel.
* Cgroups and network namespace are disabled at the moment. The future plan is to enable them by running a couple of SUID-enabled binaries before running `containerd`.

## Usage

### Terminal 1:

```
$ unshare -U -m
unshared$ echo $$ > /tmp/pid
```

Unsharing mountns (and userns) is required for mounting filesystems without real root privileges.

### Terminal 2:

```
$ id -u
1001
$ grep $(whoami) /etc/subuid
penguin:231072:65536
$ grep $(whoami) /etc/subgid
penguin:231072:65536
$ newuidmap $(cat /tmp/pid) 0 1001 1 1 231072 65536
$ newgidmap $(cat /tmp/pid) 0 1001 1 1 231072 65536
```

### Terminal 1:

```
unshared# containerd
```

### Terminal 2:

```
$ nsenter -U -m -t $(cat /tmp/pid)
unshared# ctr -a /run/user/1001/containerd/containerd.sock images pull docker.io/library/debian:latest
unshared# ctr -a /run/user/1001/containerd/containerd.sock run -t --rm --rootless --net-host docker.io/library/debian:latest foo
foo#
```
