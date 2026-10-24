# Version 1 contract

This document freezes the supported behavior of `v1.0.0`. Anything not listed here is not a compatibility promise.

## Runtime boundary

- Go module: `github.com/gotthboard/gotth-pg-migrate`.
- Sole public package: `github.com/gotthboard/gotth-pg-migrate/pkg/migration`.
- Supported database: PostgreSQL 17.x only. Other major versions are unsupported.
- Driver boundary: pgx/v5. `Apply` accepts a parsed `*pgx.ConnConfig` and owns a new direct connection for the entire call; it never uses or creates a pool.

## Public API and errors

The v1 API is exactly `Apply`, `NewReleaseVerifier`, `(*ReleaseVerifier).Verify`, `VerificationQuerier`, and `ErrCommitOutcomeUnknown`. Only `ErrCommitOutcomeUnknown` is a stable package error identity. Every failed transaction `Commit` wraps it, and no other failure does. Callers use `errors.Is`; error strings, wrapping depth, and unexported errors are not API. Driver and context causes are preserved when available.

## Release input and checksums

The supplied `fs.FS` is a complete release history, not only new migrations. It must be flat and nonempty. Every entry is a regular file named `NNNNNN_description.sql`; versions begin at `000001`, are contiguous, and descriptions contain lowercase letters, digits, and single underscores between nonempty words. Each file contains non-whitespace SQL and is at most 1 MiB.

SHA-256 covers the exact file bytes. Line endings, trailing whitespace, comments,
and every other byte are significant. An applied file is immutable: changing
its name or bytes is drift.

Migration SQL is trusted release code, not sandboxed input. A v1 migration may
contain multiple statements but must not issue transaction-control statements,
directly read or modify `public.gotth_schema_migrations`, call advisory-lock
functions, or otherwise alter the package transaction, ledger, or lock. The
package does not parse SQL to enforce this boundary. Files that violate it are
outside the v1 contract and can defeat atomicity or serialization.

## Ledger and history

The sole supported ledger is `public.gotth_schema_migrations` with the exact
physical columns (including absence of dropped-column residue), order, types,
constraints, default, persistence, row-security, inheritance, trigger,
rewrite-rule, sole default btree primary-key index, and
expression-dependency shape defined by the implementation. It may be neither parent nor child in table inheritance.
`Apply` may create it while holding the migration lock, then attests its exact
catalog shape before trusting rows. A lookalike or modified table is rejected.

Rows must be the exact contiguous prefix of the supplied release, starting at
version one. Missing, reordered, renamed, checksum-mismatched, or
release-unknown rows are rejected. Reads use
`ONLY public.gotth_schema_migrations`, are bounded to release length plus one
row, and cannot consume rows from inherited child tables. Existing rows are
never edited or deleted.

## Concurrency and advisory lock

`Apply` uses one blocking PostgreSQL session advisory lock with signed bigint key `-8593188949330966034`. It is the big-endian signed first eight bytes of SHA-256(`gotth-pg-migrate:v1`) and is frozen for v1. The dedicated connection owns it; v1 callers using the same database serialize. It does not coordinate with other keys, migration libraries, or ad hoc SQL.

Legacy product-local migrators, including the admitted GOTTH Board engine, use a different lock key even when ledger and migration bytes are compatible. A cutover must stop or drain every legacy migrator before starting v1. Concurrent old/new migration writers are unsupported.

Unlock and connection close use bounded five-second contexts detached from caller cancellation. Unlock or close failure is joined to the primary failure. Closing the dedicated connection is the final lock-release backstop.

## Transactions, failure, and dirty state

For contract-compliant SQL, each pending file is one PostgreSQL transaction:
execute exact SQL, insert its ledger row, then commit. The whole release is not
one transaction. If file N fails before commit, previously committed files
remain an exact prefix and file N SQL plus ledger insert roll back together. A
later invocation may resume that prefix after the failure is understood and
corrected.

There is no dirty flag. Atomic per-file SQL plus the ledger prefix is the only state model. If `Commit` returns an error, the outcome is unknown; the error matches `ErrCommitOutcomeUnknown`, the package does not retry, and the operator must inspect the ledger and affected schema before deciding whether retry is safe. Down migrations and automatic repair are absent.

Cancellation stops new work but cannot revoke a commit already accepted by PostgreSQL. Rollback, unlock, and close are bounded cleanup attempts; failures remain visible.

## Verification semantics

`ReleaseVerifier` loads and hashes the complete release once. `Verify` attests the ledger and requires exact head without writes and without taking the advisory lock. It can race with an active migrator and may transiently fail between files. Readiness callers must report failure, not reinterpret it as success or mutate state.

## Ownership exclusions

Consumers own migration SQL, privileges, backups, rollout sequencing, legacy cutover, forward repair, and restore. The package owns no application schema, configuration system, daemon, ORM, seeding, down migration, or retry policy.
