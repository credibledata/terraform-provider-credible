package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccProjectPermission_basic(t *testing.T) {
	orgName := randomName("test-org-tf")
	projName := randomName("test-proj-tf")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccProjectPermissionConfig(orgName, projName, "user:testuser@example.com", "modeler"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("credible_project_permission.test", "user_group_id", "user:testuser@example.com"),
					resource.TestCheckResourceAttr("credible_project_permission.test", "permission", "modeler"),
					resource.TestCheckResourceAttr("credible_project_permission.test", "organization", orgName),
					resource.TestCheckResourceAttr("credible_project_permission.test", "project", projName),
				),
			},
			// Import
			{
				ResourceName:      "credible_project_permission.test",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/%s/user:testuser@example.com", orgName, projName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccProjectPermission_update(t *testing.T) {
	orgName := randomName("test-org-tf")
	projName := randomName("test-proj-tf")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccProjectPermissionConfig(orgName, projName, "user:testuser@example.com", "modeler"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("credible_project_permission.test", "permission", "modeler"),
				),
			},
			{
				Config: testAccProjectPermissionConfig(orgName, projName, "user:testuser@example.com", "viewer"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("credible_project_permission.test", "permission", "viewer"),
				),
			},
		},
	})
}

func testAccProjectPermissionConfig(orgName, projName, userGroupID, permission string) string {
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

resource "credible_project_permission" "test" {
  organization  = credible_organization.test.name
  project       = credible_project.test.name
  user_group_id = %q
  permission    = %q
}
`, orgName, projName, userGroupID, permission)
}
