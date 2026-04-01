package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateProject(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.URL.Path != "/api/v0/organizations/my-org/projects" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}

			var body Project
			json.NewDecoder(r.Body).Decode(&body)
			if body.Name != "test-proj" {
				t.Errorf("expected name %q, got %q", "test-proj", body.Name)
			}

			rc := 2
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(Project{
				Name:             "test-proj",
				Readme:           "# Test",
				ReplicationCount: &rc,
				CreatedAt:        "2025-01-01T00:00:00Z",
				UpdatedAt:        "2025-01-01T00:00:00Z",
			})
		}))
		defer server.Close()

		c := NewClient(server.URL, "ApiKey k", "org")
		rc := 2
		result, err := c.CreateProject("my-org", &Project{Name: "test-proj", ReplicationCount: &rc})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Name != "test-proj" {
			t.Errorf("expected name %q, got %q", "test-proj", result.Name)
		}
		if result.ReplicationCount == nil || *result.ReplicationCount != 2 {
			t.Errorf("expected replication_count 2, got %v", result.ReplicationCount)
		}
	})

	t.Run("error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(APIError{Message: "invalid name"})
		}))
		defer server.Close()

		c := NewClient(server.URL, "ApiKey k", "org")
		_, err := c.CreateProject("my-org", &Project{Name: ""})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestGetProject(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("expected GET, got %s", r.Method)
			}
			if r.URL.Path != "/api/v0/organizations/my-org/projects/my-proj" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(Project{
				Name:      "my-proj",
				CreatedAt: "2025-01-01T00:00:00Z",
				UpdatedAt: "2025-01-01T00:00:00Z",
			})
		}))
		defer server.Close()

		c := NewClient(server.URL, "ApiKey k", "org")
		result, err := c.GetProject("my-org", "my-proj")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Name != "my-proj" {
			t.Errorf("expected name %q, got %q", "my-proj", result.Name)
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(APIError{Message: "not found"})
		}))
		defer server.Close()

		c := NewClient(server.URL, "ApiKey k", "org")
		_, err := c.GetProject("my-org", "missing")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !IsNotFound(err) {
			t.Error("expected IsNotFound to return true")
		}
	})
}

func TestUpdateProject(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/organizations/my-org/projects/my-proj" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Project{
			Name:      "my-proj",
			Readme:    "# Updated",
			CreatedAt: "2025-01-01T00:00:00Z",
			UpdatedAt: "2025-01-02T00:00:00Z",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "ApiKey k", "org")
	result, err := c.UpdateProject("my-org", "my-proj", &Project{Readme: "# Updated"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Readme != "# Updated" {
		t.Errorf("expected readme %q, got %q", "# Updated", result.Readme)
	}
}

func TestDeleteProject(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "DELETE" {
				t.Errorf("expected DELETE, got %s", r.Method)
			}
			if r.URL.Path != "/api/v0/organizations/my-org/projects/my-proj" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		c := NewClient(server.URL, "ApiKey k", "org")
		err := c.DeleteProject("my-org", "my-proj")
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
		err := c.DeleteProject("my-org", "my-proj")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestListProjects(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/organizations/my-org/projects" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]Project{
			{Name: "proj-1"},
			{Name: "proj-2"},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "ApiKey k", "org")
	result, err := c.ListProjects("my-org")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 projects, got %d", len(result))
	}
}
