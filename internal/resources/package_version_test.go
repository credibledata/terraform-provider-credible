package resources_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// createTestSourceDir creates a temporary directory with a sample Malloy file for package version tests.
func createTestSourceDir(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "sample_package")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "model.malloy"), []byte("source: test_source is duckdb.table('test.csv') extend {\n  measure: row_count is count()\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestAccPackageVersion_basic(t *testing.T) {
	orgName := randomName("test-org-tf")
	projName := randomName("test-proj-tf")
	pkgName := randomName("test-pkg-tf")
	sourceDir := createTestSourceDir(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccPackageVersionConfig(orgName, projName, pkgName, "1.0.0", sourceDir),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("credible_package_version.test", "version_id", "1.0.0"),
					resource.TestCheckResourceAttr("credible_package_version.test", "organization", orgName),
					resource.TestCheckResourceAttr("credible_package_version.test", "project", projName),
					resource.TestCheckResourceAttr("credible_package_version.test", "package_name", pkgName),
					resource.TestCheckResourceAttrSet("credible_package_version.test", "archive_status"),
					resource.TestCheckResourceAttrSet("credible_package_version.test", "index_status"),
					resource.TestCheckResourceAttrSet("credible_package_version.test", "created_at"),
					resource.TestCheckResourceAttrSet("credible_package_version.test", "updated_at"),
				),
			},
		},
	})
}

func TestAccPackageVersion_archive(t *testing.T) {
	orgName := randomName("test-org-tf")
	projName := randomName("test-proj-tf")
	pkgName := randomName("test-pkg-tf")
	sourceDir := createTestSourceDir(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccPackageVersionConfig(orgName, projName, pkgName, "1.0.0", sourceDir),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("credible_package_version.test", "version_id", "1.0.0"),
				),
			},
			{
				Config: testAccPackageVersionConfigWithArchive(orgName, projName, pkgName, "1.0.0", "archive", sourceDir),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("credible_package_version.test", "archive_status", "archive"),
				),
			},
		},
	})
}

func testAccPackageVersionConfig(orgName, projName, pkgName, versionID, sourceDir string) string {
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
  deletion_protection = false
}

resource "credible_package_version" "test" {
  organization = credible_organization.test.name
  project      = credible_project.test.name
  package_name = credible_package.test.name
  version_id   = %q
  source_dir   = %q
}
`, orgName, projName, pkgName, versionID, sourceDir)
}

func testAccPackageVersionConfigWithArchive(orgName, projName, pkgName, versionID, archiveStatus, sourceDir string) string {
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
  deletion_protection = false
}

resource "credible_package_version" "test" {
  organization   = credible_organization.test.name
  project        = credible_project.test.name
  package_name   = credible_package.test.name
  version_id     = %q
  source_dir     = %q
  archive_status = %q
}
`, orgName, projName, pkgName, versionID, sourceDir, archiveStatus)
}
