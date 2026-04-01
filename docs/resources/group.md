---
page_title: "credible_group Resource - credible"
subcategory: ""
description: |-
  Manages a group within a Credible organization.
---

# credible_group (Resource)

Manages a group within a Credible organization. Groups let you assign permissions to multiple users at once by granting the group a permission, then adding users to the group.

## Example Usage

### Create a new group

```hcl
resource "credible_group" "data_eng" {
  name        = "data-engineering"
  description = "Data engineering team"
}
```

### Use a group with permissions

```hcl
resource "credible_group" "data_eng" {
  name        = "data-engineering"
  description = "Data engineering team"
}

resource "credible_group_member" "alice" {
  group_name    = credible_group.data_eng.name
  user_group_id = "user:alice@example.com"
  status        = "admin"
}

resource "credible_project_permission" "data_eng_access" {
  project       = "analytics"
  user_group_id = "group:data-engineering"
  permission    = "modeler"
}
```

### Import an existing group into Terraform

**Step 1:** Write the resource block:

```hcl
resource "credible_group" "data_eng" {
  name        = "data-engineering"
  description = "Data engineering team"
}
```

**Step 2:** Import using `<organization>/<group>`:

```shell
terraform import credible_group.data_eng my-org/data-engineering
```

**Step 3:** Run `terraform plan` to verify. Adjust `description` if it doesn't match.

## Schema

### Required

- `name` (String) — Unique group name within the organization. **Immutable** — changing forces destroy and recreate.

### Optional

- `organization` (String) — Organization name. **Default: provider's `organization`**. **Immutable**.
- `description` (String) — Human-readable group description. Can be updated in place.

### Read-Only

- `created_at` (String) — ISO 8601 creation timestamp.
- `updated_at` (String) — ISO 8601 last-update timestamp.

## Import

Import format: `<organization>/<group>`

```shell
terraform import credible_group.data_eng my-org/data-engineering
```
