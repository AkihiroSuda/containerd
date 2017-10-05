// Package docker1 provides the importer for legacy Docker save/load spec.
package docker1

// manifest is an entry in manifest.json.
type manifest struct {
	Config   string
	RepoTags []string
	Layers   []string
	// Parent is unsupported
	Parent string
}
