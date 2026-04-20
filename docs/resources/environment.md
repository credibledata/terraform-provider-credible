---
page_title: "credible_environment Resource - credible"
subcategory: ""
description: |-
  Manages an environment within a Credible organization.
---

# credible_environment (Resource)

Manages an environment within a Credible organization. Environments contain connections, packages, and package versions.

## Example Usage

### Create a new environment

```hcl
resource "credible_environment" "analytics" {
  name              = "analytics"
  readme            = "Analytics data models and dashboards"
  replication_count = 1
}
```

The `organization` attribute is optional — it defaults to the provider's `organization` value.

### Explicit organization

```hcl
resource "credible_environment" "analytics" {
  organization    = "my-org"
  name            = "analytics"
  readme          = "Analytics data models"
}
```

### Import an existing environment into Terraform

**Step 1:** Write the resource block:

```hcl
resource "credible_environment" "analytics" {
  name = "analytics"
}
```

**Step 2:** Import using the format `<organization>/<environment>`:

```shell
terraform import credible_environment.analytics my-org/analytics
```

**Step 3:** Run `terraform plan` and adjust your HCL until there are no diffs.

-> **Note:** After import, `deletion_protection` defaults to `true` and `force_cascade` defaults to `false`. These flags are Terraform-only and are not stored by the API.

## Schema

### Required

- `name` (String) — Unique environment name within the organization. **Immutable** — changing this forces destroy and recreate.

### Optional

- `organization` (String) — Organization name. **Default: provider's `organization`**. **Immutable** — changing forces destroy and recreate.
- `readme` (String) — Markdown-formatted environment description. Can be updated in place.
- `replication_count` (Number) — Number of replicas for high availability. Valid range: 1–10. Can be updated in place.
- `deletion_protection` (Boolean) — Prevents `terraform destroy`. **Default: `true`**.
- `force_cascade` (Boolean) — Allows deletion when the environment still contains packages or connections. **Default: `false`**. When `false`, Terraform blocks destroy if child resources exist.

### Read-Only

- `created_at` (String) — ISO 8601 creation timestamp.
- `updated_at` (String) — ISO 8601 last-update timestamp.

## Destroying an Environment

```hcl
resource "credible_environment" "analytics" {
  name                = "analytics"
  deletion_protection = false  # Step 1: disable protection
  force_cascade       = true   # Optional: skip child-resource check
}
```

```shell
terraform apply   # Step 2: apply flag changes
terraform destroy # Step 3: destroy the environment
```

If `force_cascade` is `false` and the environment contains packages or connections, Terraform will show an error telling you how many child resources exist.

## Import

Import format: `<organization>/<environment>`

```shell
terraform import credible_environment.analytics my-org/analytics
```
