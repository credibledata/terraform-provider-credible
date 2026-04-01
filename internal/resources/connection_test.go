package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccConnection_postgres(t *testing.T) {
	orgName := randomName("test-org-tf")
	projName := randomName("test-proj-tf")
	connName := randomName("test-conn-tf")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionPostgresConfig(orgName, projName, connName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("credible_connection.test", "name", connName),
					resource.TestCheckResourceAttr("credible_connection.test", "type", "postgres"),
					resource.TestCheckResourceAttr("credible_connection.test", "organization", orgName),
					resource.TestCheckResourceAttr("credible_connection.test", "project", projName),
					resource.TestCheckResourceAttrSet("credible_connection.test", "indexing_status"),
				),
			},
			// Import
			{
				ResourceName:            "credible_connection.test",
				ImportState:             true,
				ImportStateId:           fmt.Sprintf("%s/%s/%s", orgName, projName, connName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"postgres.password", "postgres.connection_string"},
			},
		},
	})
}

func TestAccConnection_duckdb(t *testing.T) {
	orgName := randomName("test-org-tf")
	projName := randomName("test-proj-tf")
	connName := randomName("test-conn-tf")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionDuckdbConfig(orgName, projName, connName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("credible_connection.test", "name", connName),
					resource.TestCheckResourceAttr("credible_connection.test", "type", "duckdb"),
				),
			},
		},
	})
}

func testAccConnectionPostgresConfig(orgName, projName, connName string) string {
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

resource "credible_connection" "test" {
  organization = credible_organization.test.name
  project      = credible_project.test.name
  name         = %q
  type         = "postgres"

  postgres {
    host          = "localhost"
    port          = 5432
    database_name = "testdb"
    user_name     = "testuser"
    password      = "testpass"
  }
}
`, orgName, projName, connName)
}

func testAccConnectionDuckdbConfig(orgName, projName, connName string) string {
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

resource "credible_connection" "test" {
  organization = credible_organization.test.name
  project      = credible_project.test.name
  name         = %q
  type         = "duckdb"

  duckdb {
    url = "duckdb:///test.db"
  }
}
`, orgName, projName, connName)
}
