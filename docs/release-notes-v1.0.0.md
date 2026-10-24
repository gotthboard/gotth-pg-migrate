# gotth-pg-migrate v1.0.0

First stable release of the small forward-only PostgreSQL migration library extracted from GOTTH Board.

## Install

```sh
go get github.com/gotthboard/gotth-pg-migrate/pkg/migration@v1.0.0
```

```go
import migration "github.com/gotthboard/gotth-pg-migrate/pkg/migration"
```

## Stable boundary

- PostgreSQL 17.x and pgx/v5.
- One exact `public.gotth_schema_migrations` ledger per database.
- Contiguous six-digit filenames and SHA-256 over exact SQL bytes.
- One frozen package session advisory lock.
- One transaction per file and ledger row; no dirty flag or down migration.
- Exact-prefix drift rejection and exact-head read-only verification.
- Machine-detectable unknown commit outcomes through `errors.Is(err, migration.ErrCommitOutcomeUnknown)`.

See [the complete v1 contract](contract-v1.md). A cutover from a legacy product-local migrator requires a stopped/drained writer boundary because its advisory-lock key may differ even when ledger bytes are compatible.

## Candidate status

This document describes the local `v1.0.0` candidate. No tag or hosted Release exists yet. The final tag must point to the exact admitted candidate; publication must not rebuild from another tree.

The release consists of source plus deterministic metadata: SHA-256 checksums, SLSA-compatible provenance, an SPDX 2.3 SBOM, the MIT license, and exact third-party license texts. `scripts/build-release-candidate.sh` rebuilds those artifacts from a clean full commit.
