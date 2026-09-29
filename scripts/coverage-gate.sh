#!/usr/bin/env bash
#
# The gate. One command decides whether the project passes.
#
# `nix flake check` and .github/workflows/gate.yml both call this script and
# restate none of it, so a pass in one place means the same as a pass in the
# other.
#
# The floors below are the coverage measured when each was last set. They track
# what is achieved and rise as tested code lands. Lowering one is allowed only
# as a deliberate edit here, with the reason written beside it; it must never
# happen as a side effect of adding untested code.
#
# When raising a floor, check both a local run and `nix flake check`. Coverage
# is not always identical in the two, and a floor has to track the lowest
# reading or the gate will disagree with itself.
#
# Run it by hand with: bash scripts/coverage-gate.sh

set -euo pipefail

module="github.com/speclib/openspec-schema-manager"

floor_for() {
  case "$1" in
    # Everything but main() is reachable from a test, because run() takes
    # writers rather than files. main() is the one uncovered function.
    "$module/cmd/ossm")          echo "96.0" ;;
    "$module/internal/compose")  echo "" ;;
    "$module/internal/config")   echo "" ;;
    "$module/internal/graph")    echo "" ;;
    "$module/internal/openspec") echo "" ;;
    "$module/internal/registry") echo "" ;;
    "$module/internal/schema")   echo "" ;;
    "$module/internal/source")   echo "" ;;
    "$module/internal/tui")      echo "" ;;
    TOTAL)                       echo "96.0" ;;
    *)                           echo "" ;;
  esac
}

profile="$(mktemp -t ossm-cover.XXXXXX)"
trap 'rm -f "$profile"' EXIT

echo "==> go vet"
go vet ./...

echo
echo "==> go test with coverage"
if ! test_output="$(go test -coverprofile="$profile" -covermode=set ./... 2>&1)"; then
  echo "$test_output"
  echo
  echo "Tests FAILED. Fix them before the coverage floors are even considered."
  exit 1
fi
echo "$test_output"

status=0

check() {
  local label="$1" actual="$2" floor="$3"
  if [ -z "$floor" ]; then
    printf '  %-55s %6s%%   (no floor yet)\n' "$label" "$actual"
    return
  fi
  if awk "BEGIN{exit !($actual >= $floor)}"; then
    printf '  %-55s %6s%%   >= %s%% ok\n' "$label" "$actual" "$floor"
  else
    printf '  %-55s %6s%%   <  %s%% FAILED\n' "$label" "$actual" "$floor"
    status=1
  fi
}

echo
echo "==> coverage floors"

while read -r pkg cov; do
  [ -n "$pkg" ] || continue
  check "$pkg" "$cov" "$(floor_for "$pkg")"
done < <(echo "$test_output" \
  | sed -n 's|^ok[[:space:]]\{1,\}\([^[:space:]]\{1,\}\).*coverage: \([0-9.]\{1,\}\)% of statements.*|\1 \2|p')

total="$(go tool cover -func="$profile" | awk '$1=="total:"{sub(/%/,"",$3); print $3}')"
check "TOTAL" "$total" "$(floor_for TOTAL)"

echo
if [ "$status" -ne 0 ]; then
  echo "Coverage gate FAILED. Add tests, or lower a floor here on purpose and say why."
  exit 1
fi
echo "Gate passed: vet clean, tests green, coverage at or above every floor."
