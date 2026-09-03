# Alpha.3 library admission

## Scope and authority

This pass admits `pkg/migration` as the canonical package for GOTTH Board
alpha.3. It does not change database state, migration SQL, the ledger, the
advisory lock, or rollback policy.

## Requirement traceability

| Requirement | Design/specification | Code | Verification |
|---|---|---|---|
| `PGM-A3-01` | `docs/architecture.md` | `pkg/migration/` | canonical outside-package test |
| `PGM-A3-02` | `docs/implementation-spec.md` | `pkg/migration/` | sentinel/API contract tests |
| `PGM-A3-03` | repository layout in `README.md` | separated repository trees | tracked-file inventory |
| `PGM-A3-04` | `docs/verification.md` | unit/integration tests | PostgreSQL 17, clean clone, Judge passes |
| `PGM-A3-05` | architecture and specification | coordinator/ledger/lock code | drift, transaction, and concurrency tests |

## Runtime boundary

- Go 1.26.6, pgx v5, and PostgreSQL 17 are the admitted targets.
- PostgreSQL catalog shape, transaction isolation, session advisory-lock
  ownership, commit ambiguity, and row completeness are correctness boundaries.
- File size is bounded to one MiB; applied rows are read with release length
  plus one as a completeness oracle.
- PostgreSQL integration tests cover malformed ledger/catalog state,
  concurrency, transaction failure, lock ownership, cancellation cleanup, and
  exact release verification. Other PostgreSQL majors remain unsupported.

## Performance admission

No algorithm, query, round-trip count, row bound, or migration I/O changes.
The original implementation moved intact to the canonical package. No speedup
is claimed; benchmark/Amdahl evidence is N/A. Future query or
migration-performance changes must preserve matched PostgreSQL plans and
end-to-end measurements.

## Failure and rollback

Rollback is a revert before the first consumer pin. Unknown commit outcomes
remain machine-classified and are never retried. No live database is a valid
admission target.
