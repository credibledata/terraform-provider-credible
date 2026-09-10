# Terraform / OpenTofu Provider for Credible

Manage [Credible](https://credibledata.com) resources declaratively — organizations, environments, database connections, permissions, groups, and packages.

The same provider works with both Terraform and OpenTofu; it is published to both registries from the same release.

- **Terraform Registry** — [registry.terraform.io/providers/credibledata/credible](https://registry.terraform.io/providers/credibledata/credible/latest)
- **OpenTofu Registry** — [search.opentofu.org/provider/credibledata/credible](https://search.opentofu.org/provider/credibledata/credible/latest)

## Requirements

- **Terraform** >= 1.0, or **OpenTofu** >= 1.6 — the provider speaks plugin protocol 6 only
- [Go](https://go.dev/dl/) >= 1.25 — only to build from source

## Installation

Declare the provider and let `terraform init` / `tofu init` fetch it.

### Terraform

```hcl
terraform {
  required_providers {
    credible = {
      source  = "credibledata/credible"
      version = "~> 0.0.7"
    }
  }
}
```

### OpenTofu

```hcl
terraform {
  required_providers {
    credible = {
      source  = "registry.opentofu.org/credibledata/credible"
      version = "~> 0.0.7"
    }
  }
}
```

OpenTofu resolves an unqualified `credibledata/credible` against its own registry, so the explicit host is optional. Naming it keeps the intent obvious in a repo that also runs Terraform.

> Versions 0.0.1–0.0.5 are indexed as plugin protocol 5.0 because of a packaging bug ([#11](https://github.com/credibledata/terraform-provider-credible/pull/11), [#13](https://github.com/credibledata/terraform-provider-credible/pull/13)); 0.0.6 is absent from the Terraform Registry for the same reason. Use **0.0.7 or later**, which both registries index correctly.

## Provider Configuration

```hcl
provider "credible" {
  url          = "https://app.credibledata.com"  # Credible API URL
  organization = "my-org"                        # Default organization
  api_key      = var.credible_api_key            # API key (recommended)
}
```

All four attributes are optional; each can come from the environment instead.

### Authentication

The provider resolves credentials in this order:

1. **API key** (`api_key`) — a service account key, sent as `Authorization: ApiKey <token>`. Recommended for CI/CD. Mutually exclusive with `bearer_token`.
2. **Bearer token** (`bearer_token`) — an OAuth2/Auth0 access token, sent as `Authorization: Bearer <token>`. Mutually exclusive with `api_key`.
3. **Credible CLI config** — if neither is set, the provider reads `~/.cred`, as written by the Credible CLI, and uses the active environment's credentials and organization.

### Environment Variables

| Attribute      | Environment variable      |
|----------------|---------------------------|
| `url`          | `CREDIBLE_URL`            |
| `organization` | `CREDIBLE_ORGANIZATION`   |
| `api_key`      | `CREDIBLE_API_KEY`        |
| `bearer_token` | `CREDIBLE_BEARER_TOKEN`   |
| *(none)*       | `CREDIBLE_CONFIG_PATH` — overrides the `~/.cred` path used by the CLI-config fallback |

## Resources

| Resource | Manages |
|---|---|
| [`credible_organization`](docs/resources/organization.md) | A Credible organization |
| [`credible_environment`](docs/resources/environment.md) | An environment within an organization |
| [`credible_connection`](docs/resources/connection.md) | A database connection within an environment |
| [`credible_package`](docs/resources/package.md) | A package within an environment |
| [`credible_package_version`](docs/resources/package_version.md) | A published, immutable version of a package |
| [`credible_group`](docs/resources/group.md) | A group within an organization |
| [`credible_group_member`](docs/resources/group_member.md) | Membership of a user or group within a group |
| [`credible_organization_permission`](docs/resources/organization_permission.md) | A permission assignment at organization level |
| [`credible_environment_permission`](docs/resources/environment_permission.md) | A permission assignment at environment level |

There are no data sources.

Full argument reference for each resource — including import ID formats and computed attributes — is in the registry documentation linked above, or under [`docs/resources/`](docs/resources/). A complete worked example is in [`examples/main.tf`](examples/main.tf).

> **Before writing a `terraform destroy`**, read [Resource lifecycle semantics](docs/resource-lifecycle.md). Organizations are owned by Credible and destroy only drops them from state; package versions are archived rather than deleted; deleting a parent orphans its children in state. That page states, per resource, what each lifecycle operation really does.

### Deletion protection

`credible_organization`, `credible_environment`, and `credible_package` set `deletion_protection = true` by default. Set it to `false` and apply *before* attempting a destroy:

```hcl
resource "credible_environment" "analytics" {
  name                = "analytics"
  deletion_protection = false
}
```

Note that imported resources are not deletion-protected until you apply — see the lifecycle page.

### Force cascade

`credible_organization` and `credible_environment` support `force_cascade` (default `false`). While it is false, deletion is refused for an organization that still contains environments, or an environment that still contains packages or connections:

```hcl
resource "credible_environment" "analytics" {
  name                = "analytics"
  deletion_protection = false
  force_cascade       = true
}
```

## Documentation

| Where | What |
|---|---|
| [Registry docs](https://registry.terraform.io/providers/credibledata/credible/latest/docs) | Rendered per-resource reference (also on [OpenTofu](https://search.opentofu.org/provider/credibledata/credible/latest)) |
| [`docs/resource-lifecycle.md`](docs/resource-lifecycle.md) | What create/read/update/destroy really do, per resource |
| [`docs/usage.md`](docs/usage.md) | Usage guide, import workflow and ID formats, do's and don'ts |
| [`examples/main.tf`](examples/main.tf) | Complete worked example |

## Development

### Building

```bash
go build -o terraform-provider-credible
```

### Running a local build

Point your CLI at the local binary with a dev override, which bypasses the registry and needs no version or namespace:

```hcl
# ~/.terraformrc  (Terraform)  or  ~/.tofurc  (OpenTofu)
provider_installation {
  dev_overrides {
    "credibledata/credible" = "/path/to/your/clone"
  }
  direct {}
}
```

With an override in place, skip `init` and run `plan` / `apply` directly.

Alternatively, `make install` builds and copies the binary into the filesystem mirror at `~/.terraform.d/plugins/`, under both the `registry.terraform.io` and `registry.opentofu.org` namespaces so either source address resolves. It installs as version `0.1.0` unless you override it (`make install VERSION=0.0.8`), and a config using the mirror must pin whichever version you chose.

### Testing

```bash
make test   # go test ./... -v  — hermetic
make vet    # go vet ./...
make fmt    # gofmt -s -w .
```

Acceptance tests live alongside the unit tests and self-skip unless credentials are present, so `go test ./...` is safe to run offline. To exercise them against a real Credible instance:

```bash
export TF_ACC=1
export CREDIBLE_URL=https://app.credibledata.com
export CREDIBLE_API_KEY=...   # or CREDIBLE_BEARER_TOKEN
go test ./internal/resources/ -v
```

Acceptance tests create and destroy real resources. Point them at a disposable organization, never production.

## Contributing

Issues and pull requests are welcome at [credibledata/terraform-provider-credible](https://github.com/credibledata/terraform-provider-credible).

Endpoints and resource shapes track the Credible Admin API, so a change to a resource's arguments usually needs the matching update in [`docs/resources/`](docs/resources/) in the same PR — the registry renders that directory, not this file.

## Support

- **Bugs and feature requests** — [open an issue](https://github.com/credibledata/terraform-provider-credible/issues)
- **Product and account questions** — [credibledata.com](https://credibledata.com)

## License

This provider is licensed under the [Mozilla Public License 2.0](LICENSE).

The source code for every released binary is this repository at the matching `v<version>` tag. Third-party dependencies retain their own licenses, including the HashiCorp `terraform-plugin-*` libraries (MPL-2.0) the provider is built on.

Use of the Credible Platform itself is governed separately by the [Credible Platform Agreement](https://credibledata.com/terms); this license covers the provider only and grants no rights in the service.
