package client

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
)

// Version represents a Credible package version.
type Version struct {
	ID            string `json:"id,omitempty"`
	ArchiveStatus string `json:"archiveStatus,omitempty"`
	BuildStatus   string `json:"buildStatus,omitempty"`
	CreatedAt     string `json:"createdAt,omitempty"`
	UpdatedAt     string `json:"updatedAt,omitempty"`
}

// PublishPackageVersion publishes a package version, which is also how a package
// is created: the Admin API has no metadata-only create and no endpoint that adds
// a version to an existing package. One multipart POST to the package's own path
// (createOrUpdatePackage) carries the whole thing.
//
// Two shapes here are the API's, not preferences. The parts are named `package`,
// `version`, `packageFile` and `md5Hash` -- omitting `package` fails the request
// ("Create package request must have a Package JSON"). And packageFile must be a
// ZIP: the publisher reads the archive's bytes, so a tar.gz fails with a 500
// whatever content type it is declared as.
//
// The response is the Package, not the Version -- callers wanting version fields
// read them back with GetVersion.
func (c *Client) PublishPackageVersion(org, environment, pkg string, description string, version *Version, zipPath string) (*Package, error) {
	url := fmt.Sprintf("%s/api/v0/organizations/%s/environments/%s/packages/%s", c.BaseURL, org, environment, pkg)

	file, err := os.Open(zipPath)
	if err != nil {
		return nil, fmt.Errorf("opening package archive %q: %w", zipPath, err)
	}
	defer file.Close()

	// The API verifies the upload against md5Hash, so it is computed from the
	// same bytes that get sent rather than supplied by the caller.
	hasher := md5.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return nil, fmt.Errorf("hashing package archive %q: %w", zipPath, err)
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("rewinding package archive %q: %w", zipPath, err)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	packageJSON, err := json.Marshal(&Package{Name: pkg, Description: description})
	if err != nil {
		return nil, fmt.Errorf("marshaling package: %w", err)
	}
	if err := writeJSONPart(writer, "package", packageJSON); err != nil {
		return nil, err
	}

	versionJSON, err := json.Marshal(version)
	if err != nil {
		return nil, fmt.Errorf("marshaling version: %w", err)
	}
	if err := writeJSONPart(writer, "version", versionJSON); err != nil {
		return nil, err
	}

	filePart, err := writer.CreateFormFile("packageFile", filepath.Base(zipPath))
	if err != nil {
		return nil, fmt.Errorf("creating packageFile form field: %w", err)
	}
	if _, err := io.Copy(filePart, file); err != nil {
		return nil, fmt.Errorf("copying package archive data: %w", err)
	}

	if err := writer.WriteField("md5Hash", hex.EncodeToString(hasher.Sum(nil))); err != nil {
		return nil, fmt.Errorf("writing md5Hash field: %w", err)
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
		return nil, fmt.Errorf("publishing package version: %w", err)
	}

	if statusCode < 200 || statusCode >= 300 {
		var apiErr APIError
		if json.Unmarshal(respBody, &apiErr) == nil && apiErr.Message != "" {
			return nil, fmt.Errorf("API error (HTTP %d): %s", statusCode, apiErr.Message)
		}
		return nil, fmt.Errorf("API error (HTTP %d): %s", statusCode, string(respBody))
	}

	var result Package
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling package response: %w", err)
	}

	return &result, nil
}

// writeJSONPart adds a part the API parses as JSON. multipart.Writer.WriteField
// would send it as text/plain, which the API rejects.
func writeJSONPart(writer *multipart.Writer, name string, payload []byte) error {
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name=%q`, name))
	header.Set("Content-Type", "application/json")

	part, err := writer.CreatePart(header)
	if err != nil {
		return fmt.Errorf("creating %s form field: %w", name, err)
	}
	if _, err := part.Write(payload); err != nil {
		return fmt.Errorf("writing %s JSON: %w", name, err)
	}
	return nil
}

func (c *Client) GetVersion(org, environment, pkg, versionID string) (*Version, error) {
	var result Version
	err := c.doJSON("GET", fmt.Sprintf("/organizations/%s/environments/%s/packages/%s/versions/%s", org, environment, pkg, versionID), nil, &result)
	if err != nil {
		return nil, fmt.Errorf("getting version %q: %w", versionID, err)
	}
	return &result, nil
}

func (c *Client) UpdateVersion(org, environment, pkg, versionID string, version *Version) (*Version, error) {
	var result Version
	err := c.doJSON("PATCH", fmt.Sprintf("/organizations/%s/environments/%s/packages/%s/versions/%s", org, environment, pkg, versionID), version, &result)
	if err != nil {
		return nil, fmt.Errorf("updating version %q: %w", versionID, err)
	}
	return &result, nil
}
