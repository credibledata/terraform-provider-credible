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

// Environment permissions

func (c *Client) CreateEnvironmentPermission(org, environment string, perm *Permission) (*Permission, error) {
	var result Permission
	err := c.doJSON("POST", fmt.Sprintf("/organizations/%s/environments/%s/permissions", org, environment), perm, &result)
	if err != nil {
		return nil, fmt.Errorf("creating environment permission: %w", err)
	}
	return &result, nil
}

func (c *Client) GetEnvironmentPermission(org, environment, userGroupID string) (*Permission, error) {
	var result Permission
	err := c.doJSON("GET", fmt.Sprintf("/organizations/%s/environments/%s/permissions/%s", org, environment, userGroupID), nil, &result)
	if err != nil {
		return nil, fmt.Errorf("getting environment permission for %q: %w", userGroupID, err)
	}
	return &result, nil
}

func (c *Client) UpdateEnvironmentPermission(org, environment, userGroupID string, perm *Permission) (*Permission, error) {
	var result Permission
	err := c.doJSON("PATCH", fmt.Sprintf("/organizations/%s/environments/%s/permissions/%s", org, environment, userGroupID), perm, &result)
	if err != nil {
		return nil, fmt.Errorf("updating environment permission for %q: %w", userGroupID, err)
	}
	return &result, nil
}

func (c *Client) DeleteEnvironmentPermission(org, environment, userGroupID string) error {
	err := c.doJSON("DELETE", fmt.Sprintf("/organizations/%s/environments/%s/permissions/%s", org, environment, userGroupID), nil, nil)
	if err != nil {
		return fmt.Errorf("deleting environment permission for %q: %w", userGroupID, err)
	}
	return nil
}
