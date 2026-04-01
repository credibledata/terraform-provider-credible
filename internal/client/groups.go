package client

import "fmt"

// Group represents a group in an organization.
type Group struct {
	OrganizationName string `json:"organizationName,omitempty"`
	GroupName        string `json:"groupName,omitempty"`
	Description      string `json:"description,omitempty"`
	CreatedAt        string `json:"createdAt,omitempty"`
	UpdatedAt        string `json:"updatedAt,omitempty"`
}

func (c *Client) CreateGroup(org string, group *Group) (*Group, error) {
	var result Group
	err := c.doJSON("POST", fmt.Sprintf("/organizations/%s/groups", org), group, &result)
	if err != nil {
		return nil, fmt.Errorf("creating group: %w", err)
	}
	return &result, nil
}

func (c *Client) GetGroup(org, groupName string) (*Group, error) {
	var result Group
	err := c.doJSON("GET", fmt.Sprintf("/organizations/%s/groups/%s", org, groupName), nil, &result)
	if err != nil {
		return nil, fmt.Errorf("getting group %q: %w", groupName, err)
	}
	return &result, nil
}

func (c *Client) UpdateGroup(org, groupName string, group *Group) (*Group, error) {
	var result Group
	err := c.doJSON("PATCH", fmt.Sprintf("/organizations/%s/groups/%s", org, groupName), group, &result)
	if err != nil {
		return nil, fmt.Errorf("updating group %q: %w", groupName, err)
	}
	return &result, nil
}

func (c *Client) DeleteGroup(org, groupName string) error {
	err := c.doJSON("DELETE", fmt.Sprintf("/organizations/%s/groups/%s", org, groupName), nil, nil)
	if err != nil {
		return fmt.Errorf("deleting group %q: %w", groupName, err)
	}
	return nil
}
