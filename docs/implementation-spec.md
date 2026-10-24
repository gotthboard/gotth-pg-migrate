# Implementation specification

- Canonical package:
  `github.com/gotthboard/gotth-pg-migrate/pkg/migration`.
- The module root contains no Go package. `pkg/migration` is the only public Go
  implementation and import path.

- Migration filenames match `^[0-9]{6}_[a-z0-9]+(?:_[a-z0-9]+)*\.sql$` and are
  contiguous from version one.
- The migration filesystem is flat and nonempty, every entry is a regular
  file, and each file is bounded to one MiB.
- The ledger has exactly four physical attributes with no dropped-column
  residue: `version bigint`, `name text`, `sha256 bytea`, and `applied_at
  timestamptz default clock_timestamp()` with exact modifiers, collations, and
  constraints.
- The package attests columns, defaults, constraints, the sole btree
  primary-key index with default tablespace and no relation options, persistence,
  row-security state, inheritance in either direction, user
  triggers, rewrite rules, and expression dependencies before trusting rows.
- Applied history is read from `ONLY public.gotth_schema_migrations` with a
  bound of release length plus one; the ledger may have no inheritance edge in
  either direction.
- A migration transaction executes the exact SQL string, inserts its digest,
  commits once, and otherwise performs bounded rollback. Migration SQL is
  trusted and must not contain transaction control or manipulate the package
  ledger or advisory lock; v1 does not parse SQL to enforce that restriction.
- `ReleaseVerifier` hashes files once and performs read-only exact-head checks.
- `VerificationQuerier` names only the `Query` and `QueryRow` methods needed by
  readiness and is satisfied by direct pgx connections and pools.
- Every transaction commit failure wraps `ErrCommitOutcomeUnknown`; execution,
  ledger-insert, and rollback failures do not.
- Public functions reject nil and canceled inputs before external work.
- Package-owned PostgreSQL catalog and advisory-lock function references are
  schema-qualified and do not depend on caller `search_path`.

Every PostgreSQL 17 patch release is within the v1 contract; major-version
expansion requires a new compatibility decision and catalog integration
evidence. Changes to ledger shape, filename grammar, advisory-lock identity,
transaction/dirty-state semantics, or unknown-commit classification are
breaking. Error strings and unexported error identities are not API.
