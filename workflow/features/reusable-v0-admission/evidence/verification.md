# Reusable v0 migration admission

Verified implementation commit
`a12915573f7ee84d95d439b66fb4929157b101c2` on 2026-09-03 UTC.

## Contract evidence

- The application retains one fixed, compatible
  `public.gotth_schema_migrations` ledger and package advisory lock per
  database. No runtime SQL identifier configuration was added.
- External consumers can name `VerificationQuerier`; compile assertions prove
  both `pgx.Conn` and `pgxpool.Pool` satisfy the readiness boundary.
- Commit failures wrap `ErrCommitOutcomeUnknown` and preserve the driver cause.
  Execution, record, and rollback failures do not carry that classification.
- The integration harness accepts the supported PostgreSQL 17 major series
  instead of accidentally pinning one patch number.

## Verification

- Go toolchain: `go1.26.6-X:nodwarf5`.
- `make verify`: pass; formatting, vet, race, and 99.6% statement coverage.
- `go test -mod=readonly -race -count=50 ./...`: pass.
- Clean local clone of the committed feature branch followed by `make verify`:
  pass with no generated worktree changes.
- Disposable PostgreSQL container: official `postgres:17`, reported
  `server_version_num=170010`.
- `make verify-integration`: pass against that disposable database. The
  container and temporary checkout were removed afterward; no shared or live
  database was used.
- The remaining 0.4% statement gap is the defensive integer-overflow branch in
  the applied-row query limit. It cannot be reached with an allocatable Go
  migration slice and does not hide external database behavior.

## Graph evidence

- Graphify 0.9.32 code-only graph: 215 nodes, 336 directed edges, 15
  communities, no self-loops, exact duplicate edges, or same-endpoint relation
  groups.
- Graph SHA-256:
  `4fea57febb61c8c5fb3a04341ef134b3c5c557fdb4fd7ef842f5acdc7979386b`.
- Graph cache:
  `/home/linus/.cache/openclaw-graphify/gotth-pg-migrate-reusable/graphify-out/graph.json`.
  Extraction changed no repository file.

No tag or consumer pin was created. The next irreversible compatibility event
belongs to a real consumer integration.
