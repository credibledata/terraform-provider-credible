---
page_title: "credible_organization Resource - credible"
subcategory: ""
description: |-
  Manages a Credible organization.
---

# credible_organization (Resource)

Manages a Credible organization.

~> **Important:** The Credible API does not support organization deletion. Running `terraform destroy` on this resource only removes it from Terraform state — the organization continues to exist on the platform.

## Example Usage

### Create a new organization

```hcl
resource "credible_organization" "main" {
  name         = "my-org"
  display_name = "My Organization"
}
```

### Import an existing organization into Terraform

If you already have an organization on the Credible platform and want to manage it with Terraform:

**Step 1:** Write the resource block matching its current state:

```hcl
resource "credible_organization" "main" {
  name         = "my-org"
  display_name = "My Organization"
}
```

**Step 2:** Import it:

```shell
terraform import credible_organization.main my-org
```

**Step 3:** Run `terraform plan` to verify there are no diffs. Adjust your HCL if needed until the plan is clean.

-> **Note:** After import, `deletion_protection` and `force_cascade` will not be set in state (the API does not store them). Terraform will apply their defaults (`true` and `false` respectively) on the next `terraform apply`.

## Schema

### Required

- `name` (String) — Unique organization name. **Immutable** — changing this forces the resource to be destroyed and recreated.

### Optional

- `display_name` (String) — Human-readable display name. Can be updated in place.
- `deletion_protection` (Boolean) — Prevents `terraform destroy` from removing this resource. **Default: `true`**. You must explicitly set this to `false` and run `terraform apply` before you can destroy the organization.
- `force_cascade` (Boolean) — Controls whether the organization can be destroyed when it still contains projects. **Default: `false`**. When `false`, Terraform will refuse to destroy an organization that has projects — you must remove the projects first or set this to `true`.

### Read-Only

- `created_at` (String) — ISO 8601 creation timestamp.
- `updated_at` (String) — ISO 8601 last-update timestamp.

## Destroying an Organization

Because `deletion_protection` defaults to `true`, you cannot destroy this resource without first disabling it:

```hcl
resource "credible_organization" "main" {
  name                = "my-org"
  deletion_protection = false  # Step 1: disable protection
  force_cascade       = true   # Optional: allow deletion with child projects
}
```

```shell
terraform apply   # Step 2: apply the flag change
terraform destroy # Step 3: now destroy works (removes from state only)
```

## Import

The import ID is the organization name.

```shell
terraform import credible_organization.main my-org
```
