---
page_title: "Credible Provider"
subcategory: ""
description: |-
  The Credible provider manages organizations, projects, database connections, permissions, groups, and packages on the Credible platform.
---

# Credible Provider

The Credible provider allows you to manage [Credible](https://credibledata.com) resources declaratively with Terraform — organizations, projects, database connections, permissions, groups, and Malloy model packages.

## Example Usage

```hcl
terraform {
  required_providers {
    credible = {
      source = "credibledata/credible"
    }
  }
}

provider "credible" {
  url          = "https://app.credibledata.com"
  organization = "my-org"
  api_key      = var.credible_api_key
}
```

## Authentication

The provider supports three authentication methods, checked in this order:

1. **API Key** (`api_key`) — A service account API key. **Recommended for CI/CD and automation.**
2. **Bearer Token** (`bearer_token`) — An OAuth2/Auth0 access token. Useful for short-lived sessions.
3. **CLI Config Fallback** — If neither is set, the provider reads credentials from the Credible CLI config file (`~/.cred`) for the active environment.

### Environment Variables

All provider attributes can be set via environment variables. Environment variables are used as fallbacks when the attribute is not set in the HCL configuration.

| Attribute      | Environment Variable       |
|----------------|----------------------------|
| `url`          | `CREDIBLE_URL`             |
| `organization` | `CREDIBLE_ORGANIZATION`    |
| `api_key`      | `CREDIBLE_API_KEY`         |
| `bearer_token` | `CREDIBLE_BEARER_TOKEN`    |

## Schema

### Optional

- `url` (String) — Credible API URL. Falls back to `CREDIBLE_URL` environment variable.
- `organization` (String) — Default organization name. Used by all resources that have an `organization` attribute when it is not explicitly set on the resource. Falls back to `CREDIBLE_ORGANIZATION` environment variable.
- `api_key` (String, Sensitive) — Service account API key. Falls back to `CREDIBLE_API_KEY` environment variable.
- `bearer_token` (String, Sensitive) — OAuth2/Auth0 access token. Falls back to `CREDIBLE_BEARER_TOKEN` environment variable.

## Default Behaviors

- **`organization`** — If set on the provider, all resources inherit it. You can override it per-resource.
- **`deletion_protection`** — Defaults to `true` on organizations, projects, and packages. You must explicitly set it to `false` before `terraform destroy` will work.
- **`force_cascade`** — Defaults to `false` on organizations and projects. When `false`, Terraform refuses to delete a resource that still contains children (e.g., an org with projects, a project with packages/connections).
- **Immutable fields** — `name`, `type`, `version_id`, and `user_group_id` cannot be changed after creation. Changing them in your config will trigger a destroy-and-recreate.
- **Sensitive fields** — Passwords, keys, and tokens are never returned by the API after creation. Terraform preserves them from your configuration. Plan diffs for these fields are expected.

## Supported Resources

| Resource | Description |
|---|---|
| [credible_organization](resources/organization.md) | Organizations |
| [credible_project](resources/project.md) | Projects within an organization |
| [credible_connection](resources/connection.md) | Database connections (7 types) |
| [credible_organization_permission](resources/organization_permission.md) | Org-level user/group permissions |
| [credible_project_permission](resources/project_permission.md) | Project-level user/group permissions |
| [credible_group](resources/group.md) | Groups within an organization |
| [credible_group_member](resources/group_member.md) | Group membership |
| [credible_package](resources/package.md) | Malloy model packages |
| [credible_package_version](resources/package_version.md) | Immutable package versions |
