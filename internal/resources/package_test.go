package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPackage_basic(t *testing.T) {
	orgName := randomName("test-org-tf")
	projName := randomName("test-proj-tf")
	pkgName := randomName("test-pkg-tf")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccPackageConfig(orgName, projName, pkgName, "A test package"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("credible_package.test", "name", pkgName),
					resource.TestCheckResourceAttr("credible_package.test", "organization", orgName),
					resource.TestCheckResourceAttr("credible_package.test", "project", projName),
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
				ImportStateId:           fmt.Sprintf("%s/%s/%s", orgName, projName, pkgName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"deletion_protection"},
			},
		},
	})
}

func TestAccPackage_updateDescription(t *testing.T) {
	orgName := randomName("test-org-tf")
	projName := randomName("test-proj-tf")
	pkgName := randomName("test-pkg-tf")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccPackageConfig(orgName, projName, pkgName, "Original"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("credible_package.test", "description", "Original"),
				),
			},
			{
				Config: testAccPackageConfig(orgName, projName, pkgName, "Updated"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("credible_package.test", "description", "Updated"),
				),
			},
		},
	})
}

func testAccPackageConfig(orgName, projName, pkgName, description string) string {
	return providerConfig() + fmt.Sprintf(`
resource "credible_organization" "test" {
  name                = %q
  deletion_protection = false
}

resource "credible_project" "test" {
  organization        = credible_organization.test.name
  name                = %q
  deletion_protection = false
  force_cascade       = true
}

resource "credible_package" "test" {
  organization        = credible_organization.test.name
  project             = credible_project.test.name
  name                = %q
  description         = %q
  deletion_protection = false
}
`, orgName, projName, pkgName, description)
}
