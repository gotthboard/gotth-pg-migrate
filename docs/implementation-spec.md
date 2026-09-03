# Implementation specification

- Canonical package:
  `github.com/gotthboard/gotth-pg-migrate/pkg/migration`.
- The module root contains no Go package. `pkg/migration` is the only public Go
  implementation and import path.

- Migration filenames match `^[0-9]{6}_[a-z0-9]+(?:_[a-z0-9]+)*\.sql$` and are
  contiguous from version one.
- The migration filesystem is flat, nonempty, and bounded to one MiB per file.
- The ledger columns are `version bigint`, `name text`, `sha256 bytea`, and
  `applied_at timestamptz default clock_timestamp()` with exact constraints.
- The package attests columns, defaults, constraints, persistence, row-security
  state, inheritance, user triggers, and rewrite rules before trusting rows.
- Applied history is read with a bound of release length plus one.
- A migration transaction executes the exact SQL string, inserts its digest,
  commits once, and otherwise performs bounded rollback.
- `ReleaseVerifier` hashes files once and performs read-only exact-head checks.
- `VerificationQuerier` names only the `Query` and `QueryRow` methods needed by
  readiness and is satisfied by direct pgx connections and pools.
- Every transaction commit failure wraps `ErrCommitOutcomeUnknown`; execution,
  ledger-insert, and rollback failures do not.
- Public functions reject nil and canceled inputs before external work.

Every PostgreSQL 17 patch release is within the initial verified contract;
major-version expansion requires its own catalog integration evidence. Changes to ledger shape,
filename grammar, advisory lock identity, or failure semantics are breaking.
