package main

func defaultConfig() *config {
	return &config{
		Root:  "/var/lib/containerd",
		State: "/run/containerd",
		GRPC: grpcConfig{
			Address: "/run/containerd/containerd.sock",
		},
		Debug: debug{
			Level:   "info",
			Address: "/run/containerd/debug.sock",
		},
		Snapshotters: []snapshotterConfig{
			{
				Name:   "overlay",
				Plugin: "snapshot-overlay",
			},
			{
				Name:   "btrfs",
				Plugin: "snapshot-btrfs",
			},
		},
	}
}
