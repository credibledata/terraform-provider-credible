#!/usr/bin/env bash
# Asserts .goreleaser.yml declares the registry manifest everywhere it is needed.
#
# Registries require an asset named <project>_<version>_manifest.json that is
# also listed in SHA256SUMS. Goreleaser needs the manifest in two separate
# blocks to produce that, and each omission fails differently:
#
#   release.extra_files   - uploads the asset. Without a name_template it is
#                           uploaded under its source name, which no registry
#                           looks for, and the version indexes with a default
#                           protocol instead of the one it declares.
#   checksum.extra_files  - adds it to SHA256SUMS. Without this the Terraform
#                           registry refuses the version outright.
#
# This runs on the config, so it holds before a tag exists. A snapshot build
# cannot cover it: --snapshot skips publishing, so release.extra_files never
# executes there.
#
# Usage: scripts/check-goreleaser-config.sh [config]
set -euo pipefail

CONFIG="${1:-.goreleaser.yml}"

if [ ! -f "$CONFIG" ]; then
  echo "FAIL: $CONFIG not found" >&2
  exit 1
fi

python3 - "$CONFIG" <<'PY'
import sys

try:
    import yaml
except ImportError:
    sys.exit("SKIP: pyyaml unavailable, cannot assert goreleaser config")

path = sys.argv[1]
with open(path, encoding="utf-8") as fh:
    config = yaml.safe_load(fh)

MANIFEST = "terraform-registry-manifest.json"
EXPECTED = "{{ .ProjectName }}_{{ .Version }}_manifest.json"

problems = []

for block in ("checksum", "release"):
    entries = (config.get(block) or {}).get("extra_files") or []
    match = next(
        (e for e in entries if MANIFEST in str((e or {}).get("glob", ""))),
        None,
    )
    if match is None:
        problems.append(
            f"{block}.extra_files does not declare {MANIFEST}.\n"
            f"      Add:\n"
            f"        {block}:\n"
            f"          extra_files:\n"
            f"            - glob: ./{MANIFEST}\n"
            f'              name_template: "{EXPECTED}"'
        )
        continue

    name_template = str(match.get("name_template") or "")
    if not name_template:
        problems.append(
            f"{block}.extra_files declares {MANIFEST} without a name_template,\n"
            f"      so it is published under its source name, which registries\n"
            f'      do not look for. Add: name_template: "{EXPECTED}"'
        )
    elif "_manifest.json" not in name_template:
        problems.append(
            f"{block}.extra_files name_template is {name_template!r},\n"
            f'      which does not end in "_manifest.json" as registries require.'
        )

if problems:
    for problem in problems:
        print(f"FAIL: {problem}", file=sys.stderr)
    sys.exit(1)

print(f"OK: {path} declares the manifest in checksum.extra_files and release.extra_files")
PY
