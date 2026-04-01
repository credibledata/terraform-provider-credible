package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGroup_basic(t *testing.T) {
	orgName := randomName("test-org-tf")
	groupName := randomName("test-group-tf")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig(orgName, groupName, "A test group"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("credible_group.test", "name", groupName),
					resource.TestCheckResourceAttr("credible_group.test", "organization", orgName),
					resource.TestCheckResourceAttr("credible_group.test", "description", "A test group"),
					resource.TestCheckResourceAttrSet("credible_group.test", "created_at"),
					resource.TestCheckResourceAttrSet("credible_group.test", "updated_at"),
				),
			},
			// Import
			{
				ResourceName:      "credible_group.test",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/%s", orgName, groupName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccGroup_updateDescription(t *testing.T) {
	orgName := randomName("test-org-tf")
	groupName := randomName("test-group-tf")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig(orgName, groupName, "Original"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("credible_group.test", "description", "Original"),
				),
			},
			{
				Config: testAccGroupConfig(orgName, groupName, "Updated"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("credible_group.test", "description", "Updated"),
				),
			},
		},
	})
}

func testAccGroupConfig(orgName, groupName, description string) string {
	return providerConfig() + fmt.Sprintf(`
resource "credible_organization" "test" {
  name                = %q
  deletion_protection = false
}

resource "credible_group" "test" {
  organization = credible_organization.test.name
  name         = %q
  description  = %q
}
`, orgName, groupName, description)
}
