package resources_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func randomName(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, rand.Intn(99999))
}

func TestAccOrganization_basic(t *testing.T) {
	orgName := randomName("test-org-tf")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			// Create and read
			{
				Config: testAccOrganizationConfig(orgName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("credible_organization.test", "name", orgName),
					resource.TestCheckResourceAttrSet("credible_organization.test", "created_at"),
					resource.TestCheckResourceAttrSet("credible_organization.test", "updated_at"),
					resource.TestCheckResourceAttr("credible_organization.test", "deletion_protection", "false"),
				),
			},
			// Import
			{
				ResourceName:            "credible_organization.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"deletion_protection", "force_cascade"},
			},
		},
	})
}

func TestAccOrganization_updateDisplayName(t *testing.T) {
	orgName := randomName("test-org-tf")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfigWithDisplayName(orgName, "Original Name"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("credible_organization.test", "display_name", "Original Name"),
				),
			},
			{
				Config: testAccOrganizationConfigWithDisplayName(orgName, "Updated Name"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("credible_organization.test", "display_name", "Updated Name"),
				),
			},
		},
	})
}

func testAccOrganizationConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "credible_organization" "test" {
  name                = %q
  deletion_protection = false
}
`, name)
}

func testAccOrganizationConfigWithDisplayName(name, displayName string) string {
	return providerConfig() + fmt.Sprintf(`
resource "credible_organization" "test" {
  name                = %q
  display_name        = %q
  deletion_protection = false
}
`, name, displayName)
}
