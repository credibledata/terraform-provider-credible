package client

import "fmt"

// Package represents a Credible package.
type Package struct {
	Name          string `json:"name,omitempty"`
	Description   string `json:"description,omitempty"`
	LatestVersion string `json:"latestVersion,omitempty"`
	CreatedAt     string `json:"createdAt,omitempty"`
	UpdatedAt     string `json:"updatedAt,omitempty"`
}

func (c *Client) ListPackages(org, environment string) ([]Package, error) {
	var result []Package
	err := c.doJSON("GET", fmt.Sprintf("/organizations/%s/environments/%s/packages", org, environment), nil, &result)
	if err != nil {
		return nil, fmt.Errorf("listing packages: %w", err)
	}
	return result, nil
}

func (c *Client) CreatePackage(org, environment string, pkg *Package) (*Package, error) {
	var result Package
	err := c.doJSON("POST", fmt.Sprintf("/organizations/%s/environments/%s/packages", org, environment), pkg, &result)
	if err != nil {
		return nil, fmt.Errorf("creating package: %w", err)
	}
	return &result, nil
}

func (c *Client) GetPackage(org, environment, name string) (*Package, error) {
	var result Package
	err := c.doJSON("GET", fmt.Sprintf("/organizations/%s/environments/%s/packages/%s", org, environment, name), nil, &result)
	if err != nil {
		return nil, fmt.Errorf("getting package %q: %w", name, err)
	}
	return &result, nil
}

func (c *Client) UpdatePackage(org, environment, name string, pkg *Package) (*Package, error) {
	var result Package
	err := c.doJSON("PATCH", fmt.Sprintf("/organizations/%s/environments/%s/packages/%s", org, environment, name), pkg, &result)
	if err != nil {
		return nil, fmt.Errorf("updating package %q: %w", name, err)
	}
	return &result, nil
}

func (c *Client) DeletePackage(org, environment, name string) error {
	err := c.doJSON("DELETE", fmt.Sprintf("/organizations/%s/environments/%s/packages/%s", org, environment, name), nil, nil)
	if err != nil {
		return fmt.Errorf("deleting package %q: %w", name, err)
	}
	return nil
}
