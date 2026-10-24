# Release provenance and reproducibility

The v1 candidate is source-only. Its second archive contains release metadata,
not a compiled binary.

`scripts/build-release-candidate.sh` accepts one clean full commit, requires
that commit to equal `HEAD`, disables Git replacement objects, and requires the
executing script bytes to match the script in that commit. It fetches the raw
commit into a fresh bare repository with an explicit empty template before
archiving, so ambient repository, replacement objects, templates, and
global/system attributes cannot change release bytes. It then compares every
archived member, mode, and blob ID with the raw commit tree. It emits:

- a deterministic `git archive` source tarball;
- an SPDX 2.3 module inventory;
- an in-toto source-provenance statement using the project predicate;
- the owner MIT license and dependency `LICENSE`, `COPYING`, `NOTICE`, and
  `PATENTS` files with individual SHA-256 values;
- a deterministic metadata tarball and `SHA256SUMS`.

The SPDX inventory deliberately records third-party license declarations as
`NOASSERTION`; the bundled upstream legal files are the review evidence. This
avoids inventing license classifications.

## Builder contract

Byte reproduction requires the exact tool contract enforced by the script:
Bash 5.3.20, Go 1.26.6, Git 2.56.0, GNU tar 1.35, gzip 1.15, jq
1.8.2, GNU `sha256sum` and `mv` from coreutils 9.12, GNU awk 5.4.1,
GNU findutils
4.11.0-modified, and GNU grep 3.12-modified. The source commit timestamp
supplies archive mtimes; gzip stores no name or time. `SOURCE_DATE_EPOCH` is a
required recorded input representing the original release-document creation
time; exact reproductions reuse it, so SPDX `creationInfo.created` remains both
truthful and byte-stable. The builder clears
`TAR_OPTIONS`/`GZIP`, fixes `umask`, overrides Git `tar.umask`, normalizes
metadata modes and owner/group, sorts archive entries, and verifies the exact
metadata member set. The requested output directory must not exist; the builder
stages and verifies every output in a sibling directory, then uses pinned GNU
`mv` no-clobber/no-target-directory semantics plus a postcondition to atomically
rename that complete directory into place or fail if the destination appeared.

The builder clears `GOFLAGS`, disables workspaces and persistent Go environment
configuration, rejects every declared `go.mod` replacement, and rejects any
selected dependency replacement. Read-only
`go list -deps -test` discovers only modules linked by production or test
packages using fresh isolated module/build caches and the canonical Go proxy and
sum database. Network access is therefore required for the release build. Every
inventoried module has complete hashes in committed `go.sum`; the builder
rejects any discovery-time `go.sum` change and runs `go mod verify` before
reading legal files from the cache. Downloaded module bytes are bound by those
committed sums. The workflow therefore marks the artifact gate as networked.

The provenance is an in-toto Statement v1 with a project-specific deterministic
predicate. It records source commit/time, source archive digest, release-script
digest, `go.sum` digest, and exact builder-tool contract. It does not claim SLSA
build-run timestamps, invocation identity, or hosted-builder attestation.
