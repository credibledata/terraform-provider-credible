package resources_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// createTestSourceDir writes a self-contained package: a publisher.json manifest
// at the archive root plus a model over an inline CSV. The publish compiles the
// package server-side, so a model referencing a table that is not in the archive
// fails the apply with a 424 rather than a useful message.
func createTestSourceDir(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "sample_package")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"publisher.json": `{"name":"acc-test-pkg","version":"1.0.0"}`,
		"rows.csv":       "id,amount\n1,10\n2,20\n",
		"model.malloy": "source: rows is duckdb.table('rows.csv') extend {\n" +
			"  measure: row_count is count()\n}\n",
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// Publishing is what creates a package, so this needs no credible_package
// resource: the version resource creates the package on its own. It also needs an
// existing organization and environment, because organization creation is a system
// privilege no API credential holds.
//
// CREDIBLE_ORGANIZATION and CREDIBLE_ENVIRONMENT name where to publish; the package
// name is randomized per run so a re-run does not collide with the previous one
// (re-publishing an existing version_id is a 409).
func TestAccPackageVersion_publishCreatesThePackage(t *testing.T) {
	orgName := os.Getenv("CREDIBLE_ORGANIZATION")
	envName := os.Getenv("CREDIBLE_ENVIRONMENT")
	if orgName == "" || envName == "" {
		t.Skip("CREDIBLE_ORGANIZATION and CREDIBLE_ENVIRONMENT must name an existing organization " +
			"and environment to publish into. Each run publishes a new randomly named package there.")
	}
	pkgName := randomName("test-pkg-tf")
	sourceDir := createTestSourceDir(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccPackageVersionConfig(orgName, envName, pkgName, "1.0.0", "", sourceDir),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("credible_package_version.test", "version_id", "1.0.0"),
					resource.TestCheckResourceAttr("credible_package_version.test", "organization", orgName),
					resource.TestCheckResourceAttr("credible_package_version.test", "environment", envName),
					resource.TestCheckResourceAttr("credible_package_version.test", "package_name", pkgName),
					resource.TestCheckResourceAttrSet("credible_package_version.test", "archive_status"),
					resource.TestCheckResourceAttrSet("credible_package_version.test", "build_status"),
					resource.TestCheckResourceAttrSet("credible_package_version.test", "created_at"),
					resource.TestCheckResourceAttrSet("credible_package_version.test", "updated_at"),
				),
			},
		},
	})
}

// A second version published while the first is still the promoted one. The
// publish response's latestVersion names the *promoted* version, so it still reads
// 1.0.0 here -- state must carry the version that was published, not that one.
func TestAccPackageVersion_secondVersionTakesItsOwnID(t *testing.T) {
	orgName := os.Getenv("CREDIBLE_ORGANIZATION")
	envName := os.Getenv("CREDIBLE_ENVIRONMENT")
	if orgName == "" || envName == "" {
		t.Skip("CREDIBLE_ORGANIZATION and CREDIBLE_ENVIRONMENT must name an existing organization and environment.")
	}
	pkgName := randomName("test-pkg-tf")
	sourceDir := createTestSourceDir(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccPackageVersionConfig(orgName, envName, pkgName, "1.0.0", "", sourceDir),
				Check: resource.TestCheckResourceAttr(
					"credible_package_version.test", "version_id", "1.0.0"),
			},
			{
				Config: testAccPackageVersionConfig(orgName, envName, pkgName, "2.0.0", "", sourceDir),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("credible_package_version.test", "version_id", "2.0.0"),
					resource.TestCheckResourceAttrSet("credible_package_version.test", "build_status"),
				),
			},
		},
	})
}

func TestAccPackageVersion_archive(t *testing.T) {
	orgName := os.Getenv("CREDIBLE_ORGANIZATION")
	envName := os.Getenv("CREDIBLE_ENVIRONMENT")
	if orgName == "" || envName == "" {
		t.Skip("CREDIBLE_ORGANIZATION and CREDIBLE_ENVIRONMENT must name an existing organization and environment.")
	}
	pkgName := randomName("test-pkg-tf")
	sourceDir := createTestSourceDir(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccPackageVersionConfig(orgName, envName, pkgName, "1.0.0", "", sourceDir),
				Check: resource.TestCheckResourceAttr(
					"credible_package_version.test", "version_id", "1.0.0"),
			},
			{
				Config: testAccPackageVersionConfig(orgName, envName, pkgName, "1.0.0", "archive", sourceDir),
				Check: resource.TestCheckResourceAttr(
					"credible_package_version.test", "archive_status", "archive"),
			},
		},
	})
}

// archiveStatus is omitted entirely when empty so the attribute stays unset rather
// than being sent as "".
func testAccPackageVersionConfig(orgName, envName, pkgName, versionID, archiveStatus, sourceDir string) string {
	archiveLine := ""
	if archiveStatus != "" {
		archiveLine = fmt.Sprintf("\n  archive_status = %q", archiveStatus)
	}
	return providerConfig() + fmt.Sprintf(`
resource "credible_package_version" "test" {
  organization = %q
  environment  = %q
  package_name = %q
  version_id   = %q
  source_dir   = %q%s
}
`, orgName, envName, pkgName, versionID, sourceDir, archiveLine)
}
