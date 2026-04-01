---
page_title: "credible_group_member Resource - credible"
subcategory: ""
description: |-
  Manages membership of a user or group within a Credible group.
---

# credible_group_member (Resource)

Manages membership of a user or group within a Credible group. Each member has a status of either `admin` (can manage the group) or `member` (regular member).

## Example Usage

### Add a user as a group member

```hcl
resource "credible_group_member" "alice" {
  group_name    = "data-engineering"
  user_group_id = "user:alice@example.com"
  status        = "member"
}
```

### Add a user as a group admin

```hcl
resource "credible_group_member" "bob" {
  group_name    = credible_group.data_eng.name
  user_group_id = "user:bob@example.com"
  status        = "admin"
}
```

### Nest a group inside another group

```hcl
resource "credible_group_member" "nested_group" {
  group_name    = "platform-team"
  user_group_id = "group:data-engineering"
  status        = "member"
}
```

### Import an existing group member into Terraform

**Step 1:** Write the resource block:

```hcl
resource "credible_group_member" "alice" {
  group_name    = "data-engineering"
  user_group_id = "user:alice@example.com"
  status        = "member"
}
```

**Step 2:** Import using `<organization>/<group>/<user_group_id>`:

```shell
terraform import credible_group_member.alice my-org/data-engineering/user:alice@example.com
```

**Step 3:** Run `terraform plan` to verify. Adjust `status` if it doesn't match.

## Schema

### Required

- `group_name` (String) — The group to add the member to. **Immutable** — changing forces destroy and recreate.
- `user_group_id` (String) — The user or group to add. Format: `user:<email>` or `group:<group-name>`. **Immutable**.
- `status` (String) — Membership role. One of: `admin`, `member`. Can be updated in place.

### Optional

- `organization` (String) — Organization name. **Default: provider's `organization`**. **Immutable**.

### Member Status

| Status | Description |
|---|---|
| `admin` | Can manage the group (add/remove members, change settings) |
| `member` | Regular group member |

## Import

Import format: `<organization>/<group>/<user_group_id>`

```shell
terraform import credible_group_member.alice my-org/data-engineering/user:alice@example.com
```
