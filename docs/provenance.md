# Release provenance and reproducibility

The v1 candidate is source-only. Its second archive contains release metadata,
not a compiled binary.

`scripts/build-release-candidate.sh` accepts one clean full commit and emits:

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
1.8.2, GNU coreutils 9.12, GNU awk 5.4.1, GNU findutils 4.11.0-modified,
and GNU grep 3.12-modified. The source commit timestamp supplies archive mtimes; gzip stores no name or
time. Owner/group are normalized and archive entries are sorted.

`go mod download all` may read the configured module proxy when the verified
module cache is cold. Every module is still pinned by `go.sum`; the network is
not a source of unrecorded release bytes. Offline builds are supported only
with a complete verified cache. The workflow therefore marks the artifact gate
as network-capable, not network-free.

The provenance is an in-toto Statement v1 with a project-specific deterministic
predicate. It records source commit/time, source archive digest, release-script
digest, `go.sum` digest, and exact builder-tool contract. It does not claim SLSA
build-run timestamps, invocation identity, or hosted-builder attestation.
