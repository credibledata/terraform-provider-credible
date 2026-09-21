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
#
# Snapshot builds (goreleaser release --snapshot) skip the publishing stage, and
# release.extra_files is attached there, so the manifest is never copied into
# dist. Checksumming does run, so the checks that matter -- is the manifest in
# SHA256SUMS, does the digest match -- are asserted against the manifest at its
# source path instead.
set -euo pipefail

DIST="${1:-dist}"
MANIFEST_SRC="${MANIFEST_SRC:-terraform-registry-manifest.json}"
# The build matrix in .goreleaser.yml, minus the combinations goreleaser skips.
EXPECTED_ARCHIVES="${EXPECTED_ARCHIVES:-13}"

if [ ! -d "$DIST" ]; then
  echo "FAIL: dist directory '$DIST' does not exist" >&2
  exit 1
fi

sums=$(find "$DIST" -maxdepth 1 -name '*_SHA256SUMS' -print -quit)
if [ -z "$sums" ]; then
  echo "FAIL: no *_SHA256SUMS found in $DIST" >&2
  exit 1
fi

# The manifest lands in dist only on a real release; on a snapshot it exists
# only at its source path. Either way it must be the checksummed file, so
# derive the expected asset name from the SHA256SUMS name and locate a body.
manifest_name="$(basename "$sums" _SHA256SUMS)_manifest.json"
manifest=$(find "$DIST" -maxdepth 1 -name "$manifest_name" -print -quit)
if [ -z "$manifest" ] && [ -f "$MANIFEST_SRC" ]; then
  manifest="$MANIFEST_SRC"
fi
if [ -z "$manifest" ]; then
  cat >&2 <<MSG
FAIL: cannot find $manifest_name in $DIST, nor $MANIFEST_SRC to check.
      Registries look for the <project>_<version>_manifest.json name. Declare
      the manifest under release.extra_files with a name_template.
MSG
  exit 1
fi
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

# Pin the platform count so dropping a goos/goarch from the build matrix fails
# here instead of silently shipping fewer platforms. Update EXPECTED_ARCHIVES
# deliberately when the matrix changes.
archives=$(grep -c '\.zip$' "$sums" || true)
if [ "$archives" -ne "$EXPECTED_ARCHIVES" ]; then
  echo "FAIL: $(basename "$sums") lists $archives archives, expected $EXPECTED_ARCHIVES." >&2
  echo "      A platform was added or dropped from the build matrix. If intended," >&2
  echo "      update EXPECTED_ARCHIVES in $(basename "$0")." >&2
  exit 1
fi

echo "OK: ${manifest_name} checksummed (${expected:0:12}...), protocol 6.0, ${archives} archives"
