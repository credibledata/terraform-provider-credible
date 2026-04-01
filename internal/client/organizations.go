package client

import "fmt"

// Organization represents a Credible organization.
type Organization struct {
	Name        string `json:"name,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	CreatedAt   string `json:"createdAt,omitempty"`
	UpdatedAt   string `json:"updatedAt,omitempty"`
}

func (c *Client) CreateOrganization(org *Organization) (*Organization, error) {
	var result Organization
	err := c.doJSON("POST", "/organizations", org, &result)
	if err != nil {
		return nil, fmt.Errorf("creating organization: %w", err)
	}
	return &result, nil
}

func (c *Client) GetOrganization(name string) (*Organization, error) {
	var result Organization
	err := c.doJSON("GET", fmt.Sprintf("/organizations/%s", name), nil, &result)
	if err != nil {
		return nil, fmt.Errorf("getting organization %q: %w", name, err)
	}
	return &result, nil
}

func (c *Client) UpdateOrganization(name string, org *Organization) (*Organization, error) {
	var result Organization
	err := c.doJSON("PATCH", fmt.Sprintf("/organizations/%s", name), org, &result)
	if err != nil {
		return nil, fmt.Errorf("updating organization %q: %w", name, err)
	}
	return &result, nil
}
