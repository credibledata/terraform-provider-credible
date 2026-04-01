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

# 2) Create a project
resource "credible_project" "analytics" {
  name             = "analytics"
  readme           = "Analytics data models and dashboards"
  replication_count = 1
}

# 3) Create a database connection (BigQuery example)
resource "credible_connection" "warehouse" {
  project = credible_project.analytics.name
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
  project = credible_project.analytics.name
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

# Project-level viewer
resource "credible_project_permission" "carol_viewer" {
  project       = credible_project.analytics.name
  user_group_id = "user:carol@example.com"
  permission    = "viewer"
}

# Project-level admin for a group
resource "credible_project_permission" "data_team_admin" {
  project       = credible_project.analytics.name
  user_group_id = "group:data-engineering"
  permission    = "admin"
}

# 5) Create a package and publish a version

resource "credible_package" "models" {
  project     = credible_project.analytics.name
  name        = "analytics-models"
  description = "Core analytics Malloy models"
}

# Option A: Publish from a local directory (provider creates the .tar.gz)
resource "credible_package_version" "v1" {
  project      = credible_project.analytics.name
  package_name = credible_package.models.name
  version_id   = "1.0.0"
  source_dir   = "${path.module}/models/analytics"
}

# Option B: Publish from a pre-built archive
# resource "credible_package_version" "v1_archive" {
#   project      = credible_project.analytics.name
#   package_name = credible_package.models.name
#   version_id   = "1.0.0"
#   source_file  = "${path.module}/dist/analytics-models.tar.gz"
#   source_hash  = filemd5("${path.module}/dist/analytics-models.tar.gz")
# }
