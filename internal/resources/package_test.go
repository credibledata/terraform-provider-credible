package resources_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Covers what credible_package can do against a live API: adopt a package that
// was published outside Terraform, then manage its metadata. There is deliberately
// no create case -- the API has no metadata-only create. The refusal is pinned by
// TestPackageCreateUnsupported, which, unlike this file, runs without credentials.
//
// The package has to exist already, so this needs CREDIBLE_PACKAGE_NAME to name a
// published one on top of the usual acceptance-test variables. It skips rather
// than creating a fixture it has no way to create.
//
// `deletion_protection` stays TRUE here, which is unusual for an acceptance test
// and load-bearing. The harness runs its post-test destroy whenever state is
// non-empty (helper/resource/testing_new.go) and offers no way to opt out, so the
// import below -- which persists state deliberately, to give the update step
// something to work on -- guarantees a teardown DELETE. Against a package this
// test did not publish, and which this provider can no longer recreate, that
// teardown would consume the fixture permanently. Leaving protection on turns it
// into a loud protected-package failure instead of a silent deletion.
func TestAccPackage_importAndUpdateDescription(t *testing.T) {
	orgName := os.Getenv("CREDIBLE_ORGANIZATION")
	envName := os.Getenv("CREDIBLE_ENVIRONMENT")
	pkgName := os.Getenv("CREDIBLE_PACKAGE_NAME")
	if orgName == "" || envName == "" || pkgName == "" {
		t.Skip("CREDIBLE_ORGANIZATION, CREDIBLE_ENVIRONMENT and CREDIBLE_PACKAGE_NAME must name " +
			"an already-published package. It is left in place -- deletion_protection is on, so the " +
			"harness's unavoidable teardown destroy fails rather than deleting it -- but point this at " +
			"a throwaway package rather than one you care about.")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config:             testAccPackageConfig(orgName, envName, pkgName, "Original"),
				ResourceName:       "credible_package.test",
				ImportState:        true,
				ImportStateId:      fmt.Sprintf("%s/%s/%s", orgName, envName, pkgName),
				ImportStatePersist: true,
			},
			{
				Config: testAccPackageConfig(orgName, envName, pkgName, "Updated"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("credible_package.test", "name", pkgName),
					resource.TestCheckResourceAttr("credible_package.test", "organization", orgName),
					resource.TestCheckResourceAttr("credible_package.test", "environment", envName),
					resource.TestCheckResourceAttr("credible_package.test", "description", "Updated"),
					resource.TestCheckResourceAttrSet("credible_package.test", "created_at"),
					resource.TestCheckResourceAttrSet("credible_package.test", "updated_at"),
				),
			},
		},
	})
}

func testAccPackageConfig(orgName, envName, pkgName, description string) string {
	return providerConfig() + fmt.Sprintf(`
resource "credible_package" "test" {
  organization        = %q
  environment         = %q
  name                = %q
  description         = %q
  deletion_protection = true
}
`, orgName, envName, pkgName, description)
}
