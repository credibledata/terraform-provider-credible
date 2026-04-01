package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateConnection(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.URL.Path != "/api/v0/organizations/my-org/projects/my-proj/connections" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}

			var body Connection
			json.NewDecoder(r.Body).Decode(&body)
			if body.Name != "my-conn" {
				t.Errorf("expected name %q, got %q", "my-conn", body.Name)
			}
			if body.Type != "postgres" {
				t.Errorf("expected type %q, got %q", "postgres", body.Type)
			}

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(Connection{
				Name:           "my-conn",
				Type:           "postgres",
				IndexingStatus: "pending",
			})
		}))
		defer server.Close()

		c := NewClient(server.URL, "ApiKey k", "org")
		conn := &Connection{
			Name: "my-conn",
			Type: "postgres",
			PostgresConnection: &PostgresConnection{
				Host:         "localhost",
				DatabaseName: "testdb",
				UserName:     "user",
				Password:     "pass",
			},
		}
		result, err := c.CreateConnection("my-org", "my-proj", conn)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Name != "my-conn" {
			t.Errorf("expected name %q, got %q", "my-conn", result.Name)
		}
	})

	t.Run("error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(APIError{Message: "invalid connection"})
		}))
		defer server.Close()

		c := NewClient(server.URL, "ApiKey k", "org")
		_, err := c.CreateConnection("my-org", "my-proj", &Connection{Name: "bad"})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestGetConnection(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		excludeAll := false
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("expected GET, got %s", r.Method)
			}
			if r.URL.Path != "/api/v0/organizations/my-org/projects/my-proj/connections/my-conn" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(Connection{
				Name:             "my-conn",
				Type:             "postgres",
				ExcludeAllTables: &excludeAll,
				IndexingStatus:   "complete",
			})
		}))
		defer server.Close()

		c := NewClient(server.URL, "ApiKey k", "org")
		result, err := c.GetConnection("my-org", "my-proj", "my-conn")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Name != "my-conn" {
			t.Errorf("expected name %q, got %q", "my-conn", result.Name)
		}
		if result.IndexingStatus != "complete" {
			t.Errorf("expected indexing_status %q, got %q", "complete", result.IndexingStatus)
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(APIError{Message: "not found"})
		}))
		defer server.Close()

		c := NewClient(server.URL, "ApiKey k", "org")
		_, err := c.GetConnection("my-org", "my-proj", "missing")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !IsNotFound(err) {
			t.Error("expected IsNotFound to return true")
		}
	})
}

func TestUpdateConnection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/organizations/my-org/projects/my-proj/connections/my-conn" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Connection{
			Name:           "my-conn",
			Type:           "postgres",
			IndexingStatus: "pending",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "ApiKey k", "org")
	result, err := c.UpdateConnection("my-org", "my-proj", "my-conn", &Connection{
		PostgresConnection: &PostgresConnection{Host: "newhost"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "my-conn" {
		t.Errorf("expected name %q, got %q", "my-conn", result.Name)
	}
}

func TestDeleteConnection(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "DELETE" {
				t.Errorf("expected DELETE, got %s", r.Method)
			}
			if r.URL.Path != "/api/v0/organizations/my-org/projects/my-proj/connections/my-conn" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		c := NewClient(server.URL, "ApiKey k", "org")
		err := c.DeleteConnection("my-org", "my-proj", "my-conn")
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
		err := c.DeleteConnection("my-org", "my-proj", "my-conn")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestListConnections(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]Connection{
			{Name: "conn-1", Type: "postgres"},
			{Name: "conn-2", Type: "bigquery"},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "ApiKey k", "org")
	result, err := c.ListConnections("my-org", "my-proj")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 connections, got %d", len(result))
	}
}
