package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestCreatePackage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/organizations/my-org/projects/my-proj/packages" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var body Package
		json.NewDecoder(r.Body).Decode(&body)
		if body.Name != "my-pkg" {
			t.Errorf("expected name %q, got %q", "my-pkg", body.Name)
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Package{
			Name:        "my-pkg",
			Description: "A test package",
			CreatedAt:   "2025-01-01T00:00:00Z",
			UpdatedAt:   "2025-01-01T00:00:00Z",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "ApiKey k", "org")
	result, err := c.CreatePackage("my-org", "my-proj", &Package{Name: "my-pkg", Description: "A test package"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "my-pkg" {
		t.Errorf("expected name %q, got %q", "my-pkg", result.Name)
	}
}

func TestGetPackage(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/v0/organizations/my-org/projects/my-proj/packages/my-pkg" {
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
			if r.URL.Path != "/api/v0/organizations/my-org/projects/my-proj/packages/my-pkg" {
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
		if r.URL.Path != "/api/v0/organizations/my-org/projects/my-proj/packages" {
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
		if r.URL.Path != "/api/v0/organizations/my-org/projects/my-proj/packages/my-pkg/versions/1.0.0" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Version{
			ID:            "1.0.0",
			ArchiveStatus: "unarchive",
			IndexStatus:   "complete",
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
		if r.URL.Path != "/api/v0/organizations/my-org/projects/my-proj/packages/my-pkg/versions/1.0.0" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Version{
			ID:            "1.0.0",
			ArchiveStatus: "archive",
			IndexStatus:   "complete",
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

func TestCreateVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/organizations/my-org/projects/my-proj/packages/my-pkg/versions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Verify it's multipart
		ct := r.Header.Get("Content-Type")
		if ct == "" {
			t.Error("expected Content-Type header")
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Version{
			ID:            "1.0.0",
			ArchiveStatus: "unarchive",
			IndexStatus:   "pending",
			CreatedAt:     "2025-01-01T00:00:00Z",
			UpdatedAt:     "2025-01-01T00:00:00Z",
		})
	}))
	defer server.Close()

	// Create a temp file to upload
	tmpFile, err := os.CreateTemp("", "test-pkg-*.tar.gz")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Write([]byte("fake archive data"))
	tmpFile.Close()

	c := NewClient(server.URL, "ApiKey k", "org")
	result, err := c.CreateVersion("my-org", "my-proj", "my-pkg", &Version{ID: "1.0.0"}, tmpFile.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "1.0.0" {
		t.Errorf("expected ID %q, got %q", "1.0.0", result.ID)
	}
}
