#!/usr/bin/env bash
set -Eeuo pipefail

export LC_ALL=C
version=v1.0.0
repo=github.com/gotthboard/gotth-pg-migrate
expected_go=go1.26.6
expected_bash='5.3.20(1)-release'
expected_git='git version 2.56.0'
expected_tar='tar (GNU tar) 1.35'
expected_gzip='gzip 1.15'
expected_jq='jq-1.8.2'
expected_coreutils='sha256sum (GNU coreutils) 9.12'
expected_awk='GNU Awk 5.4.1, API 4.1, PMA Avon 8-g1, (GNU MPFR 4.2.2, GNU MP 6.3.0)'
expected_find='find (GNU findutils) 4.11.0-modified'
expected_grep='grep (GNU grep) 3.12-modified'

first_line() {
  local output
  output=$("$@" 2>&1)
  printf '%s\n' "${output%%$'\n'*}"
}

[[ "$BASH_VERSION" == "$expected_bash" ]] || { echo "Bash $expected_bash is required" >&2; exit 1; }
[[ "$(go env GOVERSION)" == "$expected_go" ]] || { echo "Go $expected_go is required" >&2; exit 1; }
[[ "$(git --version)" == "$expected_git" ]] || { echo "$expected_git is required" >&2; exit 1; }
[[ "$(first_line tar --version)" == "$expected_tar" ]] || { echo "$expected_tar is required" >&2; exit 1; }
[[ "$(first_line gzip --version)" == "$expected_gzip" ]] || { echo "$expected_gzip is required" >&2; exit 1; }
[[ "$(jq --version)" == "$expected_jq" ]] || { echo "$expected_jq is required" >&2; exit 1; }
[[ "$(first_line sha256sum --version)" == "$expected_coreutils" ]] || { echo "$expected_coreutils is required" >&2; exit 1; }
[[ "$(first_line awk --version)" == "$expected_awk" ]] || { echo "$expected_awk is required" >&2; exit 1; }
[[ "$(first_line find --version)" == "$expected_find" ]] || { echo "$expected_find is required" >&2; exit 1; }
[[ "$(first_line grep --version)" == "$expected_grep" ]] || { echo "$expected_grep is required" >&2; exit 1; }

usage="usage: $0 OUTPUT_DIRECTORY COMMIT"
[[ "$#" -eq 2 ]] || { echo "$usage" >&2; exit 2; }
out=$1
commit=$2
[[ "$commit" =~ ^[0-9a-f]{40}$ ]] || { echo "COMMIT must be a full lowercase object ID" >&2; exit 2; }
[[ "$(git rev-parse "$commit^{commit}")" == "$commit" ]] || { echo "COMMIT does not identify the exact commit" >&2; exit 1; }
[[ "$(git rev-parse HEAD)" == "$commit" ]] || { echo "HEAD must equal COMMIT so the executed builder is attested" >&2; exit 1; }
[[ -z "$(git status --short)" ]] || { echo "release worktree must be clean" >&2; exit 1; }

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

git archive --format=tar --prefix="gotth-pg-migrate-$version/" --output="$root/source.tar" "$commit"
gzip -n -9 -c "$root/source.tar" >"$out/$archive"
git archive --format=tar --output="$root/source-tree.tar" "$commit"
tar -xf "$root/source-tree.tar" -C "$source_dir"
source_sha=$(sha256sum "$out/$archive" | awk '{print $1}')
printf '%s  %s\n' "$source_sha" "$archive" >"$stage/SOURCE_ARCHIVE.sha256"

(
  cd "$source_dir"
  go_sum_before=$(sha256sum go.sum | awk '{print $1}')
  go list -mod=readonly -deps -test -json ./... >"$root/packages.json.stream"
  go_sum_after=$(sha256sum go.sum | awk '{print $1}')
  [[ "$go_sum_after" == "$go_sum_before" ]] || { echo "dependency discovery changed the committed go.sum" >&2; exit 1; }
  go mod verify
)
jq -cs '[.[].Module | select(. != null)] | unique_by(.Path) | sort_by(.Path) | .[]' \
  "$root/packages.json.stream" >"$root/modules.json.stream"
jq -r 'select(.Main | not) | [.Path, .Version, .Dir] | @tsv' \
  "$root/modules.json.stream" >"$root/modules.list"
module_count=$(jq -s 'length' "$root/modules.json.stream")
dependency_count=$(awk 'NF { count++ } END { print count + 0 }' "$root/modules.list")
[[ "$module_count" -eq $((dependency_count + 1)) && "$dependency_count" -gt 0 ]] || { echo "module inventory is incomplete" >&2; exit 1; }

jq -sS --arg commit "$commit" --arg created "$created" --arg version "$version" --arg repo "$repo" '
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
' "$root/modules.json.stream" >"$stage/gotth-pg-migrate-$version.spdx.json"
jq -e --argjson modules "$module_count" --argjson dependencies "$dependency_count" '
  .spdxVersion == "SPDX-2.3" and
  (.packages | length) == $modules and
  ([.relationships[] | select(.relationshipType == "DEPENDS_ON")] | length) == $dependencies
' "$stage/gotth-pg-migrate-$version.spdx.json" >/dev/null

legal_bundle="$stage/THIRD_PARTY_LICENSES_AND_NOTICES.txt"
module_index=0
{
  printf 'Third-party license and notice evidence for gotth-pg-migrate %s\n' "$version"
  printf 'Commit: %s\n\n' "$commit"
  while read -r path module_version dir; do
    [[ -n "$path" && -n "$module_version" && -d "$dir" ]] || { echo "incomplete module directory for $path $module_version" >&2; exit 1; }
    module_index=$((module_index + 1))
    license_list="$root/licenses-$module_index.list"
    legal_list="$root/legal-$module_index.list"
    find "$dir" -maxdepth 1 -type f \( -iname LICENSE -o -iname 'LICENSE.*' -o -iname COPYING -o -iname 'COPYING.*' \) -print | sort >"$license_list"
    mapfile -t licenses <"$license_list"
    [[ ${#licenses[@]} -gt 0 ]] || { echo "no root license file for $path $module_version" >&2; exit 1; }
    find "$dir" -maxdepth 1 -type f \( -iname LICENSE -o -iname 'LICENSE.*' -o -iname COPYING -o -iname 'COPYING.*' -o -iname NOTICE -o -iname 'NOTICE.*' -o -iname PATENTS -o -iname 'PATENTS.*' \) -print | sort >"$legal_list"
    mapfile -t legal_files <"$legal_list"
    for legal_file in "${legal_files[@]}"; do
      legal_sha=$(sha256sum "$legal_file" | awk '{print $1}')
      printf '===== %s %s / %s / sha256:%s =====\n' "$path" "$module_version" "$(basename "$legal_file")" "$legal_sha"
      cat "$legal_file"
      printf '\n\n'
    done
  done <"$root/modules.list"
} >"$legal_bundle"
[[ "$module_index" -eq "$dependency_count" ]] || { echo "dependency legal-file scan is incomplete" >&2; exit 1; }
legal_dependency_count=$(grep -Eic '^===== .* / (LICENSE|COPYING)' "$legal_bundle")
[[ "$legal_dependency_count" -ge "$dependency_count" ]] || { echo "dependency license bundle is incomplete" >&2; exit 1; }

cp "$source_dir/LICENSE" "$stage/LICENSE"
cp "$source_dir/docs/release-notes-v1.0.0.md" "$stage/RELEASE_NOTES.md"
script_sha=$(sha256sum "$source_dir/scripts/build-release-candidate.sh" | awk '{print $1}')
go_sum_sha=$(sha256sum "$source_dir/go.sum" | awk '{print $1}')
jq -nS --arg commit "$commit" --arg source_created "$created" --arg archive "$archive" --arg source_sha "$source_sha" --arg script_sha "$script_sha" --arg go_sum_sha "$go_sum_sha" '
  {
    _type:"https://in-toto.io/Statement/v1",
    subject:[{name:$archive,digest:{sha256:$source_sha}}],
    predicateType:"https://github.com/gotthboard/gotth-pg-migrate/provenance/v1",
    predicate:{
      source:{uri:"git+https://github.com/gotthboard/gotth-pg-migrate.git",commit:$commit,commitTime:$source_created},
      builderContract:{
        releaseScriptSHA256:$script_sha,
        goSumSHA256:$go_sum_sha,
        tools:{bash:"5.3.20",go:"go1.26.6",git:"2.56.0",tar:"GNU 1.35",gzip:"1.15",jq:"1.8.2",coreutils:"9.12",gawk:"5.4.1",findutils:"4.11.0-modified",grep:"3.12-modified"}
      }
    }
  }
' >"$stage/provenance.intoto.json"
jq -e --arg commit "$commit" --arg source_sha "$source_sha" '
  .predicate.source.commit == $commit and .subject[0].digest.sha256 == $source_sha
' "$stage/provenance.intoto.json" >/dev/null

find "$stage" -exec touch -d "@$epoch" {} +
tar --sort=name --owner=0 --group=0 --numeric-owner --mtime="@$epoch" -C "$root" -cf "$root/metadata.tar" "$(basename "$stage")"
gzip -n -9 -c "$root/metadata.tar" >"$out/$metadata"
(
  cd "$out"
  sha256sum "$archive" "$metadata" | sort -k2 >SHA256SUMS
  sha256sum -c SHA256SUMS
)
printf 'commit=%s\nsource=%s\nmetadata=%s\n' "$commit" "$out/$archive" "$out/$metadata"
