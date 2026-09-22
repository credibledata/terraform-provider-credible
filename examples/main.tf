terraform {
  required_providers {
    credible = {
      source = "registry.terraform.io/credibledata/credible"
    }
  }
}

variable "credible_api_key" {
  description = "API key for the Credible API"
  type        = string
  sensitive   = true
}

variable "bq_service_account_key" {
  description = "BigQuery service account key JSON"
  type        = string
  sensitive   = true
  default     = ""
}

# Provider configuration
provider "credible" {
  url          = "https://credible.example.com"
  organization = "my-org"
  api_key      = var.credible_api_key
}

# 1) Create an organization
resource "credible_organization" "main" {
  name         = "my-org"
  display_name = "My Organization"
}

# 2) Create an environment
resource "credible_environment" "analytics" {
  name             = "analytics"
  readme           = "Analytics data models and dashboards"
  replication_count = 1
}

# 3) Create a database connection (BigQuery example)
resource "credible_connection" "warehouse" {
  environment = credible_environment.analytics.name
  name    = "main-warehouse"
  type    = "bigquery"

  bigquery {
    default_project_id       = "my-gcp-project"
    service_account_key_json = var.bq_service_account_key
  }

  include_tables = ["analytics.*", "sales.*"]
}

# 3b) PostgreSQL connection example
resource "credible_connection" "postgres_db" {
  environment = credible_environment.analytics.name
  name    = "app-database"
  type    = "postgres"

  postgres {
    host          = "db.example.com"
    port          = 5432
    database_name = "appdb"
    user_name     = "readonly_user"
    password      = "secret" # Use a secret manager in production
  }

  include_tables = ["public.*"]
}

# 4) Assign users/groups with permissions

# Organization-level admin
resource "credible_organization_permission" "alice_admin" {
  user_group_id = "user:alice@example.com"
  permission    = "admin"
}

# Organization-level modeler
resource "credible_organization_permission" "bob_modeler" {
  user_group_id = "user:bob@example.com"
  permission    = "modeler"
}

# Environment-level viewer
resource "credible_environment_permission" "carol_viewer" {
  environment   = credible_environment.analytics.name
  user_group_id = "user:carol@example.com"
  permission    = "viewer"
}

# Environment-level admin for a group
resource "credible_environment_permission" "data_team_admin" {
  environment   = credible_environment.analytics.name
  user_group_id = "group:data-engineering"
  permission    = "admin"
}

# 5) Publish a package, and manage an existing one
#
# Publishing is what creates a package: one multipart POST carries the package, the
# version and the model archive together. So credible_package_version stands alone
# below -- no credible_package resource is needed to create it. Use credible_package
# to manage the metadata of a package that already exists (import it if it was
# published by `cred publish` or the Admin API); it cannot create one.
# See docs/resources/package.md and docs/resources/package_version.md.

# Option A: Publish from a local directory (the provider zips it)
resource "credible_package_version" "v1" {
  environment  = credible_environment.analytics.name
  package_name = "analytics-models"
  version_id   = "1.0.0"
  description  = "Core analytics Malloy models"
  source_dir   = "${path.module}/models/analytics"
}

# Option B: Publish from a pre-built zip. The API reads the archive's bytes, so it
# must be a zip -- a .tar.gz is rejected whatever the file is named.
# resource "credible_package_version" "v1_archive" {
#   environment  = credible_environment.analytics.name
#   package_name = "analytics-models"
#   version_id   = "1.0.0"
#   source_file  = "${path.module}/dist/analytics-models.zip"
#   source_hash  = filemd5("${path.module}/dist/analytics-models.zip")
# }
