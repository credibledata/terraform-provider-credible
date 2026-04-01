---
page_title: "credible_connection Resource - credible"
subcategory: ""
description: |-
  Manages a database connection within a Credible project. Supports PostgreSQL, BigQuery, Snowflake, Trino, MySQL, DuckDB, and MotherDuck.
---

# credible_connection (Resource)

Manages a database connection within a Credible project. Exactly one connection type block must be specified, matching the `type` attribute.

Supported connection types: `postgres`, `bigquery`, `snowflake`, `trino`, `mysql`, `duckdb`, `motherduck`.

~> **Sensitive fields** (passwords, keys, tokens) are never returned by the API after creation. Terraform preserves them from your configuration. You will see plan diffs for these fields if you import an existing connection — you must fill in the sensitive values in your HCL.

## Example Usage

### BigQuery

```hcl
resource "credible_connection" "warehouse" {
  project = credible_project.analytics.name
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
  project = "analytics"
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
  project = "analytics"
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
  project = "analytics"
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
  project = "analytics"
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

### DuckDB / MotherDuck

```hcl
resource "credible_connection" "duckdb" {
  project = "analytics"
  name    = "local-duckdb"
  type    = "duckdb"

  duckdb {
    url      = "md:my_database"
    md_token = var.motherduck_token
  }
}
```

### Import an existing connection into Terraform

**Step 1:** Write the resource block with the connection type block:

```hcl
resource "credible_connection" "warehouse" {
  project = "analytics"
  name    = "main-warehouse"
  type    = "bigquery"

  bigquery {
    default_project_id       = "my-gcp-project"
    service_account_key_json = var.bq_service_account_key  # must provide — API won't return it
  }
}
```

**Step 2:** Import using `<organization>/<project>/<connection>`:

```shell
terraform import credible_connection.warehouse my-org/analytics/main-warehouse
```

**Step 3:** Run `terraform plan`. You will see diffs for sensitive fields — this is expected. Ensure your HCL has the correct credential values, then run `terraform apply` to sync state.

## Schema

### Required

- `project` (String) — Project name. **Immutable** — changing forces destroy and recreate.
- `name` (String) — Unique connection name within the project. **Immutable**.
- `type` (String) — Connection type. One of: `postgres`, `bigquery`, `snowflake`, `trino`, `mysql`, `duckdb`, `motherduck`. **Immutable**.

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

#### `bigquery`

| Attribute | Type | Sensitive | Description |
|---|---|---|---|
| `default_project_id` | String | No | Default GCP project ID |
| `billing_project_id` | String | No | Billing project ID |
| `location` | String | No | Dataset location |
| `service_account_key_json` | String | **Yes** | Service account key JSON |
| `maximum_bytes_billed` | String | No | Maximum bytes billed per query |
| `query_timeout_milliseconds` | String | No | Query timeout in milliseconds |

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

#### `mysql`

| Attribute | Type | Sensitive | Description |
|---|---|---|---|
| `host` | String | No | MySQL server hostname |
| `port` | Number | No | MySQL server port |
| `database` | String | No | Database name |
| `user` | String | No | Username |
| `password` | String | **Yes** | Password |

#### `duckdb`

| Attribute | Type | Sensitive | Description |
|---|---|---|---|
| `url` | String | No | DuckDB connection URL |
| `md_token` | String | **Yes** | MotherDuck token |

#### `motherduck`

| Attribute | Type | Sensitive | Description |
|---|---|---|---|
| `url` | String | No | MotherDuck connection URL |
| `md_token` | String | **Yes** | MotherDuck token |

## Import

Import format: `<organization>/<project>/<connection>`

```shell
terraform import credible_connection.warehouse my-org/analytics/main-warehouse
```

~> After importing a connection, sensitive fields (passwords, keys, tokens) will be empty in state. You must provide them in your HCL configuration and run `terraform apply` to sync.
