#!/usr/bin/env bash
# Asserts that a built release carries every asset the Terraform and OpenTofu
# registries require, and that the manifest is listed in SHA256SUMS.
#
# The manifest is the trap: goreleaser checksums build artifacts (the archives),
# while release.extra_files attaches files after checksumming. Declaring the
# manifest only under release.extra_files uploads it unchecksummed, which the
# Terraform registry rejects with "missing SHA256 checksum for [...]".
#
# Usage: scripts/check-release-assets.sh <dist-dir>
set -euo pipefail

DIST="${1:-dist}"

if [ ! -d "$DIST" ]; then
  echo "FAIL: dist directory '$DIST' does not exist" >&2
  exit 1
fi

sums=$(find "$DIST" -maxdepth 1 -name '*_SHA256SUMS' -print -quit)
if [ -z "$sums" ]; then
  echo "FAIL: no *_SHA256SUMS found in $DIST" >&2
  exit 1
fi

manifest=$(find "$DIST" -maxdepth 1 -name '*_manifest.json' -print -quit)
if [ -z "$manifest" ]; then
  cat >&2 <<'MSG'
FAIL: no <project>_<version>_manifest.json in dist.
      Registries look for that exact name. Declare it under release.extra_files
      with a name_template.
MSG
  exit 1
fi

# The registry matches on the asset's basename, so compare basenames.
manifest_name=$(basename "$manifest")
if ! grep -qF "  ${manifest_name}" "$sums"; then
  cat >&2 <<MSG
FAIL: ${manifest_name} is not listed in $(basename "$sums").
      The Terraform registry refuses the version with:
        missing SHA256 checksum for ["${manifest_name}"]
      Add the manifest to checksum.extra_files in .goreleaser.yml (it must be
      declared in BOTH checksum.extra_files and release.extra_files).
MSG
  exit 1
fi

# A protocol-6 provider that advertises 5.0 is selectable by CLIs that cannot
# talk to it, so assert the manifest says what the provider actually speaks.
if ! grep -q '"6.0"' "$manifest"; then
  echo "FAIL: ${manifest_name} does not declare protocol 6.0:" >&2
  cat "$manifest" >&2
  exit 1
fi

# Verify the recorded digest matches the file actually being shipped.
expected=$(grep -F "  ${manifest_name}" "$sums" | awk '{print $1}')
actual=$(shasum -a 256 "$manifest" | awk '{print $1}')
if [ "$expected" != "$actual" ]; then
  echo "FAIL: ${manifest_name} digest mismatch (sums=$expected actual=$actual)" >&2
  exit 1
fi

archives=$(grep -c '\.zip$' "$sums" || true)
if [ "$archives" -lt 1 ]; then
  echo "FAIL: no .zip archives listed in $(basename "$sums")" >&2
  exit 1
fi

echo "OK: ${manifest_name} checksummed (${expected:0:12}...), protocol 6.0, ${archives} archives"
