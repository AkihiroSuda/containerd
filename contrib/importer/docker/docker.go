// Package docker provides the importer for legacy Docker save/load spec.
package docker

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"strings"

	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/importer"
	"github.com/containerd/containerd/reference"
	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
)

// Dockerv1Importer implements Docker Image Spec v1.1.
// An image MUST have `manifest.json`.
// `repositories` file in Docker Image Spec v1.0 is not supported (yet).
//
// Selector spec: Docker RepoTag string (e.g. example.com/foo/bar:tag)
// Unlike containerd ref string, Docker RepoTag string MUST not contain digest. So don't put digest in selector.
// Selector needs to contain locator part at the moment. (e.g. "docker.io/library/busybox:latest" is correct but "busybox:latest" not)
var Dockerv1Importer importer.Importer = &dockerImporter{}

type dockerImporter struct {
}

// manifest is an entry in manifest.json.
type manifest struct {
	Config   string
	RepoTags []string
	Layers   []string
	// Parent is unsupported
	Parent string
}

func resolveManifest(mfsts []manifest, selector string) (*manifest, error) {
	repoTag, err := resolveDockerSelector(selector)
	if err != nil {
		return nil, err
	}
	for _, m := range mfsts {
		for _, rt := range m.RepoTags {
			if !strings.Contains(rt, "/") {
				if "docker.io/library/"+rt == repoTag {
					return &m, nil
				}
			}
			if rt == repoTag {
				return &m, nil
			}
		}
	}
	return nil, errors.Errorf("not found: %q", selector)
}

// resolveDockerSelector resolves selector string into Docker RepoTag string (e.g. example.com/foo/bar:tag)
// Unlike containerd ref string, Docker RepoTag string MUST not contain digest. So don't put digest in selector.
// Selector needs to contain locator part at the moment. (e.g. "docker.io/library/busybox:latest" is correct but "busybox:latest" not)
func resolveDockerSelector(selector string) (string, error) {
	parsed, err := reference.Parse(selector)
	if err != nil {
		return "", errors.Wrapf(err, "unexpected selector (looks like containerd reference string): %q", selector)
	}
	tag, dgst := reference.SplitObject(parsed.Object)
	if tag == "" || dgst != "" {
		return "",
			errors.Errorf("unexpected selector (looks like containerd reference string with object %q): %q", parsed.Object, selector)
	}
	return selector, nil
}

func onUntarManifestJSON(r io.Reader, selector string) (*manifest, error) {
	// name: "manifest.json"
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	var mfsts []manifest
	if err := json.Unmarshal(b, &mfsts); err != nil {
		return nil, err
	}
	return resolveManifest(mfsts, selector)
}

func onUntarLayerTar(ctx context.Context, r io.Reader, store content.Store, name string, size int64) (*ocispec.Descriptor, error) {
	// name is like "deadbeeddeadbeef/layer.tar""
	split := strings.Split(name, "/")
	if len(split) != 2 || !strings.HasSuffix(name, "/layer.tar") {
		return nil, errors.Errorf("unexpected name: %q", name)
	}
	// split[0] is not expected digest here
	cw, err := store.Writer(ctx, "unknown-"+split[0], size, "")
	if err != nil {
		return nil, err
	}
	defer cw.Close()
	_, err = io.Copy(cw, r)
	if err != nil {
		return nil, err
	}
	if err = cw.Commit(ctx, size, ""); err != nil {
		return nil, err
	}
	desc := ocispec.Descriptor{
		MediaType: images.MediaTypeDockerSchema2Layer,
		Size:      size,
	}
	desc.Digest = cw.Digest()
	return &desc, nil
}

type dotJSON struct {
	desc ocispec.Descriptor
	img  ocispec.Image
}

func onUntarDotJSON(ctx context.Context, r io.Reader, store content.Store, name string, size int64) (*dotJSON, error) {
	config := dotJSON{}
	config.desc.MediaType = images.MediaTypeDockerSchema2Config
	config.desc.Size = size
	// name is like "deadbeeddeadbeef.json""
	if strings.Contains(name, "/") || !strings.HasSuffix(name, ".json") {
		return nil, errors.Errorf("unexpected name: %q", name)
	}
	cw, err := store.Writer(ctx, "unknown-"+name, size, "")
	if err != nil {
		return nil, err
	}
	defer cw.Close()
	var buf bytes.Buffer
	tr := io.TeeReader(r, &buf)
	_, err = io.Copy(cw, tr)
	if err != nil {
		return nil, err
	}
	if err = cw.Commit(ctx, size, ""); err != nil {
		return nil, err
	}
	config.desc.Digest = cw.Digest()
	if err := json.Unmarshal(buf.Bytes(), &config.img); err != nil {
		return nil, err
	}
	return &config, nil
}

func (oi *dockerImporter) Import(ctx context.Context, store content.Store, ref string, reader io.Reader, selector string) (images.Image, error) {
	imgrec := images.Image{Name: ref}
	tr := tar.NewReader(reader)
	var (
		mfst    *manifest
		layers  = make(map[string]ocispec.Descriptor, 0) // key: filename (deadbeeddeadbeef/layer.tar)
		configs = make(map[string]dotJSON, 0)            // key: filename (deadbeeddeadbeef.json)
	)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return imgrec, err
		}
		if hdr.Typeflag != tar.TypeReg && hdr.Typeflag != tar.TypeRegA {
			continue
		}
		if hdr.Name == "manifest.json" {
			mfst, err = onUntarManifestJSON(tr, selector)
			if err != nil {
				return imgrec, err
			}
			continue
		}
		slashes := len(strings.Split(hdr.Name, "/"))
		if slashes == 2 && strings.HasSuffix(hdr.Name, "/layer.tar") {
			desc, err := onUntarLayerTar(ctx, tr, store, hdr.Name, hdr.Size)
			if err != nil {
				return imgrec, err
			}
			layers[hdr.Name] = *desc
			continue
		}
		if slashes == 1 && strings.HasSuffix(hdr.Name, ".json") {
			c, err := onUntarDotJSON(ctx, tr, store, hdr.Name, hdr.Size)
			if err != nil {
				return imgrec, err
			}
			configs[hdr.Name] = *c
			continue
		}
	}
	if mfst == nil {
		return imgrec, errors.Errorf("no manifest found for selector %q", selector)
	}
	config, ok := configs[mfst.Config]
	if !ok {
		return imgrec, errors.Errorf("image config not %q found for selector %q", mfst.Config, selector)
	}
	manifest := ocispec.Manifest{
		Versioned: specs.Versioned{
			SchemaVersion: 2,
		},
		Config: config.desc,
	}
	for _, f := range mfst.Layers {
		desc, ok := layers[f]
		if !ok {
			return imgrec, errors.Errorf("layer %q not found", f)
		}
		manifest.Layers = append(manifest.Layers, desc)
	}

	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		return imgrec, err
	}
	manifestBytesR := bytes.NewReader(manifestBytes)
	manifestDigest := digest.FromBytes(manifestBytes)
	if err := content.WriteBlob(ctx, store, "unknown-"+manifestDigest.String(), manifestBytesR, int64(len(manifestBytes)), manifestDigest); err != nil {
		return imgrec, err
	}

	imgrec.Target = ocispec.Descriptor{
		MediaType: images.MediaTypeDockerSchema2Manifest,
		Digest:    manifestDigest,
		Size:      int64(len(manifestBytes)),
		// TODO(AkihiroSuda): set platform
	}
	return imgrec, nil
}
