// Package registry is a chimera of moby/moby#36310 and buildkit testutil but simplified
package registry

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"
	"text/template"
	"time"

	"github.com/opencontainers/go-digest"
)

// Registry represents a registry version 2
type Registry struct {
	cmd     *exec.Cmd
	cleanup func()
	dir     string
	url     string // unset until start
}

type configOptions struct {
	RootDirectory string
}

func generateConfig(t *testing.T, c configOptions) []byte {
	config := `version: 0.1
loglevel: debug
storage:
    filesystem:
        rootdirectory: {{.RootDirectory}}
http:
    addr: 127.0.0.1:0
`
	tmpl := template.Must(template.New("config").Parse(config))
	var b bytes.Buffer
	if err := tmpl.Execute(&b, c); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func lookupBinary(t *testing.T) string {
	candidates := []string{"registry", "registry-v2", "docker-registry"}
	for _, c := range candidates {
		bin, err := exec.LookPath(c)
		if err != nil {
			return bin
		}
	}
	t.Skip("registry is not installed")
	return ""
}

// New creates a v2 registry server
func New(t *testing.T) *Registry {
	bin := lookupBinary(t)
	tmp, err := ioutil.TempDir("", "registry-test-")
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(tmp, "root")
	config := filepath.Join(tmp, "config.yml")
	configBytes := generateConfig(t, configOptions{RootDirectory: root})
	if err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(config, configBytes, 0600); err != nil {
		t.Fatal(err)
	}
	cleanup := func() {
		os.RemoveAll(tmp)
	}
	cmd := exec.Command(bin, config)
	return &Registry{
		cmd:     cmd,
		cleanup: cleanup,
		dir:     root,
		// url is unset here
	}
}

func (r *Registry) Start(t *testing.T) {
	rc, err := r.cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := r.cmd.Start(); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	r.url = detectURL(ctx, t, rc)
}

func detectURL(ctx context.Context, t *testing.T, rc io.ReadCloser) string {
	// TODO: registry itself should support port number notification
	r := regexp.MustCompile("listening on 127\\.0\\.0\\.1:(\\d+)")
	s := bufio.NewScanner(rc)
	found := make(chan struct{})
	defer func() {
		close(found)
		go io.Copy(ioutil.Discard, rc)
	}()

	go func() {
		select {
		case <-ctx.Done():
			select {
			case <-found:
				return
			default:
				rc.Close()
			}
		case <-found:
		}
	}()

	for s.Scan() {
		res := r.FindSubmatch(s.Bytes())
		if len(res) > 1 {
			return "http://localhost:" + string(res[1])
		}
	}
	t.Fatal("no listening address found")
	return ""
}

// Close kills the registry server
func (r *Registry) Close() {
	r.cmd.Process.Kill()
	r.cmd.Process.Wait()
	r.cleanup()
}

func (r *Registry) getBlobFilename(blobDigest digest.Digest) string {
	// Split the digest into its algorithm and hex components.
	dgstAlg, dgstHex := blobDigest.Algorithm(), blobDigest.Hex()

	// The path to the target blob data looks something like:
	//   baseDir + "docker/registry/v2/blobs/sha256/a3/a3ed...46d4/data"
	return fmt.Sprintf("%s/docker/registry/v2/blobs/%s/%s/%s/data", r.dir, dgstAlg, dgstHex[:2], dgstHex)
}

// ReadBlobContents read the file corresponding to the specified digest
func (r *Registry) ReadBlobContents(t *testing.T, blobDigest digest.Digest) []byte {
	// Load the target manifest blob.
	manifestBlob, err := ioutil.ReadFile(r.getBlobFilename(blobDigest))
	if err != nil {
		t.Fatalf("unable to read blob: %s", err)
	}

	return manifestBlob
}

// WriteBlobContents write the file corresponding to the specified digest with the given content
func (r *Registry) WriteBlobContents(t *testing.T, blobDigest digest.Digest, data []byte) {
	if err := ioutil.WriteFile(r.getBlobFilename(blobDigest), data, os.FileMode(0644)); err != nil {
		t.Fatalf("unable to write malicious data blob: %s", err)
	}
}

// ManifestDigest does not verify args, as this is just for test utility.
// usage: ManifestDigest("library/hello-world", latest)
func (r *Registry) ManifestDigest(t *testing.T, repo, tag string) digest.Digest {
	path := filepath.Join(r.Path(), "repositories", repo, "_manifests", "tags", tag, "current", "link")
	b, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return digest.Digest(string(b))
}

// URL of the registry
func (r *Registry) URL() string {
	return r.url
}

// Path returns the path where the registry write data
func (r *Registry) Path() string {
	return filepath.Join(r.dir, "docker", "registry", "v2")
}
