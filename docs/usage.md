# Credible Terraform Provider — Usage Guide

## Installation

### From Terraform Registry (recommended)

```hcl
terraform {
  required_providers {
    credible = {
      source  = "credibledata/credible"
      version = "~> 0.1"
    }
  }
}
```

### From Source

```bash
git clone https://github.com/credibledata/terraform-provider-credible.git
cd terraform-provider-credible
make install
```

## Provider Configuration

```hcl
provider "credible" {
  url          = "https://app.credibledata.com"
  organization = "my-org"
  api_key      = var.credible_api_key  # recommended for CI/CD
}
```

Or use environment variables:

```bash
export CREDIBLE_URL="https://app.credibledata.com"
export CREDIBLE_ORGANIZATION="my-org"
export CREDIBLE_API_KEY="your-api-key"
```

## Supported Resources

| Resource | Description |
|---|---|
| `credible_organization` | Manage organizations |
| `credible_environment` | Manage environments within an organization |
| `credible_connection` | Database connections (postgres, bigquery, snowflake, trino, mysql, duckdb, motherduck) |
| `credible_organization_permission` | Org-level user/group permissions (admin, modeler, member) |
| `credible_environment_permission` | Environment-level user/group permissions (admin, modeler, viewer) |
| `credible_group` | Manage groups within an organization |
| `credible_group_member` | Manage group membership |
| `credible_package` | Manage Malloy model packages |
| `credible_package_version` | Publish immutable package versions |

---

## Importing Existing Resources

If you already have resources created in Credible and want to bring them under Terraform management:

### Step 1: Write the resource block

```hcl
resource "credible_environment" "analytics" {
  name = "analytics"
}
```

### Step 2: Run `terraform import`

```bash
terraform import <resource_type>.<name> <import_id>
```

### Import ID Formats

| Resource | Import ID Format | Example |
|---|---|---|
| `credible_organization` | `<org>` | `terraform import credible_organization.main my-org` |
| `credible_environment` | `<org>/<environment>` | `terraform import credible_environment.analytics my-org/analytics` |
| `credible_connection` | `<org>/<environment>/<connection>` | `terraform import credible_connection.db my-org/analytics/main-db` |
| `credible_organization_permission` | `<org>/<user_group_id>` | `terraform import credible_organization_permission.alice my-org/user:alice@example.com` |
| `credible_environment_permission` | `<org>/<environment>/<user_group_id>` | `terraform import credible_environment_permission.bob my-org/analytics/user:bob@example.com` |
| `credible_group` | `<org>/<group>` | `terraform import credible_group.eng my-org/data-engineering` |
| `credible_group_member` | `<org>/<group>/<user_group_id>` | `terraform import credible_group_member.alice my-org/data-engineering/user:alice@example.com` |
| `credible_package` | `<org>/<environment>/<package>` | `terraform import credible_package.models my-org/analytics/analytics-models` |
| `credible_package_version` | N/A | Versions are immutable — recreate in Terraform instead |

### Step 3: Reconcile state

After importing, run `terraform plan` and update your HCL to match the imported state. Repeat until `terraform plan` shows no changes.

---

## Do's and Don'ts

### Do's

- **Use `api_key` auth for CI/CD** — service account keys are stable and don't expire like bearer tokens.
- **Set `deletion_protection = true`** (the default) for orgs, environments, and packages. Only disable when you intentionally want to destroy.
- **Use `terraform import`** to adopt existing resources before managing them with Terraform.
- **Pin the provider version** in `required_providers` to avoid unexpected breaking changes.
- **Store sensitive values (API keys, passwords) in a secret manager** — use `var` references, not hardcoded values.
- **Use `organization` at the provider level** to avoid repeating it in every resource block.
- **Run `terraform plan` before `terraform apply`** to review changes.

### Don'ts

- **Don't delete organizations via Terraform** — the API does not support org deletion. Destroying the resource only removes it from state.
- **Don't change `name` fields** on organizations, environments, connections, packages, or groups — these are immutable and will force resource recreation (destroy + create).
- **Don't change `type` on connections** — connection type is immutable. You must destroy and recreate.
- **Don't set `force_cascade = true` unless you understand the consequences** — it allows deleting orgs/environments that still contain child resources.
- **Don't expect `terraform destroy` to delete package versions** — versions are immutable. Destroy archives them instead.
- **Don't hardcode sensitive connection credentials** — use variables with `sensitive = true` or reference a secrets manager.
- **Don't rely on reading back sensitive fields** — the API does not return passwords, keys, or tokens after creation. Terraform preserves them from your configuration.

---

## Common Patterns

### Full Stack Setup

```hcl
# Organization
resource "credible_organization" "main" {
  name         = "acme-corp"
  display_name = "Acme Corporation"
}

# Environment
resource "credible_environment" "analytics" {
  organization = credible_organization.main.name
  name         = "analytics"
  readme       = "Analytics data models"
}

# Connection
resource "credible_connection" "warehouse" {
  organization = credible_organization.main.name
  environment  = credible_environment.analytics.name
  name         = "main-warehouse"
  type         = "bigquery"

  bigquery {
    default_project_id       = "my-gcp-project"
    service_account_key_json = var.bq_key
  }
}

# Package + Version
resource "credible_package" "models" {
  organization = credible_organization.main.name
  environment  = credible_environment.analytics.name
  name         = "analytics-models"
}

resource "credible_package_version" "v1" {
  organization = credible_organization.main.name
  environment  = credible_environment.analytics.name
  package_name = credible_package.models.name
  version_id   = "1.0.0"
  source_dir   = "${path.module}/models"
}
```

### Team Permissions

```hcl
# Create a group
resource "credible_group" "data_eng" {
  name        = "data-engineering"
  description = "Data engineering team"
}

# Add members
resource "credible_group_member" "alice" {
  group_name    = credible_group.data_eng.name
  user_group_id = "user:alice@example.com"
  status        = "admin"
}

resource "credible_group_member" "bob" {
  group_name    = credible_group.data_eng.name
  user_group_id = "user:bob@example.com"
  status        = "member"
}

# Grant environment access to the group
resource "credible_environment_permission" "data_eng_access" {
  environment   = credible_environment.analytics.name
  user_group_id = "group:data-engineering"
  permission    = "modeler"
}
```

### Disabling Deletion Protection for Teardown

```hcl
# Step 1: Set deletion_protection = false and apply
resource "credible_environment" "analytics" {
  name                = "analytics"
  deletion_protection = false
  force_cascade       = true  # also delete child resources
}

# Step 2: Then run terraform destroy
```

---

## Troubleshooting

| Symptom | Cause | Fix |
|---|---|---|
| `Error: deletion protection is enabled` | Trying to destroy a protected resource | Set `deletion_protection = false`, apply, then destroy |
| `Error: Organization is not empty` (environments) | Trying to destroy org with children | Set `force_cascade = true` or destroy environments first |
| `Error: HTTP 404` after import | Import ID format is wrong | Check the import ID format table above |
| Plan shows changes to sensitive fields | API doesn't return secrets | This is expected — Terraform preserves values from config |
| `name` change shows destroy+create | Name fields are immutable | This is by design — rename requires recreation |
