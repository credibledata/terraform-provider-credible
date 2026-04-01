package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccProject_basic(t *testing.T) {
	orgName := randomName("test-org-tf")
	projName := randomName("test-proj-tf")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig(orgName, projName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("credible_project.test", "name", projName),
					resource.TestCheckResourceAttr("credible_project.test", "organization", orgName),
					resource.TestCheckResourceAttrSet("credible_project.test", "created_at"),
					resource.TestCheckResourceAttrSet("credible_project.test", "updated_at"),
				),
			},
			// Import
			{
				ResourceName:            "credible_project.test",
				ImportState:             true,
				ImportStateId:           fmt.Sprintf("%s/%s", orgName, projName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"deletion_protection", "force_cascade"},
			},
		},
	})
}

func TestAccProject_updateReadme(t *testing.T) {
	orgName := randomName("test-org-tf")
	projName := randomName("test-proj-tf")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfigWithReadme(orgName, projName, "# Original"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("credible_project.test", "readme", "# Original"),
				),
			},
			{
				Config: testAccProjectConfigWithReadme(orgName, projName, "# Updated"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("credible_project.test", "readme", "# Updated"),
				),
			},
		},
	})
}

func testAccProjectConfig(orgName, projName string) string {
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
`, orgName, projName)
}

func testAccProjectConfigWithReadme(orgName, projName, readme string) string {
	return providerConfig() + fmt.Sprintf(`
resource "credible_organization" "test" {
  name                = %q
  deletion_protection = false
}

resource "credible_project" "test" {
  organization        = credible_organization.test.name
  name                = %q
  readme              = %q
  deletion_protection = false
  force_cascade       = true
}
`, orgName, projName, readme)
}
