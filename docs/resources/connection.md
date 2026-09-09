---
page_title: "credible_connection Resource - credible"
subcategory: ""
description: |-
  Manages a database connection within a Credible environment. Supports PostgreSQL, BigQuery, Snowflake, Trino, Databricks, MySQL, DuckDB, MotherDuck, DuckLake, and Publisher.
---

# credible_connection (Resource)

Manages a database connection within a Credible environment. Exactly one connection type block must be specified, matching the `type` attribute.

Supported connection types: `postgres`, `bigquery`, `snowflake`, `trino`,
`databricks`, `mysql`, `duckdb`, `motherduck`, `ducklake`, `publisher`.

!> **Breaking change to `duckdb` and `motherduck`.** Both blocks previously took
`url` and `md_token`, neither of which the API defines. `motherduck` now takes
`access_token` and `database`. `duckdb` no longer takes a token or URL at all: it
configures the warehouses DuckDB attaches to, via repeatable `attached_databases`
blocks. A configuration using the old attributes fails validation and must be
rewritten; use `motherduck` for a MotherDuck-backed connection.

~> **Sensitive fields** (passwords, keys, tokens) are never returned by the API after creation. Terraform preserves them from your configuration. You will see plan diffs for these fields if you import an existing connection — you must fill in the sensitive values in your HCL.

## Example Usage

### BigQuery

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

### PostgreSQL

```hcl
resource "credible_connection" "app_db" {
  environment = "analytics"
  name    = "app-database"
  type    = "postgres"

  postgres {
    host          = "db.example.com"
    port          = 5432
    database_name = "appdb"
    user_name     = "readonly_user"
    password      = var.db_password
  }

  include_tables = ["public.*"]
}
```

### Snowflake

```hcl
resource "credible_connection" "snowflake" {
  environment = "analytics"
  name    = "snowflake-warehouse"
  type    = "snowflake"

  snowflake {
    account   = "xy12345.us-east-1"
    username  = "terraform_user"
    password  = var.snowflake_password
    warehouse = "COMPUTE_WH"
    database  = "ANALYTICS"
    schema    = "PUBLIC"
    role      = "SYSADMIN"
  }
}
```

### MySQL

```hcl
resource "credible_connection" "mysql" {
  environment = "analytics"
  name    = "mysql-db"
  type    = "mysql"

  mysql {
    host     = "mysql.example.com"
    port     = 3306
    database = "appdb"
    user     = "reader"
    password = var.mysql_password
  }
}
```

### Trino

```hcl
resource "credible_connection" "trino" {
  environment = "analytics"
  name    = "trino-cluster"
  type    = "trino"

  trino {
    server  = "trino.example.com"
    port    = 8080
    catalog = "hive"
    schema  = "default"
    user    = "analyst"
  }
}
```

### MotherDuck

```hcl
resource "credible_connection" "motherduck" {
  environment = "analytics"
  name        = "md-warehouse"
  type        = "motherduck"

  motherduck {
    access_token = var.motherduck_token
    database     = "my_database"
  }
}
```

### DuckDB

DuckDB exposes data-source intent only: database files, filesystem policy and
resource limits are owned by the deployment, so the block configures the warehouses
DuckDB attaches to rather than a database URL.

```hcl
resource "credible_connection" "duckdb" {
  environment = "analytics"
  name        = "lake"
  type        = "duckdb"

  duckdb {
    attached_databases {
      name = "events"
      type = "s3"

      s3 {
        provider          = "config"
        access_key_id     = var.aws_access_key_id
        secret_access_key = var.aws_secret_access_key
        region            = "us-west-2"
      }
    }
  }
}
```

### Databricks

```hcl
resource "credible_connection" "databricks" {
  environment = "analytics"
  name        = "dbx-warehouse"
  type        = "databricks"

  databricks {
    host            = "dbc-xxxxxxxx-xxxx.cloud.databricks.com"
    path            = "/sql/1.0/warehouses/abc123"
    token           = var.databricks_token
    default_catalog = "main"
    default_schema  = "public"
  }
}
```

### DuckLake

```hcl
resource "credible_connection" "ducklake" {
  environment = "analytics"
  name        = "lakehouse"
  type        = "ducklake"

  ducklake {
    storage {
      bucket_url = "s3://my-bucket/lakehouse"

      s3 {
        provider          = "config"
        access_key_id     = var.aws_access_key_id
        secret_access_key = var.aws_secret_access_key
      }
    }

    catalog {
      metadata_schema = "ducklake_meta"

      postgres {
        host          = "catalog.example.com"
        port          = 5432
        database_name = "ducklake_catalog"
        user_name     = "ducklake"
        password      = var.catalog_password
      }
    }
  }
}
```

### Publisher

Proxies SQL to a remote Publisher dataplane instead of connecting to a warehouse
directly. The remote dataplane owns authentication and access control.

```hcl
resource "credible_connection" "remote" {
  environment = "analytics"
  name        = "partner-data"
  type        = "publisher"

  publisher {
    connection_uri = "https://org.data.example.com/api/v0/environments/prod/connections/sales"
    access_token   = var.remote_token
  }
}
```

### Reaching a database through an SSH bastion

Any connection type can be reached through a proxy when the database is not
directly routable. The proxy is established below the driver, so the driver
connects to a local endpoint transparently.

```hcl
resource "credible_connection" "private_pg" {
  environment = "analytics"
  name        = "private-postgres"
  type        = "postgres"

  postgres {
    host          = "db.internal"
    port          = 5432
    database_name = "app"
    user_name     = "analyst"
    password      = var.db_password
    sslmode       = "verify-full"
  }

  proxy {
    type = "ssh"

    ssh {
      host        = "bastion.example.com"
      username    = "terraform"
      private_key = var.bastion_private_key
      host_key    = var.bastion_host_key
    }
  }
}
```

### Import an existing connection into Terraform

**Step 1:** Write the resource block with the connection type block:

```hcl
resource "credible_connection" "warehouse" {
  environment = "analytics"
  name    = "main-warehouse"
  type    = "bigquery"

  bigquery {
    default_project_id       = "my-gcp-project"
    service_account_key_json = var.bq_service_account_key  # must provide — API won't return it
  }
}
```

**Step 2:** Import using `<organization>/<environment>/<connection>`:

```shell
terraform import credible_connection.warehouse my-org/analytics/main-warehouse
```

**Step 3:** Run `terraform plan`. You will see diffs for sensitive fields — this is expected. Ensure your HCL has the correct credential values, then run `terraform apply` to sync state.

## Schema

### Required

- `environment` (String) — Environment name. **Immutable** — changing forces destroy and recreate.
- `name` (String) — Unique connection name within the environment. **Immutable**.
- `type` (String) — Connection type. One of: `postgres`, `bigquery`, `snowflake`, `trino`, `databricks`, `mysql`, `duckdb`, `motherduck`, `ducklake`, `publisher`. **Immutable**.

### Optional

- `organization` (String) — Organization name. **Default: provider's `organization`**. **Immutable**.
- `include_tables` (List of String) — Glob patterns for tables to include (e.g., `schema.*`, `schema.table`). Can be updated.
- `exclude_tables` (List of String) — Glob patterns for tables to exclude. Can be updated.
- `exclude_all_tables` (Boolean) — Exclude all tables from indexing. Can be updated.

### Read-Only

- `indexing_status` (String) — Current indexing status of the connection.

### Connection Type Blocks

Exactly one of the following blocks must be specified, matching the `type` attribute.

#### `postgres`

| Attribute | Type | Sensitive | Description |
|---|---|---|---|
| `host` | String | No | PostgreSQL server hostname |
| `port` | Number | No | PostgreSQL server port |
| `database_name` | String | No | Database name |
| `user_name` | String | No | Username |
| `password` | String | **Yes** | Password |
| `connection_string` | String | **Yes** | Full connection string (alternative to individual fields) |
| `sslmode` | String | No | TLS mode: `disable`, `no-verify`, `verify-ca`, `verify-full`. Only valid on a proxied connection; a direct connection rejects it |

#### `bigquery`

| Attribute | Type | Sensitive | Description |
|---|---|---|---|
| `default_project_id` | String | No | Default GCP project ID |
| `billing_project_id` | String | No | Billing project ID |
| `location` | String | No | Dataset location |
| `service_account_key_json` | String | **Yes** | Service account key JSON |
| `maximum_bytes_billed` | String | No | Maximum bytes billed per query |
| `query_timeout_milliseconds` | String | No | Query timeout in milliseconds |
| `impersonate_service_account` | String | No | Service account to impersonate. Mutually exclusive with `service_account_key_json` |

#### `snowflake`

| Attribute | Type | Sensitive | Description |
|---|---|---|---|
| `account` | String | No | Snowflake account identifier |
| `username` | String | No | Username |
| `password` | String | **Yes** | Password |
| `private_key` | String | **Yes** | Private key for key-pair auth |
| `private_key_pass` | String | **Yes** | Private key passphrase |
| `warehouse` | String | No | Warehouse name |
| `database` | String | No | Database name |
| `schema` | String | No | Schema name |
| `role` | String | No | Role name |
| `response_timeout_milliseconds` | Number | No | Response timeout |

#### `trino`

| Attribute | Type | Description |
|---|---|---|
| `server` | String | Trino server hostname |
| `port` | Number | Trino server port |
| `catalog` | String | Catalog name |
| `schema` | String | Schema name |
| `user` | String | Username |
| `password` | String | Password (sensitive) |
| `peaka_key` | String | Peaka API key, for Peaka-hosted Trino clusters (sensitive) |

#### `databricks`

| Attribute | Type | Sensitive | Description |
|---|---|---|---|
| `host` | String | No | Workspace host (e.g. `dbc-xxxxxxxx-xxxx.cloud.databricks.com`) |
| `path` | String | No | SQL warehouse HTTP path |
| `token` | String | **Yes** | Personal access token |
| `oauth_client_id` | String | No | OAuth M2M client ID (service principal) |
| `oauth_client_secret` | String | **Yes** | OAuth M2M client secret |
| `default_catalog` | String | No | Default Unity Catalog |
| `default_schema` | String | No | Default schema |
| `setup_sql` | String | No | SQL run when the connection is established |

#### `mysql`

| Attribute | Type | Sensitive | Description |
|---|---|---|---|
| `host` | String | No | MySQL server hostname |
| `port` | Number | No | MySQL server port |
| `database` | String | No | Database name |
| `user` | String | No | Username |
| `password` | String | **Yes** | Password |

#### `duckdb`

Contains repeatable `attached_databases` blocks; it has no direct attributes.

Each `attached_databases` block takes `name` (a plain identifier) and `type`
(`bigquery`, `snowflake`, `postgres`, `gcs`, `s3`, `azure`), plus the matching
nested credential block below.

#### `motherduck`

| Attribute | Type | Sensitive | Description |
|---|---|---|---|
| `access_token` | String | **Yes** | MotherDuck access token |
| `database` | String | No | MotherDuck database name |

#### `ducklake`

Contains a `storage` block and a `catalog` block.

`storage` takes `bucket_url` (required when the block is set) plus one of the `s3`
or `gcs` blocks. `catalog` takes a `postgres` block (required when the block is
set) and an optional `metadata_schema`, which lets several DuckLake catalogs share
one catalog database. `metadata_schema` is an organizational boundary, not an
access-control one -- any config whose catalog role can reach a sibling schema can
read that catalog's metadata.

#### `publisher`

| Attribute | Type | Sensitive | Description |
|---|---|---|---|
| `connection_uri` | String | No | Full URI of the remote connection. Required when this block is set |
| `access_token` | String | **Yes** | Bearer token for the remote dataplane |

#### `s3` (nested)

| Attribute | Type | Sensitive | Description |
|---|---|---|---|
| `provider` | String | No | `config` (requires the key pair below) or `credential_chain` (resolves from the host and rejects a supplied key). The API defaults to `config` |
| `chain` | String | No | Semicolon-separated credential providers, for `credential_chain` |
| `access_key_id` | String | No | AWS access key ID |
| `secret_access_key` | String | **Yes** | AWS secret access key |
| `region` | String | No | AWS region. The API defaults to `us-east-1` |
| `endpoint` | String | No | Custom S3-compatible endpoint (e.g. MinIO) |
| `session_token` | String | **Yes** | Session token for temporary credentials |

#### `gcs` (nested)

| Attribute | Type | Sensitive | Description |
|---|---|---|---|
| `key_id` | String | No | GCS HMAC access key ID. Required when the block is set |
| `secret` | String | **Yes** | GCS HMAC secret key. Required when the block is set |

#### `azure` (nested)

| Attribute | Type | Sensitive | Description |
|---|---|---|---|
| `auth_type` | String | No | `service_principal` or `sas_token`. Required when the block is set |
| `sas_url` | String | **Yes** | Full SAS URL including token; required for `sas_token` |
| `tenant_id` | String | No | Azure AD tenant ID; required for `service_principal` |
| `client_id` | String | No | Azure AD application (client) ID; required for `service_principal` |
| `client_secret` | String | **Yes** | Azure AD client secret; required for `service_principal` |
| `account_name` | String | No | Storage account name; required for `service_principal` |
| `file_url` | String | No | Azure file URL to query; required for `service_principal` |

### `proxy`

Optional on any connection type, for a database that is not directly routable.

| Attribute | Type | Description |
|---|---|---|
| `type` | String | Proxy mechanism. Only `ssh` is supported |

The nested `ssh` block:

| Attribute | Type | Sensitive | Description |
|---|---|---|---|
| `host` | String | No | Bastion hostname or IP address |
| `port` | Number | No | Bastion SSH port. The API defaults to 22 |
| `username` | String | No | SSH username on the bastion |
| `private_key` | String | **Yes** | PEM-encoded private key. Leave blank when updating to keep the stored key |
| `private_key_pass` | String | **Yes** | Passphrase for the private key. Leave blank when updating to keep the stored one |
| `host_key` | String | No | Pinned bastion host public key(s) as OpenSSH `known_hosts` lines. Omitted means the tunnel connects without host-key verification |

## Import

Import format: `<organization>/<environment>/<connection>`

```shell
terraform import credible_connection.warehouse my-org/analytics/main-warehouse
```

~> After importing a connection, sensitive fields (passwords, keys, tokens) will be empty in state. You must provide them in your HCL configuration and run `terraform apply` to sync.
