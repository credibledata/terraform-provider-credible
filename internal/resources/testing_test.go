package resources_test

import (
	"os"
	"testing"

	"github.com/credibledata/terraform-provider-credible/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories is used in resource.TestCase to instantiate the provider.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"credible": providerserver.NewProtocol6WithError(provider.New()),
}

// testAccPreCheck validates that required environment variables are set for acceptance tests.
func testAccPreCheck(t *testing.T) {
	t.Helper()
	if v := os.Getenv("CREDIBLE_URL"); v == "" {
		t.Skip("CREDIBLE_URL must be set for acceptance tests")
	}
	if v := os.Getenv("CREDIBLE_API_KEY"); v == "" {
		if v := os.Getenv("CREDIBLE_BEARER_TOKEN"); v == "" {
			t.Skip("CREDIBLE_API_KEY or CREDIBLE_BEARER_TOKEN must be set for acceptance tests")
		}
	}
}

// providerConfig returns the provider configuration block for acceptance tests.
// It relies on CREDIBLE_URL and CREDIBLE_API_KEY/CREDIBLE_BEARER_TOKEN environment variables.
func providerConfig() string {
	return `
provider "credible" {}
`
}
