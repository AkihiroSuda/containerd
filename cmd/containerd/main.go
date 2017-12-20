package main

import (
	"fmt"
	"os"

	"github.com/containerd/containerd/cmd/containerd/daemon"
)

func main() {
	cli := daemon.CLI()
	if err := cli.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "containerd: %s\n", err)
		os.Exit(1)
	}
}
