#!/usr/bin/env bash
#
# Builds everything a recording needs, under demo/.work, so a demo needs no
# network and produces the same thing twice.
#
# Run it before recording:  bash demo/setup.sh

set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
work="$here/.work"

rm -rf "$work"
mkdir -p "$work"/{config/ossm,cache,state,schemas,home}

schema() {
  local name="$1" description="$2"
  shift 2

  local dir="$work/schemas/$name"
  mkdir -p "$dir/templates"

  {
    echo "name: $name"
    echo "version: 1"
    echo "description: $description"
    echo "artifacts:"
    local previous=""
    for artifact in "$@"; do
      echo "  - id: $artifact"
      echo "    generates: $artifact.md"
      echo "    description: what $artifact produces"
      echo "    template: $artifact.md"
      if [ -n "$previous" ]; then
        echo "    requires: [$previous]"
      fi
      printf '# %s\n\nWrite the %s here.\n' "$artifact" "$artifact" > "$dir/templates/$artifact.md"
      previous="$artifact"
    done
    echo "apply:"
    echo "  requires: [$previous]"
    echo "  tracks: $previous.md"
  } > "$dir/schema.yaml"
}

schema minimalist "Lightweight schema for well-scoped, low-risk changes" specs tasks
schema research-first "Research before proposing" research proposal tasks
schema team-review "A review gate between proposal and specs" proposal review specs tasks

# A git repository the registry can point at, so install and fetch are real.
repo="$work/registry-source"
mkdir -p "$repo/openspec/schemas"
cp -r "$work/schemas/research-first" "$repo/openspec/schemas/"
cp -r "$work/schemas/team-review" "$repo/openspec/schemas/"

git -c init.defaultBranch=main init --quiet "$repo"
git -C "$repo" -c user.name=demo -c user.email=demo@example.test add -A
git -C "$repo" -c user.name=demo -c user.email=demo@example.test commit --quiet -m "the schemas"

cat > "$work/openspec-schemas.json" <<JSON
{
  "schemas": [
    {
      "id": "speclib/research-first",
      "name": "research-first",
      "description": "Research before proposing, then tasks.",
      "artifacts": ["research", "proposal", "tasks"],
      "source": { "repo": "$repo", "path": "openspec/schemas/research-first" }
    },
    {
      "id": "speclib/team-review",
      "name": "team-review",
      "description": "A review gate between proposal and specs.",
      "artifacts": ["proposal", "review", "specs", "tasks"],
      "source": { "repo": "$repo", "path": "openspec/schemas/team-review", "ref": "main" }
    }
  ]
}
JSON

cat > "$work/config/ossm/config.yml" <<YAML
registry_url: file://$work/openspec-schemas.json
registry_ttl: 0s
schemas_dirs:
  - $work/schemas
YAML

# A demo project, so the Project tab and installing have something to show.
project="$work/demo-app"
mkdir -p "$project"
( cd "$project" && HOME="$work/home" openspec init --tools none --no-animation >/dev/null 2>&1 )
( cd "$project" && HOME="$work/home" openspec new change add-auth --description "Sign in with a passkey" >/dev/null 2>&1 )
( cd "$project" && HOME="$work/home" openspec new change fix-export --description "Export takes the whole table" >/dev/null 2>&1 )

echo "demo fixtures ready in $work"
