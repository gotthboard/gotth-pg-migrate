#!/bin/sh
set -eu

export LC_ALL=C
version=v1.0.0
repo=github.com/gotthboard/gotth-pg-migrate
expected_go=go1.26.6
actual_go=$(go env GOVERSION)
[ "${actual_go%%-*}" = "$expected_go" ] || { echo "Go $expected_go is required" >&2; exit 1; }
usage="usage: $0 OUTPUT_DIRECTORY COMMIT"
[ "$#" -eq 2 ] || { echo "$usage" >&2; exit 2; }
out=$1
commit=$2
case "$commit" in *[!0-9a-f]*|'') echo "COMMIT must be a full lowercase object ID" >&2; exit 2;; esac
[ "${#commit}" -eq 40 ] || { echo "COMMIT must contain 40 hexadecimal characters" >&2; exit 2; }
[ "$(git rev-parse "$commit^{commit}")" = "$commit" ] || { echo "COMMIT does not identify the exact commit" >&2; exit 1; }
[ -z "$(git status --short)" ] || { echo "release worktree must be clean" >&2; exit 1; }

root=$(mktemp -d)
trap 'rm -rf "$root"' EXIT HUP INT TERM
stage="$root/gotth-pg-migrate-$version-metadata"
source_dir="$root/gotth-pg-migrate-$version"
mkdir -p "$out" "$stage" "$source_dir"
out=$(cd "$out" && pwd)
archive="gotth-pg-migrate-$version-source.tar.gz"
metadata="gotth-pg-migrate-$version-metadata.tar.gz"
epoch=$(git show -s --format=%ct "$commit")
created=$(date -u -d "@$epoch" '+%Y-%m-%dT%H:%M:%SZ')

git archive --format=tar --prefix="gotth-pg-migrate-$version/" "$commit" | gzip -n -9 >"$out/$archive"
git archive --format=tar "$commit" | tar -xf - -C "$source_dir"
source_sha=$(sha256sum "$out/$archive" | awk '{print $1}')
printf '%s  %s\n' "$source_sha" "$archive" >"$stage/SOURCE_ARCHIVE.sha256"

# Populate every go.sum module in the verified module cache so dependency
# license evidence covers build and test transitive modules alike.
(
  cd "$source_dir"
  go mod download all
)

(
  cd "$source_dir"
  go list -mod=readonly -m -json all
) | jq -sS --arg commit "$commit" --arg created "$created" --arg version "$version" --arg repo "$repo" '
  def package_id: "SPDXRef-Package-" + ((.key + 1) | tostring);
  . as $modules |
  {
    spdxVersion: "SPDX-2.3", dataLicense: "CC0-1.0", SPDXID: "SPDXRef-DOCUMENT",
    name: ("gotth-pg-migrate-" + $version),
    documentNamespace: ("https://" + $repo + "/releases/" + $version + "/spdx/" + $commit),
    creationInfo: {created: $created, creators: ["Tool: scripts/build-release-candidate.sh"]},
    packages: ($modules | to_entries | map(. as $entry | .value as $module | {
      SPDXID: ($entry | package_id), name: $module.Path,
      versionInfo: (if $module.Main then $version else $module.Version end),
      downloadLocation: (if $module.Main then "NOASSERTION" else ("https://proxy.golang.org/" + $module.Path + "/@v/" + $module.Version + ".zip") end),
      filesAnalyzed: false, licenseConcluded: "NOASSERTION",
      licenseDeclared: (if $module.Main then "MIT" else "NOASSERTION" end),
      copyrightText: "NOASSERTION",
      externalRefs: [{referenceCategory:"PACKAGE-MANAGER", referenceType:"purl", referenceLocator:("pkg:golang/" + $module.Path + "@" + (if $module.Main then $version else $module.Version end))}]
    })),
    relationships: ([{spdxElementId:"SPDXRef-DOCUMENT", relationshipType:"DESCRIBES", relatedSpdxElement:"SPDXRef-Package-1"}] +
      ($modules | to_entries | map(select(.value.Main|not) | {spdxElementId:"SPDXRef-Package-1", relationshipType:"DEPENDS_ON", relatedSpdxElement:(. | package_id)})))
  }
' >"$stage/gotth-pg-migrate-$version.spdx.json"

{
  printf 'Third-party license evidence for gotth-pg-migrate %s\n' "$version"
  printf 'Commit: %s\n\n' "$commit"
  (
    cd "$source_dir"
    go list -mod=readonly -m -f '{{if not .Main}}{{.Path}} {{.Version}} {{.Dir}}{{end}}' all
  ) | sort | while read -r path module_version dir; do
    [ -n "$path" ] || continue
    licenses=$(find "$dir" -maxdepth 1 -type f \( -iname LICENSE -o -iname 'LICENSE.*' -o -iname COPYING -o -iname 'COPYING.*' \) -print | sort)
    [ -n "$licenses" ] || { echo "no root license file for $path $module_version" >&2; exit 1; }
    for license in $licenses; do
      license_sha=$(sha256sum "$license" | awk '{print $1}')
      printf '===== %s %s / %s / sha256:%s =====\n' "$path" "$module_version" "$(basename "$license")" "$license_sha"
      cat "$license"
      printf '\n\n'
    done
  done
} >"$stage/THIRD_PARTY_LICENSES.txt"

cp "$source_dir/LICENSE" "$stage/LICENSE"
cp "$source_dir/docs/release-notes-v1.0.0.md" "$stage/RELEASE_NOTES.md"
script_sha=$(sha256sum "$source_dir/scripts/build-release-candidate.sh" | awk '{print $1}')
jq -nS --arg commit "$commit" --arg created "$created" --arg archive "$archive" --arg source_sha "$source_sha" --arg script_sha "$script_sha" '
  {
    _type:"https://in-toto.io/Statement/v1",
    subject:[{name:$archive,digest:{sha256:$source_sha}}],
    predicateType:"https://slsa.dev/provenance/v1",
    predicate:{
      buildDefinition:{
        buildType:"https://github.com/gotthboard/gotth-pg-migrate/release-candidate/v1",
        externalParameters:{version:"v1.0.0",commit:$commit},
        internalParameters:{releaseScriptSHA256:$script_sha,goToolchain:"go1.26.6"},
        resolvedDependencies:[{uri:"git+https://github.com/gotthboard/gotth-pg-migrate.git",digest:{gitCommit:$commit}}]
      },
      runDetails:{builder:{id:"https://github.com/gotthboard/gotth-pg-migrate/scripts/build-release-candidate.sh"},metadata:{invocationId:$commit,startedOn:$created,finishedOn:$created}}
    }
  }
' >"$stage/provenance.intoto.jsonl"

find "$stage" -exec touch -d "@$epoch" {} +
tar --sort=name --owner=0 --group=0 --numeric-owner --mtime="@$epoch" -C "$root" -cf - "$(basename "$stage")" | gzip -n -9 >"$out/$metadata"
(
  cd "$out"
  sha256sum "$archive" "$metadata" | sort -k2 >SHA256SUMS
  sha256sum -c SHA256SUMS
)
printf 'commit=%s\nsource=%s\nmetadata=%s\n' "$commit" "$out/$archive" "$out/$metadata"
