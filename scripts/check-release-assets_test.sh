#!/usr/bin/env bash
# Tests for check-release-assets.sh.
#
# Each case builds a dist directory and asserts the checker's exit status, so a
# regression that makes the gate unable to fail is itself a failure. Case 1 is
# the exact shape the Terraform registry rejected for v0.0.6.
set -uo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHECK="$HERE/check-release-assets.sh"
# The checker falls back to this when dist has no manifest, as on a snapshot build.
export MANIFEST_SRC
# Fixtures build a single archive. The committed default (the real build matrix)
# is asserted by its own case at the end.
export EXPECTED_ARCHIVES=1
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

PROJECT="terraform-provider-credible"
VERSION="9.9.9"
MANIFEST="${PROJECT}_${VERSION}_manifest.json"
SUMS="${PROJECT}_${VERSION}_SHA256SUMS"

fails=0

# Builds a dist dir: $1 target, $2 whether to checksum the manifest
# (yes|no|wrong), $3 manifest protocol, $4 whether to emit the manifest file.
make_dist() {
  local dir="$1" mode="$2" proto="$3" with_manifest="${4:-yes}"
  rm -rf "$dir"; mkdir -p "$dir"

  local zip="${PROJECT}_${VERSION}_linux_amd64.zip"
  printf 'not-a-real-zip\n' > "$dir/$zip"
  ( cd "$dir" && shasum -a 256 "$zip" > "$SUMS" )

  # Stand in for the repo's terraform-registry-manifest.json.
  MANIFEST_SRC="$dir/source-manifest.json"
  printf '{"version":1,"metadata":{"protocol_versions":["%s"]}}\n' "$proto" > "$MANIFEST_SRC"

  if [ "$with_manifest" = "yes" ]; then
    printf '{"version":1,"metadata":{"protocol_versions":["%s"]}}\n' "$proto" > "$dir/$MANIFEST"
    case "$mode" in
      yes)   ( cd "$dir" && shasum -a 256 "$MANIFEST" >> "$SUMS" ) ;;
      wrong) printf '%064d  %s\n' 0 "$MANIFEST" >> "$dir/$SUMS" ;;
      no)    : ;;
    esac
  fi
}

expect() {
  local name="$1" want="$2" dir="$3"
  local out; out="$("$CHECK" "$dir" 2>&1)"; local got=$?
  if [ "$got" -eq "$want" ]; then
    echo "PASS  $name (exit $got)"
  else
    echo "FAIL  $name: expected exit $want, got $got"
    echo "${out//$'\n'/$'\n'        }" | sed '1s/^/        /'
    fails=$((fails + 1))
  fi
}

# The v0.0.6 bug: manifest uploaded but never checksummed.
make_dist "$WORK/unchecksummed" no 6.0
expect "manifest absent from SHA256SUMS is rejected" 1 "$WORK/unchecksummed"

make_dist "$WORK/good" yes 6.0
expect "fully checksummed release is accepted" 0 "$WORK/good"

# Pre-#11 shape: no asset matching <project>_<version>_manifest.json.
make_dist "$WORK/nomanifest" no 6.0 no
rm -f "$WORK/nomanifest/source-manifest.json"
MANIFEST_SRC="$WORK/nomanifest/source-manifest.json"
expect "missing manifest asset is rejected" 1 "$WORK/nomanifest"

# A snapshot build skips the publishing stage, so release.extra_files never
# copies the manifest into dist. checksum.extra_files still runs, so the
# manifest is listed in SHA256SUMS and the check must pass on its source copy.
snap="$WORK/snapshot"
rm -rf "$snap"; mkdir -p "$snap"
printf 'z\n' > "$snap/${PROJECT}_${VERSION}_linux_amd64.zip"
MANIFEST_SRC="$snap/source-manifest.json"
printf '{"version":1,"metadata":{"protocol_versions":["6.0"]}}\n' > "$MANIFEST_SRC"
( cd "$snap" && shasum -a 256 "${PROJECT}_${VERSION}_linux_amd64.zip" > "$SUMS" )
printf '%s  %s\n' "$(shasum -a 256 "$MANIFEST_SRC" | awk '{print $1}')" "$MANIFEST" >> "$snap/$SUMS"
expect "snapshot build without manifest in dist is accepted" 0 "$snap"

# The same snapshot shape, but the manifest was never checksummed: the v0.0.6 bug.
rm -f "$snap/$SUMS"
( cd "$snap" && shasum -a 256 "${PROJECT}_${VERSION}_linux_amd64.zip" > "$SUMS" )
expect "snapshot build with unchecksummed manifest is rejected" 1 "$snap"

# A protocol-6 provider advertising 5.0 is selectable by CLIs that cannot talk to it.
make_dist "$WORK/proto5" yes 5.0
expect "manifest declaring protocol 5.0 is rejected" 1 "$WORK/proto5"

# A stale digest would fail signature/registry verification.
make_dist "$WORK/mismatch" wrong 6.0
expect "manifest digest mismatch is rejected" 1 "$WORK/mismatch"

expect "missing dist directory is rejected" 1 "$WORK/does-not-exist"

# Dropping a platform from the build matrix must fail rather than ship quietly.
make_dist "$WORK/matrix" yes 6.0
MANIFEST_SRC="$WORK/matrix/source-manifest.json"
EXPECTED_ARCHIVES=13 \
  expect "fewer archives than the build matrix is rejected" 1 "$WORK/matrix"

echo
if [ "$fails" -gt 0 ]; then
  echo "$fails test(s) failed"
  exit 1
fi
echo "all tests passed"
