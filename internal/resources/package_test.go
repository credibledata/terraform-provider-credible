package resources_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// The Admin API creates a package by publishing -- one multipart request carrying
// the package, its first version and the model archive -- and offers no operation
// that creates package metadata alone. So an apply that declares a package
// Terraform has not imported must refuse, and refuse in a way that names the
// operation that does exist.
//
// Asserted without a live API on purpose. The acceptance tests that covered this
// resource before all created a package, so they skipped whenever CREDIBLE_URL was
// unset -- which is every CI run -- and a resource whose create answered HTTP 405
// on every apply went unnoticed. This test needs no credentials and so actually
// runs.
func TestPackage_createIsRefused(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccPackageCreateConfig("my-org", "analytics", "analytics-models"),
				ExpectError: regexp.MustCompile(`Creating a package is not supported`),
			},
		},
	})
}

// The refusal is only useful if it tells the reader what to do instead, so the
// message is pinned rather than just its presence: the publish path the API does
// serve, and the import that adopts a package already published.
func TestPackage_createRefusalNamesThePublishPathAndImport(t *testing.T) {
	for _, want := range []string{
		`/organizations/my-org/environments/analytics/packages/analytics-models`,
		`terraform import`,
		`my-org/analytics/analytics-models`,
	} {
		t.Run(want, func(t *testing.T) {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config:      testAccPackageCreateConfig("my-org", "analytics", "analytics-models"),
						ExpectError: regexp.MustCompile(regexp.QuoteMeta(want)),
					},
				},
			})
		})
	}
}

// Declares the package alone, with the organization on the provider: the refusal
// has to happen before any API call, so the config deliberately does not create a
// parent organization or environment it would otherwise need.
func testAccPackageCreateConfig(orgName, envName, pkgName string) string {
	return fmt.Sprintf(`
provider "credible" {
  url          = "http://localhost:1"
  organization = %q
  bearer_token = "not-used-no-request-is-made"
}

resource "credible_package" "test" {
  environment         = %q
  name                = %q
  description         = "A test package"
  deletion_protection = false
}
`, orgName, envName, pkgName)
}
