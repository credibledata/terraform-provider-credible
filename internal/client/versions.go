package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

// Version represents a Credible package version.
type Version struct {
	ID            string `json:"id,omitempty"`
	ArchiveStatus string `json:"archiveStatus,omitempty"`
	IndexStatus   string `json:"indexStatus,omitempty"`
	CreatedAt     string `json:"createdAt,omitempty"`
	UpdatedAt     string `json:"updatedAt,omitempty"`
}

// CreateVersion publishes a new package version via multipart upload.
func (c *Client) CreateVersion(org, project, pkg string, version *Version, filePath string) (*Version, error) {
	url := fmt.Sprintf("%s/api/v0/organizations/%s/projects/%s/packages/%s/versions", c.BaseURL, org, project, pkg)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add the version JSON as the "body" part
	versionJSON, err := json.Marshal(version)
	if err != nil {
		return nil, fmt.Errorf("marshaling version: %w", err)
	}
	bodyPart, err := writer.CreateFormField("body")
	if err != nil {
		return nil, fmt.Errorf("creating body form field: %w", err)
	}
	if _, err := bodyPart.Write(versionJSON); err != nil {
		return nil, fmt.Errorf("writing version JSON: %w", err)
	}

	// Add the file part
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("opening file %q: %w", filePath, err)
	}
	defer file.Close()

	filePart, err := writer.CreateFormFile("file", "package.tar.gz")
	if err != nil {
		return nil, fmt.Errorf("creating file form field: %w", err)
	}
	if _, err := io.Copy(filePart, file); err != nil {
		return nil, fmt.Errorf("copying file data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("closing multipart writer: %w", err)
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	respBody, statusCode, err := c.doRequestRaw(req)
	if err != nil {
		return nil, fmt.Errorf("uploading version: %w", err)
	}

	if statusCode < 200 || statusCode >= 300 {
		return nil, fmt.Errorf("API error (HTTP %d): %s", statusCode, string(respBody))
	}

	var result Version
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling version response: %w", err)
	}

	return &result, nil
}

func (c *Client) GetVersion(org, project, pkg, versionID string) (*Version, error) {
	var result Version
	err := c.doJSON("GET", fmt.Sprintf("/organizations/%s/projects/%s/packages/%s/versions/%s", org, project, pkg, versionID), nil, &result)
	if err != nil {
		return nil, fmt.Errorf("getting version %q: %w", versionID, err)
	}
	return &result, nil
}

func (c *Client) UpdateVersion(org, project, pkg, versionID string, version *Version) (*Version, error) {
	var result Version
	err := c.doJSON("PATCH", fmt.Sprintf("/organizations/%s/projects/%s/packages/%s/versions/%s", org, project, pkg, versionID), version, &result)
	if err != nil {
		return nil, fmt.Errorf("updating version %q: %w", versionID, err)
	}
	return &result, nil
}
