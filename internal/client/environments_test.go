package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateEnvironment(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.URL.Path != "/api/v0/organizations/my-org/environments" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}

			var body Environment
			json.NewDecoder(r.Body).Decode(&body)
			if body.Name != "test-env" {
				t.Errorf("expected name %q, got %q", "test-env", body.Name)
			}

			rc := 2
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(Environment{
				Name:             "test-env",
				Readme:           "# Test",
				ReplicationCount: &rc,
				CreatedAt:        "2025-01-01T00:00:00Z",
				UpdatedAt:        "2025-01-01T00:00:00Z",
			})
		}))
		defer server.Close()

		c := NewClient(server.URL, "ApiKey k", "org")
		rc := 2
		result, err := c.CreateEnvironment("my-org", &Environment{Name: "test-env", ReplicationCount: &rc})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Name != "test-env" {
			t.Errorf("expected name %q, got %q", "test-env", result.Name)
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
		_, err := c.CreateEnvironment("my-org", &Environment{Name: ""})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestGetEnvironment(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("expected GET, got %s", r.Method)
			}
			if r.URL.Path != "/api/v0/organizations/my-org/environments/my-env" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(Environment{
				Name:      "my-env",
				CreatedAt: "2025-01-01T00:00:00Z",
				UpdatedAt: "2025-01-01T00:00:00Z",
			})
		}))
		defer server.Close()

		c := NewClient(server.URL, "ApiKey k", "org")
		result, err := c.GetEnvironment("my-org", "my-env")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Name != "my-env" {
			t.Errorf("expected name %q, got %q", "my-env", result.Name)
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(APIError{Message: "not found"})
		}))
		defer server.Close()

		c := NewClient(server.URL, "ApiKey k", "org")
		_, err := c.GetEnvironment("my-org", "missing")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !IsNotFound(err) {
			t.Error("expected IsNotFound to return true")
		}
	})
}

func TestUpdateEnvironment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/organizations/my-org/environments/my-env" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Environment{
			Name:      "my-env",
			Readme:    "# Updated",
			CreatedAt: "2025-01-01T00:00:00Z",
			UpdatedAt: "2025-01-02T00:00:00Z",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "ApiKey k", "org")
	result, err := c.UpdateEnvironment("my-org", "my-env", &Environment{Readme: "# Updated"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Readme != "# Updated" {
		t.Errorf("expected readme %q, got %q", "# Updated", result.Readme)
	}
}

func TestDeleteEnvironment(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "DELETE" {
				t.Errorf("expected DELETE, got %s", r.Method)
			}
			if r.URL.Path != "/api/v0/organizations/my-org/environments/my-env" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		c := NewClient(server.URL, "ApiKey k", "org")
		err := c.DeleteEnvironment("my-org", "my-env")
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
		err := c.DeleteEnvironment("my-org", "my-env")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestListEnvironments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/organizations/my-org/environments" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]Environment{
			{Name: "env-1"},
			{Name: "env-2"},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "ApiKey k", "org")
	result, err := c.ListEnvironments("my-org")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 environments, got %d", len(result))
	}
}
