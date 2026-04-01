package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOrganizationPermission_basic(t *testing.T) {
	orgName := randomName("test-org-tf")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationPermissionConfig(orgName, "user:testuser@example.com", "admin"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("credible_organization_permission.test", "user_group_id", "user:testuser@example.com"),
					resource.TestCheckResourceAttr("credible_organization_permission.test", "permission", "admin"),
					resource.TestCheckResourceAttr("credible_organization_permission.test", "organization", orgName),
				),
			},
			// Import
			{
				ResourceName:      "credible_organization_permission.test",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/user:testuser@example.com", orgName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccOrganizationPermission_update(t *testing.T) {
	orgName := randomName("test-org-tf")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationPermissionConfig(orgName, "user:testuser@example.com", "admin"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("credible_organization_permission.test", "permission", "admin"),
				),
			},
			{
				Config: testAccOrganizationPermissionConfig(orgName, "user:testuser@example.com", "member"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("credible_organization_permission.test", "permission", "member"),
				),
			},
		},
	})
}

func testAccOrganizationPermissionConfig(orgName, userGroupID, permission string) string {
	return providerConfig() + fmt.Sprintf(`
resource "credible_organization" "test" {
  name                = %q
  deletion_protection = false
}

resource "credible_organization_permission" "test" {
  organization  = credible_organization.test.name
  user_group_id = %q
  permission    = %q
}
`, orgName, userGroupID, permission)
}
