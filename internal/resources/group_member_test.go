package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGroupMember_basic(t *testing.T) {
	orgName := randomName("test-org-tf")
	groupName := randomName("test-group-tf")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccGroupMemberConfig(orgName, groupName, "user:testmember@example.com", "member"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("credible_group_member.test", "group_name", groupName),
					resource.TestCheckResourceAttr("credible_group_member.test", "user_group_id", "user:testmember@example.com"),
					resource.TestCheckResourceAttr("credible_group_member.test", "status", "member"),
					resource.TestCheckResourceAttr("credible_group_member.test", "organization", orgName),
				),
			},
			// Import
			{
				ResourceName:      "credible_group_member.test",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/%s/user:testmember@example.com", orgName, groupName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccGroupMember_updateStatus(t *testing.T) {
	orgName := randomName("test-org-tf")
	groupName := randomName("test-group-tf")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccGroupMemberConfig(orgName, groupName, "user:testmember@example.com", "member"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("credible_group_member.test", "status", "member"),
				),
			},
			{
				Config: testAccGroupMemberConfig(orgName, groupName, "user:testmember@example.com", "admin"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("credible_group_member.test", "status", "admin"),
				),
			},
		},
	})
}

func testAccGroupMemberConfig(orgName, groupName, userGroupID, status string) string {
	return providerConfig() + fmt.Sprintf(`
resource "credible_organization" "test" {
  name                = %q
  deletion_protection = false
}

resource "credible_group" "test" {
  organization = credible_organization.test.name
  name         = %q
  description  = "Test group for member tests"
}

resource "credible_group_member" "test" {
  organization  = credible_organization.test.name
  group_name    = credible_group.test.name
  user_group_id = %q
  status        = %q
}
`, orgName, groupName, userGroupID, status)
}
