package resources

import (
	"strings"
	"testing"
)

// The Admin API creates a package by publishing -- one multipart request carrying
// the package, its first version and the model archive -- and offers no operation
// that creates package metadata alone. So credible_package refuses to create, and
// the refusal has to tell the reader what to do instead.
//
// Asserted against the message builder rather than through the provider test
// harness, which downloads and runs the real Terraform CLI: CI has no Terraform
// binary, so a harness-based test fails there on the download instead of
// exercising anything. That distinction is the reason this resource's create was
// broken for so long -- every test covering it either needed credentials or
// needed the CLI, so none of them ran.
func TestPackageCreateUnsupported(t *testing.T) {
	summary, detail := packageCreateUnsupported("my-org", "analytics", "analytics-models")

	if summary != "Creating a package is not supported" {
		t.Errorf("summary = %q, want the create-unsupported summary", summary)
	}

	// Each of these is the part that makes the refusal actionable: what the API
	// does instead, the exact path that serves it, and the import that adopts a
	// package already published. A message that drops any of them sends the
	// practitioner back to guessing.
	for _, want := range []string{
		"creates a package and its first version together",
		"/organizations/my-org/environments/analytics/packages/analytics-models",
		"terraform import <address> my-org/analytics/analytics-models",
		"cred publish",
	} {
		if !strings.Contains(detail, want) {
			t.Errorf("detail does not mention %q:\n%s", want, detail)
		}
	}
}

// The coordinates are interpolated, so a refusal must name the resource actually
// declared rather than a fixed example -- otherwise the import command it hands
// back adopts the wrong package.
func TestPackageCreateUnsupportedNamesItsOwnCoordinates(t *testing.T) {
	_, detail := packageCreateUnsupported("other-org", "staging", "sales-models")

	for _, want := range []string{
		"/organizations/other-org/environments/staging/packages/sales-models",
		"terraform import <address> other-org/staging/sales-models",
	} {
		if !strings.Contains(detail, want) {
			t.Errorf("detail does not mention %q:\n%s", want, detail)
		}
	}
	for _, unwanted := range []string{"my-org", "analytics", "analytics-models"} {
		if strings.Contains(detail, unwanted) {
			t.Errorf("detail leaked the other test's %q:\n%s", unwanted, detail)
		}
	}
}
