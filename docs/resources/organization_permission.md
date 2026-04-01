---
page_title: "credible_organization_permission Resource - credible"
subcategory: ""
description: |-
  Manages organization-level permissions for a user or group in Credible.
---

# credible_organization_permission (Resource)

Manages organization-level permissions for a user or group. Each user or group can have exactly one permission level per organization.

## Example Usage

### Grant org admin to a user

```hcl
resource "credible_organization_permission" "alice_admin" {
  user_group_id = "user:alice@example.com"
  permission    = "admin"
}
```

### Grant org modeler to a group

```hcl
resource "credible_organization_permission" "eng_modeler" {
  user_group_id = "group:data-engineering"
  permission    = "modeler"
}
```

### Import an existing permission into Terraform

**Step 1:** Write the resource block:

```hcl
resource "credible_organization_permission" "alice_admin" {
  user_group_id = "user:alice@example.com"
  permission    = "admin"
}
```

**Step 2:** Import using `<organization>/<user_group_id>`:

```shell
terraform import credible_organization_permission.alice_admin my-org/user:alice@example.com
```

**Step 3:** Run `terraform plan` to verify. Adjust `permission` in your HCL if it doesn't match.

## Schema

### Required

- `user_group_id` (String) — The user or group to grant the permission to. Format: `user:<email>` or `group:<group-name>`. **Immutable** — changing forces destroy and recreate.
- `permission` (String) — Permission level. One of: `admin`, `modeler`, `member`. Can be updated in place.

### Optional

- `organization` (String) — Organization name. **Default: provider's `organization`**. **Immutable**.

### Permission Levels

| Level | Description |
|---|---|
| `admin` | Full control over the organization, projects, and settings |
| `modeler` | Can create and edit models within the organization's projects |
| `member` | Read-only access to the organization |

## Import

Import format: `<organization>/<user_group_id>`

```shell
# Import a user permission
terraform import credible_organization_permission.alice_admin my-org/user:alice@example.com

# Import a group permission
terraform import credible_organization_permission.eng_modeler my-org/group:data-engineering
```
