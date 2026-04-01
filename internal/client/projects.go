package client

import "fmt"

// Project represents a Credible project.
type Project struct {
	Name             string `json:"name,omitempty"`
	Readme           string `json:"readme,omitempty"`
	ReplicationCount *int   `json:"replicationCount,omitempty"`
	CreatedAt        string `json:"createdAt,omitempty"`
	UpdatedAt        string `json:"updatedAt,omitempty"`
}

func (c *Client) ListProjects(org string) ([]Project, error) {
	var result []Project
	err := c.doJSON("GET", fmt.Sprintf("/organizations/%s/projects", org), nil, &result)
	if err != nil {
		return nil, fmt.Errorf("listing projects: %w", err)
	}
	return result, nil
}

func (c *Client) CreateProject(org string, project *Project) (*Project, error) {
	var result Project
	err := c.doJSON("POST", fmt.Sprintf("/organizations/%s/projects", org), project, &result)
	if err != nil {
		return nil, fmt.Errorf("creating project: %w", err)
	}
	return &result, nil
}

func (c *Client) GetProject(org, name string) (*Project, error) {
	var result Project
	err := c.doJSON("GET", fmt.Sprintf("/organizations/%s/projects/%s", org, name), nil, &result)
	if err != nil {
		return nil, fmt.Errorf("getting project %q: %w", name, err)
	}
	return &result, nil
}

func (c *Client) UpdateProject(org, name string, project *Project) (*Project, error) {
	var result Project
	err := c.doJSON("PATCH", fmt.Sprintf("/organizations/%s/projects/%s", org, name), project, &result)
	if err != nil {
		return nil, fmt.Errorf("updating project %q: %w", name, err)
	}
	return &result, nil
}

func (c *Client) DeleteProject(org, name string) error {
	err := c.doJSON("DELETE", fmt.Sprintf("/organizations/%s/projects/%s", org, name), nil, nil)
	if err != nil {
		return fmt.Errorf("deleting project %q: %w", name, err)
	}
	return nil
}
