---
page_title: "credible_package_version Resource - credible"
subcategory: ""
description: |-
  Publishes an immutable version of a Malloy model package in Credible.
---

# credible_package_version (Resource)

Publishes a version of a Credible package. Versions are **immutable** once created — they cannot be modified or truly deleted.

~> **Important:** `terraform destroy` on a package version does not delete it. It **archives** the version instead. Archived versions can be unarchived by setting `archive_status = "unarchive"`.

## Example Usage

### Publish from a local directory

The provider automatically creates a `.tar.gz` archive from the directory contents.

```hcl
resource "credible_package_version" "v1" {
  project      = credible_project.analytics.name
  package_name = credible_package.models.name
  version_id   = "1.0.0"
  source_dir   = "${path.module}/models/analytics"
}
```

### Publish from a pre-built archive

```hcl
resource "credible_package_version" "v2" {
  project      = credible_project.analytics.name
  package_name = credible_package.models.name
  version_id   = "2.0.0"
  source_file  = "${path.module}/dist/analytics-models.tar.gz"
  source_hash  = filemd5("${path.module}/dist/analytics-models.tar.gz")
}
```

### Archive a version

```hcl
resource "credible_package_version" "v1" {
  project        = "analytics"
  package_name   = "analytics-models"
  version_id     = "1.0.0"
  source_dir     = "${path.module}/models/analytics"
  archive_status = "archive"
}
```

### Full lifecycle: package + version

```hcl
resource "credible_package" "models" {
  project     = "analytics"
  name        = "analytics-models"
  description = "Core analytics Malloy models"
}

resource "credible_package_version" "v1" {
  project      = "analytics"
  package_name = credible_package.models.name
  version_id   = "1.0.0"
  source_dir   = "${path.module}/models"
}
```

## Schema

### Required

- `project` (String) — Project name. **Immutable** — changing forces destroy and recreate.
- `package_name` (String) — Package name. **Immutable**.
- `version_id` (String) — Semantic version identifier (e.g., `1.0.0`). **Immutable**.

### Optional

- `organization` (String) — Organization name. **Default: provider's `organization`**. **Immutable**.
- `source_dir` (String) — Path to a local directory. The provider creates a `.tar.gz` archive from its contents and uploads it. **Conflicts with `source_file`**. **Immutable**.
- `source_file` (String) — Path to a pre-built `.tar.gz` archive to upload. **Conflicts with `source_dir`**. **Immutable**.
- `source_hash` (String) — Hash for change detection. Use `filemd5()` on the source file. Useful for detecting when the archive contents have changed.
- `archive_status` (String) — Set to `archive` to archive the version or `unarchive` to restore it. Can be updated in place.

### Read-Only

- `index_status` (String) — Current indexing status (e.g., `indexed`, `indexing`, `failed`).
- `created_at` (String) — ISO 8601 creation timestamp.
- `updated_at` (String) — ISO 8601 last-update timestamp.

## Important Behaviors

- **Versions are immutable.** Once created, the contents cannot be changed. To publish new content, create a new version with a bumped `version_id`.
- **Destroy = archive.** Running `terraform destroy` on a package version archives it rather than deleting it. The version still exists on the platform.
- **`source_dir` vs `source_file`** — Use `source_dir` to point at a directory of Malloy files; the provider handles archiving. Use `source_file` if you have a CI pipeline that produces the archive.
- **Change detection** — Use `source_hash = filemd5(...)` with `source_file` to detect when the archive changes. With `source_dir`, changes to the directory path trigger a recreate.

## Import

Package versions cannot be imported because `source_dir`/`source_file` are local paths that cannot be recovered from the API. To manage an existing version, add it to your Terraform config and create it fresh (the API will reject duplicate `version_id` values — use a new version number).
