package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateGroup(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/organizations/my-org/groups" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var body Group
		json.NewDecoder(r.Body).Decode(&body)
		if body.GroupName != "devs" {
			t.Errorf("expected groupName %q, got %q", "devs", body.GroupName)
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Group{
			OrganizationName: "my-org",
			GroupName:        "devs",
			Description:      "Developers group",
			CreatedAt:        "2025-01-01T00:00:00Z",
			UpdatedAt:        "2025-01-01T00:00:00Z",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "ApiKey k", "org")
	result, err := c.CreateGroup("my-org", &Group{GroupName: "devs", Description: "Developers group"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.GroupName != "devs" {
		t.Errorf("expected groupName %q, got %q", "devs", result.GroupName)
	}
}

func TestGetGroup(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("expected GET, got %s", r.Method)
			}
			if r.URL.Path != "/api/v0/organizations/my-org/groups/devs" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(Group{
				GroupName:   "devs",
				Description: "Developers",
				CreatedAt:   "2025-01-01T00:00:00Z",
				UpdatedAt:   "2025-01-01T00:00:00Z",
			})
		}))
		defer server.Close()

		c := NewClient(server.URL, "ApiKey k", "org")
		result, err := c.GetGroup("my-org", "devs")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.GroupName != "devs" {
			t.Errorf("expected groupName %q, got %q", "devs", result.GroupName)
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(APIError{Message: "not found"})
		}))
		defer server.Close()

		c := NewClient(server.URL, "ApiKey k", "org")
		_, err := c.GetGroup("my-org", "missing")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !IsNotFound(err) {
			t.Error("expected IsNotFound to return true")
		}
	})
}

func TestUpdateGroup(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/organizations/my-org/groups/devs" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Group{
			GroupName:   "devs",
			Description: "Updated description",
			CreatedAt:   "2025-01-01T00:00:00Z",
			UpdatedAt:   "2025-01-02T00:00:00Z",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "ApiKey k", "org")
	result, err := c.UpdateGroup("my-org", "devs", &Group{Description: "Updated description"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Description != "Updated description" {
		t.Errorf("expected description %q, got %q", "Updated description", result.Description)
	}
}

func TestDeleteGroup(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "DELETE" {
				t.Errorf("expected DELETE, got %s", r.Method)
			}
			if r.URL.Path != "/api/v0/organizations/my-org/groups/devs" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		c := NewClient(server.URL, "ApiKey k", "org")
		err := c.DeleteGroup("my-org", "devs")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestListGroupMembers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/organizations/my-org/groups/devs/members" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(GroupMembersResponse{
			Members: []GroupMember{
				{UserGroupID: "user:alice@example.com", Status: "admin"},
				{UserGroupID: "user:bob@example.com", Status: "member"},
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "ApiKey k", "org")
	result, err := c.ListGroupMembers("my-org", "devs")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 members, got %d", len(result))
	}
	if result[0].UserGroupID != "user:alice@example.com" {
		t.Errorf("expected first member %q, got %q", "user:alice@example.com", result[0].UserGroupID)
	}
}

func TestAddGroupMembers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/organizations/my-org/groups/devs/members" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var body GroupMembersRequest
		json.NewDecoder(r.Body).Decode(&body)
		if len(body.Members) != 1 {
			t.Errorf("expected 1 member, got %d", len(body.Members))
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewClient(server.URL, "ApiKey k", "org")
	err := c.AddGroupMembers("my-org", "devs", []GroupMember{
		{UserGroupID: "user:carol@example.com", Status: "member"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRemoveGroupMembers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/organizations/my-org/groups/devs/members" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewClient(server.URL, "ApiKey k", "org")
	err := c.RemoveGroupMembers("my-org", "devs", []GroupMember{
		{UserGroupID: "user:alice@example.com"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateGroupMemberStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/organizations/my-org/groups/devs/members/alice@example.com" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewClient(server.URL, "ApiKey k", "org")
	err := c.UpdateGroupMemberStatus("my-org", "devs", "alice@example.com", "admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
