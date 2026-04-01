---
page_title: "credible_project_permission Resource - credible"
subcategory: ""
description: |-
  Manages project-level permissions for a user or group in Credible.
---

# credible_project_permission (Resource)

Manages project-level permissions for a user or group. Each user or group can have exactly one permission level per project. Project permissions are separate from organization permissions.

## Example Usage

### Grant project viewer to a group

```hcl
resource "credible_project_permission" "data_team_viewer" {
  project       = "analytics"
  user_group_id = "group:data-engineering"
  permission    = "viewer"
}
```

### Grant project admin to a user

```hcl
resource "credible_project_permission" "alice_admin" {
  project       = credible_project.analytics.name
  user_group_id = "user:alice@example.com"
  permission    = "admin"
}
```

### Import an existing permission into Terraform

**Step 1:** Write the resource block:

```hcl
resource "credible_project_permission" "data_team_viewer" {
  project       = "analytics"
  user_group_id = "group:data-engineering"
  permission    = "viewer"
}
```

**Step 2:** Import using `<organization>/<project>/<user_group_id>`:

```shell
terraform import credible_project_permission.data_team_viewer my-org/analytics/group:data-engineering
```

**Step 3:** Run `terraform plan` to verify.

## Schema

### Required

- `project` (String) — Project name. **Immutable** — changing forces destroy and recreate.
- `user_group_id` (String) — The user or group. Format: `user:<email>` or `group:<group-name>`. **Immutable**.
- `permission` (String) — Permission level. One of: `admin`, `modeler`, `viewer`. Can be updated in place.

### Optional

- `organization` (String) — Organization name. **Default: provider's `organization`**. **Immutable**.

### Permission Levels

| Level | Description |
|---|---|
| `admin` | Full control over the project |
| `modeler` | Can create and edit models within the project |
| `viewer` | Read-only access to the project |

-> **Note:** Project permission levels (`admin`, `modeler`, `viewer`) differ from organization permission levels (`admin`, `modeler`, `member`).

## Import

Import format: `<organization>/<project>/<user_group_id>`

```shell
terraform import credible_project_permission.data_team_viewer my-org/analytics/group:data-engineering
```
