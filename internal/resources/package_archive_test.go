package resources

import (
	"archive/zip"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// The API reads the archive's bytes: a tar.gz is rejected with a 500 whatever
// content type it is declared as, so the builder must emit a zip. This reads the
// result back with archive/zip rather than trusting the extension.
func TestCreateArchiveFromDirProducesAZip(t *testing.T) {
	srcDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(srcDir, "publisher.json"), []byte(`{"name":"p"}`), 0o600); err != nil {
		t.Fatalf("writing fixture: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(srcDir, "models"), 0o755); err != nil {
		t.Fatalf("creating fixture dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "models", "orders.malloy"), []byte("source: o is _"), 0o600); err != nil {
		t.Fatalf("writing fixture: %v", err)
	}

	archivePath, err := createArchiveFromDir(srcDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer os.Remove(archivePath)

	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		t.Fatalf("archive is not a readable zip: %v", err)
	}
	defer reader.Close()

	var names []string
	for _, f := range reader.File {
		names = append(names, f.Name)
	}
	sort.Strings(names)

	// Entries are relative to srcDir, so the package's files sit at the archive
	// root -- publisher.json has to be findable there.
	want := []string{"models/", "models/orders.malloy", "publisher.json"}
	if len(names) != len(want) {
		t.Fatalf("expected entries %v, got %v", want, names)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Errorf("expected entry %q, got %q", want[i], names[i])
		}
	}
}
