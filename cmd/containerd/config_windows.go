package main

import (
	"os"
	"path/filepath"
)

func defaultConfig() *config {
	return &config{
		Root:  filepath.Join(os.Getenv("programfiles"), "containerd", "root"),
		State: filepath.Join(os.Getenv("programfiles"), "containerd", "state"),
		GRPC: grpcConfig{
			Address: `\\.\pipe\containerd-containerd`,
		},
		Debug: debug{
			Level:   "info",
			Address: `\\.\pipe\containerd-debug`,
		},
		Snapshotters: []snapshotterConfig{
			{
				Name:   "windows",
				Plugin: "snapshot-windows",
			},
		},
	}
}
