package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccEnvironment_basic(t *testing.T) {
	orgName := randomName("test-org-tf")
	envName := randomName("test-env-tf")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig(orgName, envName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("credible_environment.test", "name", envName),
					resource.TestCheckResourceAttr("credible_environment.test", "organization", orgName),
					resource.TestCheckResourceAttrSet("credible_environment.test", "created_at"),
					resource.TestCheckResourceAttrSet("credible_environment.test", "updated_at"),
				),
			},
			{
				ResourceName:            "credible_environment.test",
				ImportState:             true,
				ImportStateId:           fmt.Sprintf("%s/%s", orgName, envName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"deletion_protection", "force_cascade"},
			},
		},
	})
}

func TestAccEnvironment_updateReadme(t *testing.T) {
	orgName := randomName("test-org-tf")
	envName := randomName("test-env-tf")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfigWithReadme(orgName, envName, "# Original"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("credible_environment.test", "readme", "# Original"),
				),
			},
			{
				Config: testAccEnvironmentConfigWithReadme(orgName, envName, "# Updated"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("credible_environment.test", "readme", "# Updated"),
				),
			},
		},
	})
}

func testAccEnvironmentConfig(orgName, envName string) string {
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
`, orgName, envName)
}

func testAccEnvironmentConfigWithReadme(orgName, envName, readme string) string {
	return providerConfig() + fmt.Sprintf(`
resource "credible_organization" "test" {
  name                = %q
  deletion_protection = false
}

resource "credible_environment" "test" {
  organization        = credible_organization.test.name
  name                = %q
  readme              = %q
  deletion_protection = false
  force_cascade       = true
}
`, orgName, envName, readme)
}
