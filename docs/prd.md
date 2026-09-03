# Product requirements

## Problem

Go services need a migration mechanism whose behavior remains understandable
under drift, concurrency, cancellation, and uncertain transaction outcomes.
Copying ad hoc migration runners between products produces silent divergence.

## Goals

- Apply one immutable, contiguous, forward-only PostgreSQL release.
- Record exact filenames and SHA-256 digests in an attested ledger.
- Serialize migrators using one dedicated PostgreSQL session lock.
- Keep each migration and its ledger row in one transaction.
- Verify readiness against an exact compiled release without writes.
- Expose a nameable readiness interface and a machine-detectable unknown-commit
  outcome so external consumers never parse error strings.
- Bound files, rows, connections, cleanup, and failure output.

## Non-goals

- Down migrations, ORM integration, schema generation, or application seeding.
- Supporting databases other than PostgreSQL.
- Retrying unknown commit outcomes.
- Hosting multiple independent migration histories in one database.

## Acceptance

The library must pass unit, race, drift, malformed-ledger, cancellation,
concurrency, and PostgreSQL 17 integration tests. Applications remain the sole
owners of migration SQL and rollout/rollback decisions. A consumer-package
test must compile against only the exported API.
