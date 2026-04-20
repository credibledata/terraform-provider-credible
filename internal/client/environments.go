package client

import "fmt"

// Environment represents a Credible environment (data modeling container within an organization).
type Environment struct {
	Name             string `json:"name,omitempty"`
	Readme           string `json:"readme,omitempty"`
	ReplicationCount *int   `json:"replicationCount,omitempty"`
	CreatedAt        string `json:"createdAt,omitempty"`
	UpdatedAt        string `json:"updatedAt,omitempty"`
}

func (c *Client) ListEnvironments(org string) ([]Environment, error) {
	var result []Environment
	err := c.doJSON("GET", fmt.Sprintf("/organizations/%s/environments", org), nil, &result)
	if err != nil {
		return nil, fmt.Errorf("listing environments: %w", err)
	}
	return result, nil
}

func (c *Client) CreateEnvironment(org string, env *Environment) (*Environment, error) {
	var result Environment
	err := c.doJSON("POST", fmt.Sprintf("/organizations/%s/environments", org), env, &result)
	if err != nil {
		return nil, fmt.Errorf("creating environment: %w", err)
	}
	return &result, nil
}

func (c *Client) GetEnvironment(org, name string) (*Environment, error) {
	var result Environment
	err := c.doJSON("GET", fmt.Sprintf("/organizations/%s/environments/%s", org, name), nil, &result)
	if err != nil {
		return nil, fmt.Errorf("getting environment %q: %w", name, err)
	}
	return &result, nil
}

func (c *Client) UpdateEnvironment(org, name string, env *Environment) (*Environment, error) {
	var result Environment
	err := c.doJSON("PATCH", fmt.Sprintf("/organizations/%s/environments/%s", org, name), env, &result)
	if err != nil {
		return nil, fmt.Errorf("updating environment %q: %w", name, err)
	}
	return &result, nil
}

func (c *Client) DeleteEnvironment(org, name string) error {
	err := c.doJSON("DELETE", fmt.Sprintf("/organizations/%s/environments/%s", org, name), nil, nil)
	if err != nil {
		return fmt.Errorf("deleting environment %q: %w", name, err)
	}
	return nil
}
