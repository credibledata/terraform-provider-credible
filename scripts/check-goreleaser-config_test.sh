#!/usr/bin/env bash
# Tests for check-goreleaser-config.sh, including the two historical configs
# that produced a release the Terraform registry would not index.
set -uo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHECK="$HERE/check-goreleaser-config.sh"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

fails=0
MANIFEST_ENTRY='    - glob: ./terraform-registry-manifest.json
      name_template: "{{ .ProjectName }}_{{ .Version }}_manifest.json"'

# write_config <file> <checksum-entry> <release-entry>
write_config() {
  local f="$1" ck="$2" rel="$3"
  {
    echo "version: 2"
    echo "checksum:"
    [ -n "$ck" ] && echo "  extra_files:" && echo "$ck"
    echo '  name_template: "{{ .ProjectName }}_{{ .Version }}_SHA256SUMS"'
    echo "release:"
    [ -n "$rel" ] && echo "  extra_files:" && echo "$rel"
  } > "$f"
}

expect() {
  local name="$1" want="$2" cfg="$3"
  local out; out="$("$CHECK" "$cfg" 2>&1)"; local got=$?
  if [ "$got" -eq "$want" ]; then
    echo "PASS  $name (exit $got)"
  else
    echo "FAIL  $name: expected exit $want, got $got"
    echo "$out" | sed '1s/^/        /'
    fails=$((fails + 1))
  fi
}

write_config "$WORK/good.yml" "$MANIFEST_ENTRY" "$MANIFEST_ENTRY"
expect "manifest in both blocks is accepted" 0 "$WORK/good.yml"

# The config that left v0.0.6 unindexed: uploaded but not checksummed.
write_config "$WORK/no-checksum.yml" "" "$MANIFEST_ENTRY"
expect "manifest missing from checksum.extra_files is rejected" 1 "$WORK/no-checksum.yml"

write_config "$WORK/no-release.yml" "$MANIFEST_ENTRY" ""
expect "manifest missing from release.extra_files is rejected" 1 "$WORK/no-release.yml"

# The pre-rename config: uploaded under a name no registry looks for.
write_config "$WORK/no-template.yml" "$MANIFEST_ENTRY" \
  '    - glob: ./terraform-registry-manifest.json'
expect "release entry without a name_template is rejected" 1 "$WORK/no-template.yml"

write_config "$WORK/bad-template.yml" "$MANIFEST_ENTRY" \
  '    - glob: ./terraform-registry-manifest.json
      name_template: "manifest.txt"'
expect "name_template not ending in _manifest.json is rejected" 1 "$WORK/bad-template.yml"

expect "missing config file is rejected" 1 "$WORK/absent.yml"

# The config in the repo must pass.
expect "the committed .goreleaser.yml is accepted" 0 "$HERE/../.goreleaser.yml"

echo
if [ "$fails" -gt 0 ]; then
  echo "$fails test(s) failed"
  exit 1
fi
echo "all tests passed"
