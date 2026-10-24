#!/usr/bin/env bash
set -Eeuo pipefail

export LC_ALL=C GOFLAGS= GOWORK=off GOENV=off GIT_NO_REPLACE_OBJECTS=1 GIT_ATTR_NOSYSTEM=1
export TAR_OPTIONS= GZIP= GOPROXY=https://proxy.golang.org GOSUMDB=sum.golang.org GOPRIVATE= GONOSUMDB= GONOPROXY=
export GIT_CONFIG_NOSYSTEM=1 GIT_CONFIG_GLOBAL=/dev/null
unset GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE GIT_OBJECT_DIRECTORY GIT_ALTERNATE_OBJECT_DIRECTORIES GIT_COMMON_DIR GIT_TEMPLATE_DIR
umask 0022
version=v1.0.0
repo=github.com/gotthboard/gotth-pg-migrate
expected_go=go1.26.6
expected_bash='5.3.20(1)-release'
expected_git='git version 2.56.0'
expected_tar='tar (GNU tar) 1.35'
expected_gzip='gzip 1.15'
expected_jq='jq-1.8.2'
expected_coreutils='sha256sum (GNU coreutils) 9.12'
expected_mv='mv (GNU coreutils) 9.12'
expected_awk='GNU Awk 5.4.1, API 4.1, PMA Avon 8-g1, (GNU MPFR 4.2.2, GNU MP 6.3.0)'
expected_find='find (GNU findutils) 4.11.0-modified'
expected_grep='grep (GNU grep) 3.12-modified'

first_line() {
  local output
  output=$("$@" 2>&1)
  printf '%s\n' "${output%%$'\n'*}"
}

[[ "$BASH_VERSION" == "$expected_bash" ]] || { echo "Bash $expected_bash is required" >&2; exit 1; }
selected_go_root=$(go env GOROOT)
[[ "$("$selected_go_root/bin/go" env GOVERSION)" == "$expected_go" ]] || { echo "Go $expected_go is required" >&2; exit 1; }
export PATH="$selected_go_root/bin:$PATH" GOTOOLCHAIN=local
[[ "$(git --version)" == "$expected_git" ]] || { echo "$expected_git is required" >&2; exit 1; }
[[ "$(first_line tar --version)" == "$expected_tar" ]] || { echo "$expected_tar is required" >&2; exit 1; }
[[ "$(first_line gzip --version)" == "$expected_gzip" ]] || { echo "$expected_gzip is required" >&2; exit 1; }
[[ "$(jq --version)" == "$expected_jq" ]] || { echo "$expected_jq is required" >&2; exit 1; }
[[ "$(first_line sha256sum --version)" == "$expected_coreutils" ]] || { echo "$expected_coreutils is required" >&2; exit 1; }
[[ "$(first_line mv --version)" == "$expected_mv" ]] || { echo "$expected_mv is required" >&2; exit 1; }
[[ "$(first_line awk --version)" == "$expected_awk" ]] || { echo "$expected_awk is required" >&2; exit 1; }
[[ "$(first_line find --version)" == "$expected_find" ]] || { echo "$expected_find is required" >&2; exit 1; }
[[ "$(first_line grep --version)" == "$expected_grep" ]] || { echo "$expected_grep is required" >&2; exit 1; }

usage="usage: $0 OUTPUT_DIRECTORY COMMIT"
[[ "$#" -eq 2 ]] || { echo "$usage" >&2; exit 2; }
requested_out=${1%/}
commit=$2
[[ -n "$requested_out" ]] || { echo "OUTPUT_DIRECTORY must not be empty" >&2; exit 2; }
[[ "$commit" =~ ^[0-9a-f]{40}$ ]] || { echo "COMMIT must be a full lowercase object ID" >&2; exit 2; }
[[ "$(git rev-parse "$commit^{commit}")" == "$commit" ]] || { echo "COMMIT does not identify the exact commit" >&2; exit 1; }
[[ "$(git rev-parse HEAD)" == "$commit" ]] || { echo "HEAD must equal COMMIT so the executed builder is attested" >&2; exit 1; }
[[ -z "$(git status --short)" ]] || { echo "release worktree must be clean" >&2; exit 1; }
repo_root=$(git rev-parse --show-toplevel)
running_script=$(realpath "${BASH_SOURCE[0]}")
running_script_sha=$(sha256sum "$running_script" | awk '{print $1}')

out_parent=$(dirname "$requested_out")
out_name=$(basename "$requested_out")
mkdir -p "$out_parent"
out_parent=$(cd "$out_parent" && pwd)
final_out="$out_parent/$out_name"
[[ ! -e "$final_out" ]] || { echo "OUTPUT_DIRECTORY must not already exist" >&2; exit 1; }
publish_stage=$(mktemp -d "$out_parent/.${out_name}.tmp.XXXXXX")
root=$(mktemp -d)
source_git="$root/source.git"
empty_git_template="$root/empty-git-template"
mkdir -p "$empty_git_template"
export GOMODCACHE="$root/gomodcache" GOCACHE="$root/gocache"
git init --bare -q --template="$empty_git_template" "$source_git"
git --git-dir="$source_git" fetch -q --no-tags "$repo_root" "$commit"
[[ "$(git --git-dir="$source_git" rev-parse FETCH_HEAD)" == "$commit" ]] || { echo "isolated source fetch did not preserve COMMIT" >&2; exit 1; }
committed_script_sha=$(git --git-dir="$source_git" show "$commit:scripts/build-release-candidate.sh" | sha256sum | awk '{print $1}')
[[ "$running_script_sha" == "$committed_script_sha" ]] || { echo "executing builder does not match COMMIT" >&2; exit 1; }
cleanup() {
  chmod -R u+w "$root" 2>/dev/null || true
  rm -rf "$root" || true
  if [[ -n "$publish_stage" ]]; then
    rm -rf "$publish_stage" || true
  fi
}
trap cleanup EXIT
trap 'exit 130' HUP INT TERM
out=$publish_stage
stage="$root/gotth-pg-migrate-$version-metadata"
source_dir="$root/gotth-pg-migrate-$version"
mkdir -p "$stage" "$source_dir"
archive="gotth-pg-migrate-$version-source.tar.gz"
metadata="gotth-pg-migrate-$version-metadata.tar.gz"
epoch=$(git --git-dir="$source_git" show -s --format=%ct "$commit")
created=$(date -u -d "@$epoch" '+%Y-%m-%dT%H:%M:%SZ')

git --git-dir="$source_git" -c core.attributesFile=/dev/null -c tar.umask=0022 archive --format=tar --prefix="gotth-pg-migrate-$version/" --output="$root/source.tar" "$commit"
gzip -n -9 -c "$root/source.tar" >"$out/$archive"
git --git-dir="$source_git" -c core.attributesFile=/dev/null -c tar.umask=0022 archive --format=tar --output="$root/source-tree.tar" "$commit"
tar -xf "$root/source-tree.tar" -C "$source_dir"
git --git-dir="$source_git" ls-tree -rz "$commit" >"$root/source-tree.list"
tree_entry_count=0
while IFS= read -r -d '' entry; do
  tree_metadata=${entry%%$'\t'*}
  path=${entry#*$'\t'}
  read -r mode object_type object_id <<<"$tree_metadata"
  [[ "$object_type" == blob ]] || { echo "source tree contains unsupported object type: $path" >&2; exit 1; }
  case "$mode" in
    100644) [[ -f "$source_dir/$path" && ! -L "$source_dir/$path" && ! -x "$source_dir/$path" ]] || { echo "source mode mismatch: $path" >&2; exit 1; } ;;
    100755) [[ -f "$source_dir/$path" && ! -L "$source_dir/$path" && -x "$source_dir/$path" ]] || { echo "source mode mismatch: $path" >&2; exit 1; } ;;
    120000) [[ -L "$source_dir/$path" ]] || { echo "source symlink mismatch: $path" >&2; exit 1; } ;;
    *) echo "unsupported source mode $mode: $path" >&2; exit 1 ;;
  esac
  if [[ "$mode" == 120000 ]]; then
    actual_id=$(readlink "$source_dir/$path" | git --git-dir="$source_git" hash-object --stdin)
  else
    actual_id=$(git --git-dir="$source_git" hash-object "$source_dir/$path")
  fi
  [[ "$actual_id" == "$object_id" ]] || { echo "source content mismatch: $path" >&2; exit 1; }
  tree_entry_count=$((tree_entry_count + 1))
done <"$root/source-tree.list"
find "$source_dir" \( -type f -o -type l \) -print0 >"$root/source-files.list"
source_entry_count=0
while IFS= read -r -d '' _; do source_entry_count=$((source_entry_count + 1)); done <"$root/source-files.list"
[[ "$source_entry_count" -eq "$tree_entry_count" ]] || { echo "source archive member count does not match COMMIT" >&2; exit 1; }
source_sha=$(sha256sum "$out/$archive" | awk '{print $1}')
printf '%s  %s\n' "$source_sha" "$archive" >"$stage/SOURCE_ARCHIVE.sha256"

(
  cd "$source_dir"
  go mod edit -json >"$root/go-mod.json"
  jq -e '(.Replace // []) | length == 0' "$root/go-mod.json" >/dev/null || { echo "go.mod replacements are not allowed" >&2; exit 1; }
  go_sum_before=$(sha256sum go.sum | awk '{print $1}')
  go list -mod=readonly -deps -test -json ./... >"$root/packages.json.stream"
  go_sum_after=$(sha256sum go.sum | awk '{print $1}')
  [[ "$go_sum_after" == "$go_sum_before" ]] || { echo "dependency discovery changed the committed go.sum" >&2; exit 1; }
  go mod verify
)
jq -e -s 'all(.[].Module | select(. != null); .Replace == null)' \
  "$root/packages.json.stream" >/dev/null || { echo "module replacements are not allowed" >&2; exit 1; }
jq -cs '[.[].Module | select(. != null)] | unique_by(.Path) | sort_by(.Path) | .[]' \
  "$root/packages.json.stream" >"$root/modules.json.stream"
jq -e -s 'all(.[] | select(.Main | not); (.Sum | type == "string") and (.Sum | length > 0) and (.GoModSum | type == "string") and (.GoModSum | length > 0))' \
  "$root/modules.json.stream" >/dev/null || { echo "dependency checksums are incomplete" >&2; exit 1; }
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

find "$stage" -type d -exec chmod 0755 {} +
find "$stage" -type f -exec chmod 0644 {} +
find "$stage" -exec touch -d "@$epoch" {} +
tar --sort=name --owner=0 --group=0 --numeric-owner --mtime="@$epoch" -C "$root" -cf "$root/metadata.tar" "$(basename "$stage")"
metadata_root=$(basename "$stage")
printf '%s\n' \
  "$metadata_root/" \
  "$metadata_root/LICENSE" \
  "$metadata_root/RELEASE_NOTES.md" \
  "$metadata_root/SOURCE_ARCHIVE.sha256" \
  "$metadata_root/THIRD_PARTY_LICENSES_AND_NOTICES.txt" \
  "$metadata_root/gotth-pg-migrate-$version.spdx.json" \
  "$metadata_root/provenance.intoto.json" >"$root/metadata-expected.list"
tar -tf "$root/metadata.tar" | sort >"$root/metadata-actual.list"
[[ "$(cat "$root/metadata-actual.list")" == "$(cat "$root/metadata-expected.list")" ]] || { echo "metadata archive members are incomplete" >&2; exit 1; }
gzip -n -9 -c "$root/metadata.tar" >"$out/$metadata"
gzip -t "$out/$metadata"
(
  cd "$out"
  sha256sum "$archive" "$metadata" | sort -k2 >SHA256SUMS
  sha256sum -c SHA256SUMS
  chmod 0644 "$archive" "$metadata" SHA256SUMS
)
chmod 0755 "$publish_stage"
mv --no-clobber --no-target-directory "$publish_stage" "$final_out"
[[ ! -e "$publish_stage" && -f "$final_out/$archive" && -f "$final_out/$metadata" && -f "$final_out/SHA256SUMS" ]] || { echo "OUTPUT_DIRECTORY appeared during publication" >&2; exit 1; }
publish_stage=
printf 'commit=%s\nsource=%s\nmetadata=%s\n' "$commit" "$final_out/$archive" "$final_out/$metadata"
