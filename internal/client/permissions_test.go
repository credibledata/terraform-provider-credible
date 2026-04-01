package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// --- Organization Permissions ---

func TestCreateOrgPermission(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/organizations/my-org/permissions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var body Permission
		json.NewDecoder(r.Body).Decode(&body)
		if body.UserGroupID != "user:alice@example.com" {
			t.Errorf("unexpected userGroupId: %s", body.UserGroupID)
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Permission{
			UserGroupID: "user:alice@example.com",
			Permission:  "admin",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "ApiKey k", "org")
	result, err := c.CreateOrgPermission("my-org", &Permission{
		UserGroupID: "user:alice@example.com",
		Permission:  "admin",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Permission != "admin" {
		t.Errorf("expected permission %q, got %q", "admin", result.Permission)
	}
}

func TestGetOrgPermission(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("expected GET, got %s", r.Method)
			}
			if r.URL.Path != "/api/v0/organizations/my-org/permissions/user:alice@example.com" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(Permission{
				UserGroupID: "user:alice@example.com",
				Permission:  "admin",
			})
		}))
		defer server.Close()

		c := NewClient(server.URL, "ApiKey k", "org")
		result, err := c.GetOrgPermission("my-org", "user:alice@example.com")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Permission != "admin" {
			t.Errorf("expected permission %q, got %q", "admin", result.Permission)
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(APIError{Message: "not found"})
		}))
		defer server.Close()

		c := NewClient(server.URL, "ApiKey k", "org")
		_, err := c.GetOrgPermission("my-org", "user:missing@example.com")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !IsNotFound(err) {
			t.Error("expected IsNotFound to return true")
		}
	})
}

func TestUpdateOrgPermission(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Permission{
			UserGroupID: "user:alice@example.com",
			Permission:  "member",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "ApiKey k", "org")
	result, err := c.UpdateOrgPermission("my-org", "user:alice@example.com", &Permission{Permission: "member"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Permission != "member" {
		t.Errorf("expected permission %q, got %q", "member", result.Permission)
	}
}

func TestDeleteOrgPermission(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "DELETE" {
				t.Errorf("expected DELETE, got %s", r.Method)
			}
			if r.URL.Path != "/api/v0/organizations/my-org/permissions/user:alice@example.com" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		c := NewClient(server.URL, "ApiKey k", "org")
		err := c.DeleteOrgPermission("my-org", "user:alice@example.com")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

// --- Project Permissions ---

func TestCreateProjectPermission(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/organizations/my-org/projects/my-proj/permissions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Permission{
			UserGroupID: "group:devs",
			Permission:  "modeler",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "ApiKey k", "org")
	result, err := c.CreateProjectPermission("my-org", "my-proj", &Permission{
		UserGroupID: "group:devs",
		Permission:  "modeler",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Permission != "modeler" {
		t.Errorf("expected permission %q, got %q", "modeler", result.Permission)
	}
}

func TestGetProjectPermission(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("expected GET, got %s", r.Method)
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(Permission{
				UserGroupID: "group:devs",
				Permission:  "modeler",
			})
		}))
		defer server.Close()

		c := NewClient(server.URL, "ApiKey k", "org")
		result, err := c.GetProjectPermission("my-org", "my-proj", "group:devs")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.UserGroupID != "group:devs" {
			t.Errorf("expected userGroupId %q, got %q", "group:devs", result.UserGroupID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(APIError{Message: "not found"})
		}))
		defer server.Close()

		c := NewClient(server.URL, "ApiKey k", "org")
		_, err := c.GetProjectPermission("my-org", "my-proj", "group:missing")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !IsNotFound(err) {
			t.Error("expected IsNotFound to return true")
		}
	})
}

func TestUpdateProjectPermission(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Permission{
			UserGroupID: "group:devs",
			Permission:  "viewer",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "ApiKey k", "org")
	result, err := c.UpdateProjectPermission("my-org", "my-proj", "group:devs", &Permission{Permission: "viewer"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Permission != "viewer" {
		t.Errorf("expected permission %q, got %q", "viewer", result.Permission)
	}
}

func TestDeleteProjectPermission(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/organizations/my-org/projects/my-proj/permissions/group:devs" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := NewClient(server.URL, "ApiKey k", "org")
	err := c.DeleteProjectPermission("my-org", "my-proj", "group:devs")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
