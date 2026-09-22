---
page_title: "credible_package_version Resource - credible"
subcategory: ""
description: |-
  Publishes an immutable version of a Malloy model package in Credible, which is also how a package is created.
---

# credible_package_version (Resource)

Publishes a version of a Credible package. Publishing is also how a package is **created** — the Admin API has no metadata-only create, so this resource stands on its own and does not need a `credible_package` resource alongside it. Use `credible_package` to manage the metadata of a package that already exists.

Versions are **immutable** once published: their contents cannot be changed. To publish new content, add a version with a bumped `version_id`.

~> **Important:** `terraform destroy` on a package version does not delete it. It **archives** the version instead. Archived versions can be unarchived by setting `archive_status = "unarchive"`.

## Example Usage

### Publish from a local directory

The provider zips the directory contents and uploads them. Paths inside the archive are relative to `source_dir`, so `publisher.json` must sit at the top level of that directory.

```hcl
resource "credible_package_version" "v1" {
  environment  = credible_environment.analytics.name
  package_name = "analytics-models"
  version_id   = "1.0.0"
  description  = "Analytics models"
  source_dir   = "${path.module}/models/analytics"
}
```

### Publish from a pre-built archive

```hcl
resource "credible_package_version" "v2" {
  environment  = credible_environment.analytics.name
  package_name = "analytics-models"
  version_id   = "2.0.0"
  source_file  = "${path.module}/dist/analytics-models.zip"
  source_hash  = filemd5("${path.module}/dist/analytics-models.zip")
}
```

-> `source_file` must be a **zip**. The API inspects the archive's bytes, so another format (a `.tar.gz`, for instance) is rejected whatever the file is named.

### Archive a version

```hcl
resource "credible_package_version" "v1" {
  environment    = "analytics"
  package_name   = "analytics-models"
  version_id     = "1.0.0"
  source_dir     = "${path.module}/models/analytics"
  archive_status = "archive"
}
```

## Schema

### Required

- `environment` (String) — Environment name. **Immutable** — changing forces destroy and recreate.
- `package_name` (String) — Package name. **Immutable**.
- `version_id` (String) — Semantic version identifier (e.g., `1.0.0`). **Immutable**.

### Optional

- `organization` (String) — Organization name. **Default: provider's `organization`**. **Immutable**.
- `description` (String) — Description applied to the package by this publish. Publishing is the only operation that creates a package, so the description travels with the version. It describes the **package**, not the version, and is only ever sent by the publish: editing it afterwards updates Terraform state without changing the package, because the API ignores a description in a package update.
- `source_dir` (String) — Path to a local directory. The provider zips its contents and uploads them. **Conflicts with `source_file`**. **Immutable**.
- `source_file` (String) — Path to a pre-built `.zip` archive to upload. **Conflicts with `source_dir`**. **Immutable**.
- `source_hash` (String) — Hash for change detection. Use `filemd5()` on the source file. Useful for detecting when the archive contents have changed.
- `archive_status` (String) — Set to `archive` to archive the version or `unarchive` to restore it. Can be updated in place.

### Read-Only

- `build_status` (String) — Build status of the version (e.g., `BUILDING`, `READY`).
- `created_at` (String) — ISO 8601 creation timestamp.
- `updated_at` (String) — ISO 8601 last-update timestamp.

## Important Behaviors

- **Publishing creates the package.** A publish to a package name that does not exist yet creates it; a publish to one that does adds a version. Either way the API returns the package, and this resource reads the new version back for its own state.
- **Versions are immutable.** Once published, the contents cannot be changed. To publish new content, create a new version with a bumped `version_id`.
- **Destroy = archive.** Running `terraform destroy` on a package version archives it rather than deleting it. The version still exists on the platform.
- **Builds complete asynchronously.** `build_status` is whatever it was at the moment of the publish, commonly `BUILDING`. A later `terraform refresh` reports it as `READY` once the build finishes.
- **Re-publishing a `version_id` is rejected.** The API answers `409 VersionId already exists in package`. Because the source attributes force replacement, editing `source_dir`/`source_file`/`version_id` in place archives the existing version and then fails to re-publish it — bump `version_id` for new content instead.
- **`source_dir` vs `source_file`** — Use `source_dir` to point at a directory of Malloy files; the provider handles archiving. Use `source_file` if you have a CI pipeline that produces the zip.
- **Change detection** — `source_hash` is the only thing that makes Terraform notice edited content. It is optional, and if you omit it, editing files under `source_dir` produces **no diff at all** — Terraform reports "no changes" indefinitely. Only the directory *path* is tracked, not the files in it. With `source_file`, use `source_hash = filemd5(...)`. With `source_dir`, hash the files:

    ```hcl
    source_hash = sha256(join("", [for f in fileset(path.module, "models/**") : filesha256("${path.module}/${f}")]))
    ```

## Import

Package versions cannot be imported: `source_dir`/`source_file` are local paths that cannot be recovered from the API, so an imported version would immediately plan a replacement.

To bring an existing version under management, either publish it through this resource in the first place, or manage the package's metadata with `credible_package` (which does support import) and leave its versions out of Terraform.
