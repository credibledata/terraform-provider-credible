package client

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestGetPackage(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/v0/organizations/my-org/environments/my-proj/packages/my-pkg" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(Package{
				Name:          "my-pkg",
				Description:   "A test package",
				LatestVersion: "1.0.0",
				CreatedAt:     "2025-01-01T00:00:00Z",
				UpdatedAt:     "2025-01-01T00:00:00Z",
			})
		}))
		defer server.Close()

		c := NewClient(server.URL, "ApiKey k", "org")
		result, err := c.GetPackage("my-org", "my-proj", "my-pkg")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.LatestVersion != "1.0.0" {
			t.Errorf("expected latestVersion %q, got %q", "1.0.0", result.LatestVersion)
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(APIError{Message: "not found"})
		}))
		defer server.Close()

		c := NewClient(server.URL, "ApiKey k", "org")
		_, err := c.GetPackage("my-org", "my-proj", "missing")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !IsNotFound(err) {
			t.Error("expected IsNotFound to return true")
		}
	})
}

func TestUpdatePackage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Package{
			Name:        "my-pkg",
			Description: "Updated description",
			CreatedAt:   "2025-01-01T00:00:00Z",
			UpdatedAt:   "2025-01-02T00:00:00Z",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "ApiKey k", "org")
	result, err := c.UpdatePackage("my-org", "my-proj", "my-pkg", &Package{Description: "Updated description"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Description != "Updated description" {
		t.Errorf("expected description %q, got %q", "Updated description", result.Description)
	}
}

func TestDeletePackage(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "DELETE" {
				t.Errorf("expected DELETE, got %s", r.Method)
			}
			if r.URL.Path != "/api/v0/organizations/my-org/environments/my-proj/packages/my-pkg" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		c := NewClient(server.URL, "ApiKey k", "org")
		err := c.DeletePackage("my-org", "my-proj", "my-pkg")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("server error"))
		}))
		defer server.Close()

		c := NewClient(server.URL, "ApiKey k", "org")
		err := c.DeletePackage("my-org", "my-proj", "my-pkg")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestListPackages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v0/organizations/my-org/environments/my-proj/packages" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]Package{
			{Name: "pkg-1"},
			{Name: "pkg-2"},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "ApiKey k", "org")
	result, err := c.ListPackages("my-org", "my-proj")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 packages, got %d", len(result))
	}
}

func TestGetVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v0/organizations/my-org/environments/my-proj/packages/my-pkg/versions/1.0.0" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Version{
			ID:            "1.0.0",
			ArchiveStatus: "unarchive",
			BuildStatus:   "READY",
			CreatedAt:     "2025-01-01T00:00:00Z",
			UpdatedAt:     "2025-01-01T00:00:00Z",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "ApiKey k", "org")
	result, err := c.GetVersion("my-org", "my-proj", "my-pkg", "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "1.0.0" {
		t.Errorf("expected ID %q, got %q", "1.0.0", result.ID)
	}
	if result.ArchiveStatus != "unarchive" {
		t.Errorf("expected archiveStatus %q, got %q", "unarchive", result.ArchiveStatus)
	}
}

func TestUpdateVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/organizations/my-org/environments/my-proj/packages/my-pkg/versions/1.0.0" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Version{
			ID:            "1.0.0",
			ArchiveStatus: "archive",
			BuildStatus:   "READY",
			CreatedAt:     "2025-01-01T00:00:00Z",
			UpdatedAt:     "2025-01-02T00:00:00Z",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "ApiKey k", "org")
	result, err := c.UpdateVersion("my-org", "my-proj", "my-pkg", "1.0.0", &Version{ArchiveStatus: "archive"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ArchiveStatus != "archive" {
		t.Errorf("expected archiveStatus %q, got %q", "archive", result.ArchiveStatus)
	}
}

// Pins the publish contract that was verified against a live control plane: the
// package's own path (not .../versions, which the API serves GET only), the four
// part names the API requires, md5Hash computed from the bytes actually sent, and
// a Package -- not a Version -- decoded from the response.
func TestPublishPackageVersion(t *testing.T) {
	archive := []byte("fake archive data")
	wantHash := fmt.Sprintf("%x", md5.Sum(archive))

	var (
		gotMethod string
		gotPath   string
		gotParts  map[string]string
		gotFile   []byte
		gotTypes  map[string]string
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotParts = map[string]string{}
		gotTypes = map[string]string{}

		reader, err := r.MultipartReader()
		if err != nil {
			t.Errorf("expected a multipart body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Errorf("reading multipart body: %v", err)
				break
			}
			body, err := io.ReadAll(part)
			if err != nil {
				t.Errorf("reading part %q: %v", part.FormName(), err)
			}
			if part.FormName() == "packageFile" {
				gotFile = body
			} else {
				gotParts[part.FormName()] = string(body)
			}
			gotTypes[part.FormName()] = part.Header.Get("Content-Type")
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Package{
			Name:          "my-pkg",
			Description:   "a package",
			LatestVersion: "1.0.0",
		})
	}))
	defer server.Close()

	tmpFile, err := os.CreateTemp("", "test-pkg-*.zip")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	if _, err := tmpFile.Write(archive); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	c := NewClient(server.URL, "ApiKey k", "org")
	result, err := c.PublishPackageVersion("my-org", "my-proj", "my-pkg", "a package",
		&Version{ID: "1.0.0"}, tmpFile.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != "POST" {
		t.Errorf("expected POST, got %s", gotMethod)
	}
	// The package's own path. Posting to .../versions is what returned 405.
	if want := "/api/v0/organizations/my-org/environments/my-proj/packages/my-pkg"; gotPath != want {
		t.Errorf("expected path %q, got %q", want, gotPath)
	}

	for _, name := range []string{"package", "version", "md5Hash"} {
		if _, ok := gotParts[name]; !ok {
			t.Errorf("missing required part %q; got parts %v", name, gotParts)
		}
	}
	if gotFile == nil {
		t.Error("missing required part \"packageFile\"")
	}

	// The API parses these as JSON and rejects text/plain.
	for _, name := range []string{"package", "version"} {
		if got := gotTypes[name]; got != "application/json" {
			t.Errorf("expected part %q to be application/json, got %q", name, got)
		}
	}

	// The package part carries the name and description -- omitting it fails with
	// "Create package request must have a Package JSON".
	var sentPackage Package
	if err := json.Unmarshal([]byte(gotParts["package"]), &sentPackage); err != nil {
		t.Fatalf("package part is not valid JSON: %v", err)
	}
	if sentPackage.Name != "my-pkg" {
		t.Errorf("expected package name %q, got %q", "my-pkg", sentPackage.Name)
	}
	if sentPackage.Description != "a package" {
		t.Errorf("expected package description %q, got %q", "a package", sentPackage.Description)
	}

	var sentVersion Version
	if err := json.Unmarshal([]byte(gotParts["version"]), &sentVersion); err != nil {
		t.Fatalf("version part is not valid JSON: %v", err)
	}
	if sentVersion.ID != "1.0.0" {
		t.Errorf("expected version id %q, got %q", "1.0.0", sentVersion.ID)
	}

	// The hash must describe the uploaded bytes; the API verifies it.
	if gotParts["md5Hash"] != wantHash {
		t.Errorf("expected md5Hash %q, got %q", wantHash, gotParts["md5Hash"])
	}
	if !bytes.Equal(gotFile, archive) {
		t.Errorf("uploaded bytes differ from the source archive")
	}

	// The response is the package, and its latestVersion is what was published.
	if result.LatestVersion != "1.0.0" {
		t.Errorf("expected latestVersion %q, got %q", "1.0.0", result.LatestVersion)
	}
	if result.Name != "my-pkg" {
		t.Errorf("expected name %q, got %q", "my-pkg", result.Name)
	}
}

// A publish failure must surface the API's message rather than being reported as
// a success with empty state.
func TestPublishPackageVersionSurfacesAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusFailedDependency)
		json.NewEncoder(w).Encode(APIError{Message: "Package failed to compile."})
	}))
	defer server.Close()

	tmpFile, err := os.CreateTemp("", "test-pkg-*.zip")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Write([]byte("fake archive data"))
	tmpFile.Close()

	c := NewClient(server.URL, "ApiKey k", "org")
	_, err = c.PublishPackageVersion("my-org", "my-proj", "my-pkg", "", &Version{ID: "1.0.0"}, tmpFile.Name())
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !strings.Contains(err.Error(), "Package failed to compile.") {
		t.Errorf("expected the API message in the error, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "424") {
		t.Errorf("expected the status code in the error, got %q", err.Error())
	}
}
