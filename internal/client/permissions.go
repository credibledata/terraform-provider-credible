package client

import "fmt"

// Permission represents a permission assignment for a user or group.
type Permission struct {
	UserGroupID string `json:"userGroupId,omitempty"`
	Permission  string `json:"permission,omitempty"`
	CreatedAt   string `json:"createdAt,omitempty"`
	UpdatedAt   string `json:"updatedAt,omitempty"`
}

// Organization Permissions

func (c *Client) CreateOrgPermission(org string, perm *Permission) (*Permission, error) {
	var result Permission
	err := c.doJSON("POST", fmt.Sprintf("/organizations/%s/permissions", org), perm, &result)
	if err != nil {
		return nil, fmt.Errorf("creating org permission: %w", err)
	}
	return &result, nil
}

func (c *Client) GetOrgPermission(org, userGroupID string) (*Permission, error) {
	var result Permission
	err := c.doJSON("GET", fmt.Sprintf("/organizations/%s/permissions/%s", org, userGroupID), nil, &result)
	if err != nil {
		return nil, fmt.Errorf("getting org permission for %q: %w", userGroupID, err)
	}
	return &result, nil
}

func (c *Client) UpdateOrgPermission(org, userGroupID string, perm *Permission) (*Permission, error) {
	var result Permission
	err := c.doJSON("PATCH", fmt.Sprintf("/organizations/%s/permissions/%s", org, userGroupID), perm, &result)
	if err != nil {
		return nil, fmt.Errorf("updating org permission for %q: %w", userGroupID, err)
	}
	return &result, nil
}

func (c *Client) DeleteOrgPermission(org, userGroupID string) error {
	err := c.doJSON("DELETE", fmt.Sprintf("/organizations/%s/permissions/%s", org, userGroupID), nil, nil)
	if err != nil {
		return fmt.Errorf("deleting org permission for %q: %w", userGroupID, err)
	}
	return nil
}

// Project Permissions

func (c *Client) CreateProjectPermission(org, project string, perm *Permission) (*Permission, error) {
	var result Permission
	err := c.doJSON("POST", fmt.Sprintf("/organizations/%s/projects/%s/permissions", org, project), perm, &result)
	if err != nil {
		return nil, fmt.Errorf("creating project permission: %w", err)
	}
	return &result, nil
}

func (c *Client) GetProjectPermission(org, project, userGroupID string) (*Permission, error) {
	var result Permission
	err := c.doJSON("GET", fmt.Sprintf("/organizations/%s/projects/%s/permissions/%s", org, project, userGroupID), nil, &result)
	if err != nil {
		return nil, fmt.Errorf("getting project permission for %q: %w", userGroupID, err)
	}
	return &result, nil
}

func (c *Client) UpdateProjectPermission(org, project, userGroupID string, perm *Permission) (*Permission, error) {
	var result Permission
	err := c.doJSON("PATCH", fmt.Sprintf("/organizations/%s/projects/%s/permissions/%s", org, project, userGroupID), perm, &result)
	if err != nil {
		return nil, fmt.Errorf("updating project permission for %q: %w", userGroupID, err)
	}
	return &result, nil
}

func (c *Client) DeleteProjectPermission(org, project, userGroupID string) error {
	err := c.doJSON("DELETE", fmt.Sprintf("/organizations/%s/projects/%s/permissions/%s", org, project, userGroupID), nil, nil)
	if err != nil {
		return fmt.Errorf("deleting project permission for %q: %w", userGroupID, err)
	}
	return nil
}
