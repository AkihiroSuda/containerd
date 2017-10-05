package containerd

import (
	"runtime"
	"testing"
)

// TestOCIExportAndImport exports testImage as a tar stream,
// and import the tar stream as a new image.
func TestOCIExportAndImport(t *testing.T) {
	// TODO: support windows
	if testing.Short() || runtime.GOOS == "windows" {
		t.Skip()
	}
	ctx, cancel := testContext()
	defer cancel()

	client, err := New(address)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	pulled, err := client.Pull(ctx, testImage)
	if err != nil {
		t.Fatal(err)
	}

	exported, err := client.Export(ctx, pulled.Target())
	if err != nil {
		t.Fatal(err)
	}

	importRef := "test/export-and-import:tmp"
	// OCI import selector can be either reference string or reference object string.
	// For OCI, since we don't need have concept of Locator, we use object string here.
	_, err = client.Import(ctx, importRef, exported, WithImportSelector("@"+pulled.Target().Digest.String()))
	if err != nil {
		t.Fatal(err)
	}

	err = client.ImageService().Delete(ctx, importRef)
	if err != nil {
		t.Fatal(err)
	}
}
