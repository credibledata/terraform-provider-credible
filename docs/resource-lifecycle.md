---
page_title: "Resource lifecycle semantics - credible"
subcategory: ""
description: |-
  What create, update, destroy and import actually do against the Credible API for each resource, including where Terraform's model and the server diverge.
---

# Resource lifecycle semantics

Terraform assumes a resource it creates it can also destroy, and that removing a
resource from configuration removes it from the world. The Credible API does not
grant that everywhere: some objects are owned by Credible and cannot be created or
deleted through the API at all, and some deletes are archives rather than removals.

This page states, per resource, what each lifecycle operation really does. Read it
before writing a `terraform destroy` you expect to reclaim something.

## Summary

| Resource | Create | `terraform destroy` | Import |
|---|---|---|---|
| `credible_organization` | Not usable in practice -- see below | **Removes from state only.** The organization keeps existing | Yes -- `<org>` |
| `credible_environment` | Creates | Deletes the environment | Yes -- `<org>/<env>` |
| `credible_connection` | Creates | Deletes the connection | Yes -- `<org>/<env>/<name>` |
| `credible_package` | Creates | Deletes the package and all its versions | Yes -- `<org>/<env>/<pkg>` |
| `credible_package_version` | Publishes | **Archives, does not delete.** The version and its artifact survive | No -- not supported |
| `credible_group` | Creates | Deletes the group | Yes -- `<org>/<group>` |
| `credible_group_member` | Adds a member | Removes the member from the group | Yes -- `<org>/<group>/<user_group_id>` |
| `credible_organization_permission` | Grants | Revokes | Yes -- `<org>/<user_group_id>` |
| `credible_environment_permission` | Grants | Revokes | Yes -- `<org>/<env>/<user_group_id>` |

## Organizations are owned by Credible

`credible_organization` is effectively an **import-and-read** resource.

Creating and deleting an organization is restricted to Credible operators, so a
normal practitioner token cannot provision one. `terraform destroy` reflects that:
it performs **no API call at all**. It drops the resource from Terraform state and
logs a warning; the organization, its environments, packages and data are all still
there afterwards.

Two consequences worth internalizing:

- **`terraform destroy` is not a cleanup mechanism here.** If you need the
  organization gone, ask Credible.
- **Renaming is destructive-looking but non-destructive.** `name` forces
  replacement, so changing it plans a destroy and a create. The destroy is a no-op,
  so you end up with the original organization still live plus an attempted new one.
  Do not rename; import the correct organization instead.

The `deletion_protection` and `force_cascade` attributes gate that no-op. They are
provider-side guards, not API features.

Use it like this: import the organization you were given, manage `display_name`, and
build everything else underneath it.

## Package versions are archived, not deleted

`credible_package_version` is the other place the word "destroy" is misleading.

The API has no delete for a version -- versions are immutable once published. On
destroy the provider issues `PATCH archiveStatus: "archive"`. The version row, the
uploaded artifact and its index all remain on the server; only the archive flag
flips. Terraform then drops it from state.

Because the source attributes force replacement, changing package contents runs that
archive and then re-publishes **the same `version_id`**. If the API refuses to
re-publish an id that already exists, the destroy half succeeds and the create half
fails, leaving you with no state and an archived version. Publish a new `version_id`
for new content rather than editing a published one in place.

`credible_package_version` also does not support `terraform import`.

### Detecting content changes

`source_hash` is the only mechanism that makes Terraform notice that the files under
`source_dir` changed. It is optional, and if you omit it, editing your model files
produces **no diff at all** -- Terraform reports "no changes" indefinitely. Set it:

```hcl
source_hash = sha256(join("", [for f in fileset(path.module, "models/**") : filesha256("${path.module}/${f}")]))
```

## Deleting a parent orphans its children in state

Deletes cascade on the server, but Terraform is not told. These pairs matter:

- Deleting a `credible_environment` with `force_cascade = true` deletes the
  packages and connections inside it. Their Terraform resources stay in state and
  disappear only on a later refresh.
- Deleting a `credible_package` deletes its versions. Any
  `credible_package_version` resources become stale state entries.
- Deleting a `credible_group` does not remove the `credible_group_member` resources
  pointing at it.

Terraform's dependency graph destroys children before parents when it knows about
the dependency, so declare the relationship (`depends_on`, or reference the parent's
attributes) rather than relying on cascade.

## Fields you cannot clear

`readme` on an environment, and `description` on a group or package, cannot be set
back to empty. The API omits empty strings from the request, so the server keeps the
previous value -- and the provider then writes that retained value back into state,
so the failed clear does not even show up as drift on the next plan.

To actually clear one of these, destroy and recreate the resource.

## Imported resources are not deletion-protected

`credible_organization`, `credible_environment` and `credible_package` declare
`deletion_protection` with a default of `true`, but import does not populate it. An
imported resource therefore reads as unprotected and is destroyable on the first
attempt, and your next plan shows a diff for the attribute.

Set `deletion_protection` explicitly in configuration after importing.

## Permission vocabularies differ by level

The valid values are not the same at each level, which is easy to get wrong:

| Resource | Accepted values |
|---|---|
| `credible_organization_permission` | `admin`, `modeler`, `member` |
| `credible_environment_permission` | `admin`, `modeler`, `viewer` |
| `credible_group_member` (`status`) | `admin`, `member` |

## Secrets are never returned

Passwords, tokens, keys and service-account JSON are write-only: reads never return
them. The provider carries the configured value forward so it does not show as
drift.

The practical effect is on **import**: an imported connection has no credentials in
state, so you must supply them in configuration before the first apply. For the SSH
proxy specifically, the API treats a blank `private_key` on update as "keep the
stored key" rather than "clear it".

## Drift the provider does not detect

`credible_group_member` reads by listing the group's members and matching. If the
group itself is deleted out of band and the API returns an empty list rather than a
404, the member is silently removed from state.

Connection credentials cannot be drift-checked at all, since the API does not return
them -- Terraform will not notice a password changed in the Credible UI.
