package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPackage_basic(t *testing.T) {
	orgName := randomName("test-org-tf")
	envName := randomName("test-env-tf")
	pkgName := randomName("test-pkg-tf")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccPackageConfig(orgName, envName, pkgName, "A test package"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("credible_package.test", "name", pkgName),
					resource.TestCheckResourceAttr("credible_package.test", "organization", orgName),
					resource.TestCheckResourceAttr("credible_package.test", "environment", envName),
					resource.TestCheckResourceAttr("credible_package.test", "description", "A test package"),
					resource.TestCheckResourceAttr("credible_package.test", "deletion_protection", "false"),
					resource.TestCheckResourceAttrSet("credible_package.test", "created_at"),
					resource.TestCheckResourceAttrSet("credible_package.test", "updated_at"),
				),
			},
			// Import
			{
				ResourceName:            "credible_package.test",
				ImportState:             true,
				ImportStateId:           fmt.Sprintf("%s/%s/%s", orgName, envName, pkgName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"deletion_protection"},
			},
		},
	})
}

func TestAccPackage_updateDescription(t *testing.T) {
	orgName := randomName("test-org-tf")
	envName := randomName("test-env-tf")
	pkgName := randomName("test-pkg-tf")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccPackageConfig(orgName, envName, pkgName, "Original"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("credible_package.test", "description", "Original"),
				),
			},
			{
				Config: testAccPackageConfig(orgName, envName, pkgName, "Updated"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("credible_package.test", "description", "Updated"),
				),
			},
		},
	})
}

func testAccPackageConfig(orgName, envName, pkgName, description string) string {
	return providerConfig() + fmt.Sprintf(`
resource "credible_organization" "test" {
  name                = %q
  deletion_protection = false
}

resource "credible_environment" "test" {
  organization        = credible_organization.test.name
  name                = %q
  deletion_protection = false
  force_cascade       = true
}

resource "credible_package" "test" {
  organization        = credible_organization.test.name
  environment         = credible_environment.test.name
  name                = %q
  description         = %q
  deletion_protection = false
}
`, orgName, envName, pkgName, description)
}
