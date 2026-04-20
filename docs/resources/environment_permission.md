---
page_title: "credible_environment_permission Resource - credible"
subcategory: ""
description: |-
  Manages environment-level permissions for a user or group in Credible.
---

# credible_environment_permission (Resource)

Manages environment-level permissions for a user or group. Each user or group can have exactly one permission level per environment. Environment permissions are separate from organization permissions.

## Example Usage

### Grant environment viewer to a group

```hcl
resource "credible_environment_permission" "data_team_viewer" {
  environment   = "analytics"
  user_group_id = "group:data-engineering"
  permission    = "viewer"
}
```

### Grant environment admin to a user

```hcl
resource "credible_environment_permission" "alice_admin" {
  environment   = credible_environment.analytics.name
  user_group_id = "user:alice@example.com"
  permission    = "admin"
}
```

### Import an existing permission into Terraform

**Step 1:** Write the resource block:

```hcl
resource "credible_environment_permission" "data_team_viewer" {
  environment   = "analytics"
  user_group_id = "group:data-engineering"
  permission    = "viewer"
}
```

**Step 2:** Import using `<organization>/<environment>/<user_group_id>`:

```shell
terraform import credible_environment_permission.data_team_viewer my-org/analytics/group:data-engineering
```

**Step 3:** Run `terraform plan` to verify.

## Schema

### Required

- `environment` (String) — Environment name. **Immutable** — changing forces destroy and recreate.
- `user_group_id` (String) — The user or group. Format: `user:<email>` or `group:<group-name>`. **Immutable**.
- `permission` (String) — Permission level. One of: `admin`, `modeler`, `viewer`. Can be updated in place.

### Optional

- `organization` (String) — Organization name. **Default: provider's `organization`**. **Immutable**.

### Permission Levels

| Level | Description |
|---|---|
| `admin` | Full control over the environment |
| `modeler` | Can create and edit models within the environment |
| `viewer` | Read-only access to the environment |

-> **Note:** Environment permission levels (`admin`, `modeler`, `viewer`) differ from organization permission levels (`admin`, `modeler`, `member`).

## Import

Import format: `<organization>/<environment>/<user_group_id>`

```shell
terraform import credible_environment_permission.data_team_viewer my-org/analytics/group:data-engineering
```
