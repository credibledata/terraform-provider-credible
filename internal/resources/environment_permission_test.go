package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccEnvironmentPermission_basic(t *testing.T) {
	orgName := randomName("test-org-tf")
	envName := randomName("test-env-tf")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentPermissionConfig(orgName, envName, "user:testuser@example.com", "modeler"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("credible_environment_permission.test", "user_group_id", "user:testuser@example.com"),
					resource.TestCheckResourceAttr("credible_environment_permission.test", "permission", "modeler"),
					resource.TestCheckResourceAttr("credible_environment_permission.test", "organization", orgName),
					resource.TestCheckResourceAttr("credible_environment_permission.test", "environment", envName),
				),
			},
			{
				ResourceName:      "credible_environment_permission.test",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/%s/user:testuser@example.com", orgName, envName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEnvironmentPermission_update(t *testing.T) {
	orgName := randomName("test-org-tf")
	envName := randomName("test-env-tf")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentPermissionConfig(orgName, envName, "user:testuser@example.com", "modeler"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("credible_environment_permission.test", "permission", "modeler"),
				),
			},
			{
				Config: testAccEnvironmentPermissionConfig(orgName, envName, "user:testuser@example.com", "viewer"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("credible_environment_permission.test", "permission", "viewer"),
				),
			},
		},
	})
}

func testAccEnvironmentPermissionConfig(orgName, envName, userGroupID, permission string) string {
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

resource "credible_environment_permission" "test" {
  organization  = credible_organization.test.name
  environment   = credible_environment.test.name
  user_group_id = %q
  permission    = %q
}
`, orgName, envName, userGroupID, permission)
}
