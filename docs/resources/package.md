---
page_title: "credible_package Resource - credible"
subcategory: ""
description: |-
  Manages a Malloy model package within a Credible environment.
---

# credible_package (Resource)

Manages a Malloy model package within a Credible environment. Packages contain versioned collections of Malloy models that can be published and shared.

## Example Usage

### Create a new package

```hcl
resource "credible_package" "models" {
  environment = credible_environment.analytics.name
  name        = "analytics-models"
  description = "Core analytics Malloy models"
}
```

### Create a package and publish a version

```hcl
resource "credible_package" "models" {
  environment = "analytics"
  name        = "analytics-models"
  description = "Core analytics Malloy models"
}

resource "credible_package_version" "v1" {
  environment  = "analytics"
  package_name = credible_package.models.name
  version_id   = "1.0.0"
  source_dir   = "${path.module}/models"
}
```

### Import an existing package into Terraform

**Step 1:** Write the resource block:

```hcl
resource "credible_package" "models" {
  environment = "analytics"
  name        = "analytics-models"
  description = "Core analytics Malloy models"
}
```

**Step 2:** Import using `<organization>/<environment>/<package>`:

```shell
terraform import credible_package.models my-org/analytics/analytics-models
```

**Step 3:** Run `terraform plan` to verify.

-> **Note:** After import, `deletion_protection` defaults to `true`. You do not need to set it unless you want to destroy the package.

## Schema

### Required

- `environment` (String) — Environment name. **Immutable** — changing forces destroy and recreate.
- `name` (String) — Unique package name within the environment. **Immutable**.

### Optional

- `organization` (String) — Organization name. **Default: provider's `organization`**. **Immutable**.
- `description` (String) — Package description. Can be updated in place.
- `deletion_protection` (Boolean) — Prevents `terraform destroy`. **Default: `true`**. Must be set to `false` before the package can be destroyed.

### Read-Only

- `latest_version` (String) — The most recently published version identifier.
- `created_at` (String) — ISO 8601 creation timestamp.
- `updated_at` (String) — ISO 8601 last-update timestamp.

## Destroying a Package

```hcl
resource "credible_package" "models" {
  environment         = "analytics"
  name                = "analytics-models"
  deletion_protection = false  # Step 1: disable protection
}
```

```shell
terraform apply   # Step 2: apply the flag change
terraform destroy # Step 3: destroy the package
```

## Import

Import format: `<organization>/<environment>/<package>`

```shell
terraform import credible_package.models my-org/analytics/analytics-models
```
