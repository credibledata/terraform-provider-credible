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
func TestAccPackage_importAndUpdateDescription(t *testing.T) {
	orgName := os.Getenv("CREDIBLE_ORGANIZATION")
	envName := os.Getenv("CREDIBLE_ENVIRONMENT")
	pkgName := os.Getenv("CREDIBLE_PACKAGE_NAME")
	if orgName == "" || envName == "" || pkgName == "" {
		t.Skip("CREDIBLE_ORGANIZATION, CREDIBLE_ENVIRONMENT and CREDIBLE_PACKAGE_NAME must name an already-published package")
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
  deletion_protection = false
}
`, orgName, envName, pkgName, description)
}
