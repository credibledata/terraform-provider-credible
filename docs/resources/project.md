---
page_title: "credible_project Resource - credible"
subcategory: ""
description: |-
  Manages a project within a Credible organization.
---

# credible_project (Resource)

Manages a project within a Credible organization. Projects contain connections, packages, and package versions.

## Example Usage

### Create a new project

```hcl
resource "credible_project" "analytics" {
  name              = "analytics"
  readme            = "Analytics data models and dashboards"
  replication_count = 1
}
```

The `organization` attribute is optional — it defaults to the provider's `organization` value.

### Explicit organization

```hcl
resource "credible_project" "analytics" {
  organization    = "my-org"
  name            = "analytics"
  readme          = "Analytics data models"
}
```

### Import an existing project into Terraform

**Step 1:** Write the resource block:

```hcl
resource "credible_project" "analytics" {
  name = "analytics"
}
```

**Step 2:** Import using the format `<organization>/<project>`:

```shell
terraform import credible_project.analytics my-org/analytics
```

**Step 3:** Run `terraform plan` and adjust your HCL until there are no diffs.

-> **Note:** After import, `deletion_protection` defaults to `true` and `force_cascade` defaults to `false`. These flags are Terraform-only and are not stored by the API.

## Schema

### Required

- `name` (String) — Unique project name within the organization. **Immutable** — changing this forces destroy and recreate.

### Optional

- `organization` (String) — Organization name. **Default: provider's `organization`**. **Immutable** — changing forces destroy and recreate.
- `readme` (String) — Markdown-formatted project description. Can be updated in place.
- `replication_count` (Number) — Number of replicas for high availability. Valid range: 1–10. Can be updated in place.
- `deletion_protection` (Boolean) — Prevents `terraform destroy`. **Default: `true`**.
- `force_cascade` (Boolean) — Allows deletion when the project still contains packages or connections. **Default: `false`**. When `false`, Terraform blocks destroy if child resources exist.

### Read-Only

- `created_at` (String) — ISO 8601 creation timestamp.
- `updated_at` (String) — ISO 8601 last-update timestamp.

## Destroying a Project

```hcl
resource "credible_project" "analytics" {
  name                = "analytics"
  deletion_protection = false  # Step 1: disable protection
  force_cascade       = true   # Optional: skip child-resource check
}
```

```shell
terraform apply   # Step 2: apply flag changes
terraform destroy # Step 3: destroy the project
```

If `force_cascade` is `false` and the project contains packages or connections, Terraform will show an error telling you how many child resources exist.

## Import

Import format: `<organization>/<project>`

```shell
terraform import credible_project.analytics my-org/analytics
```
