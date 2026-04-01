package client

import "fmt"

// Package represents a Credible package.
type Package struct {
	Name           string `json:"name,omitempty"`
	Description    string `json:"description,omitempty"`
	LatestVersion  string `json:"latestVersion,omitempty"`
	CreatedAt      string `json:"createdAt,omitempty"`
	UpdatedAt      string `json:"updatedAt,omitempty"`
}

func (c *Client) ListPackages(org, project string) ([]Package, error) {
	var result []Package
	err := c.doJSON("GET", fmt.Sprintf("/organizations/%s/projects/%s/packages", org, project), nil, &result)
	if err != nil {
		return nil, fmt.Errorf("listing packages: %w", err)
	}
	return result, nil
}

func (c *Client) CreatePackage(org, project string, pkg *Package) (*Package, error) {
	var result Package
	err := c.doJSON("POST", fmt.Sprintf("/organizations/%s/projects/%s/packages", org, project), pkg, &result)
	if err != nil {
		return nil, fmt.Errorf("creating package: %w", err)
	}
	return &result, nil
}

func (c *Client) GetPackage(org, project, name string) (*Package, error) {
	var result Package
	err := c.doJSON("GET", fmt.Sprintf("/organizations/%s/projects/%s/packages/%s", org, project, name), nil, &result)
	if err != nil {
		return nil, fmt.Errorf("getting package %q: %w", name, err)
	}
	return &result, nil
}

func (c *Client) UpdatePackage(org, project, name string, pkg *Package) (*Package, error) {
	var result Package
	err := c.doJSON("PATCH", fmt.Sprintf("/organizations/%s/projects/%s/packages/%s", org, project, name), pkg, &result)
	if err != nil {
		return nil, fmt.Errorf("updating package %q: %w", name, err)
	}
	return &result, nil
}

func (c *Client) DeletePackage(org, project, name string) error {
	err := c.doJSON("DELETE", fmt.Sprintf("/organizations/%s/projects/%s/packages/%s", org, project, name), nil, nil)
	if err != nil {
		return fmt.Errorf("deleting package %q: %w", name, err)
	}
	return nil
}
