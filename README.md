# Terraform Provider for Credible

A Terraform provider to manage [Credible](https://credibledata.com) resources declaratively — organizations, environments, database connections, permissions, groups, and packages.

## Requirements

- [Terraform](https://www.terraform.io/downloads) >= 1.0
- [Go](https://go.dev/dl/) >= 1.23 (to build from source)

## Installation

### From Source

```bash
git clone https://github.com/credibledata/terraform-provider-credible.git
cd terraform-provider-credible
make install
```

This builds the provider binary and installs it into `~/.terraform.d/plugins/`.

## Provider Configuration

```hcl
terraform {
  required_providers {
    credible = {
      source = "registry.terraform.io/credibledata/credible"
    }
  }
}

provider "credible" {
  url          = "https://app.credibledata.com"  # Credible API URL
  organization = "my-org"                         # Default organization
  api_key      = var.credible_api_key             # API key (recommended)
}
```

### Authentication

The provider supports three authentication methods, in order of precedence:

1. **API Key** (`api_key`) — A service account API key. Recommended for CI/CD.
2. **Bearer Token** (`bearer_token`) — An OAuth2/Auth0 access token.
3. **CLI Config Fallback** — If neither is set, the provider reads credentials from the Credible CLI config file (`~/.cred`).

### Environment Variables

All provider attributes can be set via environment variables:

| Attribute      | Environment Variable       |
|----------------|----------------------------|
| `url`          | `CREDIBLE_URL`             |
| `organization` | `CREDIBLE_ORGANIZATION`    |
| `api_key`      | `CREDIBLE_API_KEY`         |
| `bearer_token` | `CREDIBLE_BEARER_TOKEN`    |

## Resources

> **Before writing a `terraform destroy`**, read
> [Resource lifecycle semantics](docs/resource-lifecycle.md). Organizations are
> owned by Credible and destroy only drops them from state; package versions are
> archived rather than deleted. That page states, per resource, what each
> lifecycle operation really does.

### `credible_organization`

Manages a Credible organization.

```hcl
resource "credible_organization" "main" {
  name                = "my-org"
  display_name        = "My Organization"
  deletion_protection = true   # default: true
  force_cascade       = false  # default: false
}
```

| Attribute             | Type   | Required | Description |
|-----------------------|--------|----------|-------------|
| `name`                | string | yes      | Unique organization name. Changing forces recreation. |
| `display_name`        | string | no       | Human-readable display name. |
| `deletion_protection` | bool   | no       | Prevents `terraform destroy` when `true` (default). Set to `false` to allow deletion. |
| `force_cascade`       | bool   | no       | When `false` (default), destroy is blocked if the org contains environments. |
| `created_at`          | string | computed | Creation timestamp. |
| `updated_at`          | string | computed | Last update timestamp. |

**Import:** `terraform import credible_organization.main my-org`

> **Note:** The Credible API does not support organization deletion. Destroying this resource removes it from Terraform state only.

---

### `credible_environment`

Manages an environment within an organization.

```hcl
resource "credible_environment" "analytics" {
  name                = "analytics"
  readme              = "Analytics data models"
  replication_count   = 1
  deletion_protection = true   # default: true
  force_cascade       = false  # default: false
}
```

| Attribute             | Type   | Required | Description |
|-----------------------|--------|----------|-------------|
| `organization`        | string | no       | Organization name. Defaults to the provider's organization. |
| `name`                | string | yes      | Unique environment name. Changing forces recreation. |
| `readme`              | string | no       | Markdown-formatted environment description. |
| `replication_count`   | int    | no       | Number of replicas for high availability (1–10). |
| `deletion_protection` | bool   | no       | Prevents `terraform destroy` when `true` (default). |
| `force_cascade`       | bool   | no       | When `false` (default), destroy is blocked if the environment contains packages or connections. |
| `created_at`          | string | computed | Creation timestamp. |
| `updated_at`          | string | computed | Last update timestamp. |

**Import:** `terraform import credible_environment.analytics my-org/analytics`

---

### `credible_connection`

Manages a database connection within an environment. Exactly one connection type block must be specified.

```hcl
resource "credible_connection" "warehouse" {
  environment = credible_environment.analytics.name
  name    = "main-warehouse"
  type    = "bigquery"

  bigquery {
    default_project_id       = "my-gcp-project"
    service_account_key_json = var.bq_service_account_key
  }

  include_tables = ["analytics.*", "sales.*"]
}
```

| Attribute            | Type         | Required | Description |
|----------------------|--------------|----------|-------------|
| `organization`       | string       | no       | Defaults to provider's organization. |
| `environment`        | string       | yes      | Environment name. |
| `name`               | string       | yes      | Unique connection name. Changing forces recreation. |
| `type`               | string       | yes      | One of: `postgres`, `bigquery`, `snowflake`, `trino`, `databricks`, `mysql`, `duckdb`, `motherduck`, `ducklake`, `publisher`. |
| `include_tables`     | list(string) | no       | Tables to include (glob patterns). |
| `exclude_tables`     | list(string) | no       | Tables to exclude (glob patterns). |
| `exclude_all_tables` | bool         | no       | Exclude all tables from indexing. |

**Connection type blocks:**

<details>
<summary><code>postgres</code></summary>

| Attribute           | Type   | Sensitive | Description |
|---------------------|--------|-----------|-------------|
| `host`              | string | no        | Database host. |
| `port`              | int    | no        | Database port. |
| `database_name`     | string | no        | Database name. |
| `user_name`         | string | no        | Username. |
| `password`          | string | yes       | Password. |
| `connection_string` | string | yes       | Full connection string (alternative to individual fields). |
| `sslmode`           | string | no        | `disable`, `no-verify`, `verify-ca`, `verify-full`. Only valid on a proxied connection. |
</details>

<details>
<summary><code>bigquery</code></summary>

| Attribute                    | Type   | Sensitive | Description |
|------------------------------|--------|-----------|-------------|
| `default_project_id`         | string | no        | Default GCP project ID. |
| `billing_project_id`         | string | no        | Billing project ID. |
| `location`                   | string | no        | Dataset location. |
| `service_account_key_json`   | string | yes       | Service account key JSON. |
| `maximum_bytes_billed`       | string | no        | Maximum bytes billed per query. |
| `query_timeout_milliseconds` | string | no        | Query timeout in milliseconds. |
| `impersonate_service_account` | string | no       | Service account to impersonate. Mutually exclusive with `service_account_key_json`. |
</details>

<details>
<summary><code>snowflake</code></summary>

| Attribute                        | Type   | Sensitive | Description |
|----------------------------------|--------|-----------|-------------|
| `account`                        | string | no        | Snowflake account identifier. |
| `username`                       | string | no        | Username. |
| `password`                       | string | yes       | Password. |
| `private_key`                    | string | yes       | Private key for key-pair auth. |
| `private_key_pass`               | string | yes       | Private key passphrase. |
| `warehouse`                      | string | no        | Warehouse name. |
| `database`                       | string | no        | Database name. |
| `schema`                         | string | no        | Schema name. |
| `role`                           | string | no        | Role name. |
| `response_timeout_milliseconds`  | int    | no        | Response timeout. |
</details>

<details>
<summary><code>trino</code></summary>

| Attribute | Type   | Description |
|-----------|--------|-------------|
| `server`  | string | Trino server hostname. |
| `port`    | int    | Trino server port. |
| `catalog` | string | Catalog name. |
| `schema`  | string | Schema name. |
| `user`    | string | Username. |
| `password` | string | Password (sensitive). |
| `peaka_key` | string | Peaka API key, for Peaka-hosted Trino clusters (sensitive). |
</details>

<details>
<summary><code>databricks</code></summary>

| Attribute | Type | Sensitive | Description |
|-----------|------|-----------|-------------|
| `host` | string | no | Workspace host (e.g. `dbc-xxxxxxxx-xxxx.cloud.databricks.com`). |
| `path` | string | no | SQL warehouse HTTP path. |
| `token` | string | yes | Personal access token. |
| `oauth_client_id` | string | no | OAuth M2M client ID (service principal). |
| `oauth_client_secret` | string | yes | OAuth M2M client secret. |
| `default_catalog` | string | no | Default Unity Catalog. |
| `default_schema` | string | no | Default schema. |
| `setup_sql` | string | no | SQL run when the connection is established. |
</details>

<details>
<summary><code>mysql</code></summary>

| Attribute  | Type   | Sensitive | Description |
|------------|--------|-----------|-------------|
| `host`     | string | no        | Database host. |
| `port`     | int    | no        | Database port. |
| `database` | string | no        | Database name. |
| `user`     | string | no        | Username. |
| `password` | string | yes       | Password. |
</details>

<details>
<summary><code>duckdb</code></summary>

Contains repeatable `attached_databases` blocks (`name`, `type`, plus the matching
`bigquery` / `snowflake` / `postgres` / `gcs` / `s3` / `azure` block); it has no
direct attributes. See the [connection docs](docs/resources/connection.md).
</details>

<details>
<summary><code>motherduck</code></summary>

| Attribute      | Type   | Sensitive | Description |
|----------------|--------|-----------|-------------|
| `access_token` | string | yes       | MotherDuck access token. |
| `database`     | string | no        | MotherDuck database name. |
</details>

<details>
<summary><code>ducklake</code></summary>

Contains a `storage` block (`bucket_url`, plus an `s3` or `gcs` block) and a
`catalog` block (a `postgres` block, plus an optional `metadata_schema`). See the
[connection docs](docs/resources/connection.md).
</details>

<details>
<summary><code>publisher</code></summary>

| Attribute | Type | Sensitive | Description |
|-----------|------|-----------|-------------|
| `connection_uri` | string | no | Full URI of the remote connection. Required when this block is set. |
| `access_token` | string | yes | Bearer token for the remote dataplane. |
</details>

<details>
<summary><code>proxy</code> (any connection type)</summary>

For a database that is not directly routable. `type` is `ssh`; the nested `ssh`
block takes `host`, `port` (defaults to 22), `username`, `private_key` (sensitive),
`private_key_pass` (sensitive) and `host_key`. See the
[connection docs](docs/resources/connection.md).
</details>

**Import:** `terraform import credible_connection.warehouse my-org/analytics/main-warehouse`

---

### `credible_organization_permission`

Manages organization-level permissions for a user or group.

```hcl
resource "credible_organization_permission" "alice_admin" {
  user_group_id = "user:alice@example.com"
  permission    = "admin"
}
```

| Attribute       | Type   | Required | Description |
|-----------------|--------|----------|-------------|
| `organization`  | string | no       | Defaults to provider's organization. |
| `user_group_id` | string | yes      | `user:<email>` or `group:<name>`. Changing forces recreation. |
| `permission`    | string | yes      | One of: `admin`, `modeler`, `member`. |

**Import:** `terraform import credible_organization_permission.alice_admin my-org/user:alice@example.com`

---

### `credible_environment_permission`

Manages environment-level permissions for a user or group.

```hcl
resource "credible_environment_permission" "data_team_viewer" {
  environment   = "analytics"
  user_group_id = "group:data-engineering"
  permission    = "viewer"
}
```

| Attribute       | Type   | Required | Description |
|-----------------|--------|----------|-------------|
| `organization`  | string | no       | Defaults to provider's organization. |
| `environment`   | string | yes      | Environment name. |
| `user_group_id` | string | yes      | `user:<email>` or `group:<name>`. Changing forces recreation. |
| `permission`    | string | yes      | One of: `admin`, `modeler`, `viewer`. |

**Import:** `terraform import credible_environment_permission.data_team_viewer my-org/analytics/group:data-engineering`

---

### `credible_group`

Manages a group within an organization.

```hcl
resource "credible_group" "data_eng" {
  name        = "data-engineering"
  description = "Data engineering team"
}
```

| Attribute      | Type   | Required | Description |
|----------------|--------|----------|-------------|
| `organization` | string | no       | Defaults to provider's organization. |
| `name`         | string | yes      | Unique group name. Changing forces recreation. |
| `description`  | string | no       | Group description. |

**Import:** `terraform import credible_group.data_eng my-org/data-engineering`

---

### `credible_group_member`

Manages membership of a user or group within a group.

```hcl
resource "credible_group_member" "alice" {
  group_name    = credible_group.data_eng.name
  user_group_id = "user:alice@example.com"
  status        = "member"
}
```

| Attribute       | Type   | Required | Description |
|-----------------|--------|----------|-------------|
| `organization`  | string | no       | Defaults to provider's organization. |
| `group_name`    | string | yes      | Group name. Changing forces recreation. |
| `user_group_id` | string | yes      | `user:<email>` or `group:<name>`. Changing forces recreation. |
| `status`        | string | yes      | One of: `member`, `admin`. |

**Import:** `terraform import credible_group_member.alice my-org/data-engineering/user:alice@example.com`

---

### `credible_package`

Manages a package within an environment.

```hcl
resource "credible_package" "models" {
  environment         = credible_environment.analytics.name
  name                = "analytics-models"
  description         = "Core analytics Malloy models"
  deletion_protection = true  # default: true
}
```

| Attribute             | Type   | Required | Description |
|-----------------------|--------|----------|-------------|
| `organization`        | string | no       | Defaults to provider's organization. |
| `environment`         | string | yes      | Environment name. |
| `name`                | string | yes      | Unique package name. Changing forces recreation. |
| `description`         | string | no       | Package description. |
| `deletion_protection` | bool   | no       | Prevents `terraform destroy` when `true` (default). |
| `latest_version`      | string | computed | Latest published version. |
| `created_at`          | string | computed | Creation timestamp. |
| `updated_at`          | string | computed | Last update timestamp. |

**Import:** `terraform import credible_package.models my-org/analytics/analytics-models`

---

### `credible_package_version`

Publishes a version of a package. Versions are immutable once created.

```hcl
# From a local directory (provider creates the .tar.gz archive)
resource "credible_package_version" "v1" {
  environment  = credible_environment.analytics.name
  package_name = credible_package.models.name
  version_id   = "1.0.0"
  source_dir   = "${path.module}/models/analytics"
}

# From a pre-built archive
resource "credible_package_version" "v2" {
  environment  = credible_environment.analytics.name
  package_name = credible_package.models.name
  version_id   = "2.0.0"
  source_file  = "${path.module}/dist/analytics-models.tar.gz"
  source_hash  = filemd5("${path.module}/dist/analytics-models.tar.gz")
}
```

| Attribute        | Type   | Required | Description |
|------------------|--------|----------|-------------|
| `organization`   | string | no       | Defaults to provider's organization. |
| `environment`    | string | yes      | Environment name. |
| `package_name`   | string | yes      | Package name. |
| `version_id`     | string | yes      | Semantic version (e.g. `1.0.0`). Changing forces recreation. |
| `source_dir`     | string | no       | Path to a local directory. Provider archives it as `.tar.gz`. Conflicts with `source_file`. |
| `source_file`    | string | no       | Path to a pre-built `.tar.gz` archive. Conflicts with `source_dir`. |
| `source_hash`    | string | no       | Hash for change detection. Use `filemd5()` on the source file. |
| `archive_status` | string | no       | Set to `archive` or `unarchive`. |
| `index_status`   | string | computed | Current indexing status. |
| `created_at`     | string | computed | Creation timestamp. |
| `updated_at`     | string | computed | Last update timestamp. |

---

## Deletion Protection

Organizations, environments, and packages have `deletion_protection = true` by default. This prevents accidental `terraform destroy` operations. To destroy a protected resource:

```hcl
resource "credible_environment" "analytics" {
  name                = "analytics"
  deletion_protection = false  # must be set before destroy
}
```

Then run `terraform apply` followed by `terraform destroy`.

## Force Cascade

Organizations and environments support `force_cascade` (default `false`). When disabled, Terraform will refuse to delete:

- An **organization** that still contains environments
- An **environment** that still contains packages or connections

Set `force_cascade = true` to override this check:

```hcl
resource "credible_environment" "analytics" {
  name                = "analytics"
  deletion_protection = false
  force_cascade       = true  # allow deletion even with child resources
}
```

## Full Example

See [examples/main.tf](examples/main.tf) for a complete working example.

## Development

```bash
# Build
go build -o terraform-provider-credible

# Install locally
make install

# Run against a local controlplane
cd examples
terraform init
terraform plan
terraform apply
```

## License

This provider is licensed under the [Mozilla Public License 2.0](LICENSE).

The source code for every released binary is this repository at the matching
`v<version>` tag. Third-party dependencies retain their own licenses, including
the HashiCorp `terraform-plugin-*` libraries (MPL-2.0) the provider is built on.

Use of the Credible Platform itself is governed separately by the
[Credible Platform Agreement](https://credibledata.com/terms); this license
covers the provider only and grants no rights in the service.
