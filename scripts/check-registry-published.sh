#!/usr/bin/env bash
# Verifies a released version was actually ingested by the Terraform and
# OpenTofu registries, and is advertising the right plugin protocol.
#
# Ingestion is asynchronous and slower than the registry's own published-at
# timestamp suggests: that field records the release time, not when the version
# becomes queryable. Measured on v0.0.7 (released 16:12:45Z):
#
#   terraform  absent 16:18:00Z -> present 16:19:01Z   (~6m)
#   opentofu   absent 16:28:04Z -> present 16:29:04Z   (~16m)
#
# The default window is sized well past both, and each registry is polled on
# its own clock so a slow one does not eat the other's budget. A version still
# absent at the end has been refused; the Terraform registry surfaces the
# reason only in its UI, under Provider version history.
#
# Usage: scripts/check-registry-published.sh <version> [attempts] [sleep-seconds]
#        version may be given as "1.2.3" or "v1.2.3"
set -uo pipefail

VERSION="${1:?usage: check-registry-published.sh <version> [attempts] [sleep]}"
VERSION="${VERSION#v}"
ATTEMPTS="${2:-60}"
SLEEP="${3:-30}"

NAMESPACE="credibledata"
NAME="credible"

# Echoes the protocols for $VERSION, or nothing if the version is absent.
probe() {
  curl -sf --max-time 20 "$1/v1/providers/$NAMESPACE/$NAME/versions" 2>/dev/null \
    | python3 -c '
import json, sys
try:
    doc = json.load(sys.stdin)
except Exception:
    sys.exit(0)
for v in doc.get("versions", []):
    if v.get("version") == sys.argv[1]:
        print(",".join(v.get("protocols") or []))
        break
' "$VERSION"
}

# Polls one registry until $VERSION appears. Prints its protocols on success.
await() {
  local label="$1" base="$2" i protocols
  for ((i = 1; i <= ATTEMPTS; i++)); do
    protocols="$(probe "$base")"
    if [ -n "$protocols" ]; then
      echo "$label: $VERSION present, protocols [$protocols]"
      [ "$protocols" = "6.0" ] && return 0
      echo "$label: expected protocol 6.0, got [$protocols]" >&2
      return 1
    fi
    [ "$i" -lt "$ATTEMPTS" ] && sleep "$SLEEP"
  done

  echo "$label: $VERSION did not appear after $((ATTEMPTS * SLEEP))s" >&2
  return 1
}

# Poll concurrently: sequential awaits would leave the second registry only
# whatever time the first did not use.
rc=0
await "terraform" "https://registry.terraform.io" & tf=$!
await "opentofu"  "https://registry.opentofu.org" & tofu=$!
wait "$tf"   || rc=1
wait "$tofu" || rc=1

if [ "$rc" -ne 0 ]; then
  cat >&2 <<MSG

A version that never appears was refused during ingestion. Check the reason at
  https://registry.terraform.io/providers/$NAMESPACE/$NAME
  (sign in as an org owner, then Manage provider -> Provider version history)
The usual cause is an asset the registry requires but the release did not carry;
scripts/check-release-assets.sh asserts that set before a tag is cut.
MSG
fi
exit "$rc"
