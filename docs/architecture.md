# Architecture

The canonical public implementation lives in `pkg/migration`. The module root
contains no Go package; its job is repository governance and module metadata.

The public boundary consists of `Apply`, `ReleaseVerifier`, the minimal
`VerificationQuerier` interface, and `ErrCommitOutcomeUnknown`. `Apply` owns one
dedicated `pgx.Conn`, loads and hashes the supplied `fs.FS`, acquires the fixed
package advisory lock, attests `public.gotth_schema_migrations`, rejects any
non-prefix history, and applies the pending suffix in order. Each SQL file and
ledger insert commit in the same transaction.

The fixed ledger is deliberate: one application database has one canonical
GOTTH migration history. The v1 session lock key is the frozen signed bigint
`-8593188949330966034`. It serializes this library, not legacy migrators using
other keys. Products that cut over must stop/drain legacy writers first instead
of assuming cross-key serialization or adding runtime identifier configuration.

Each file and ledger row has its own transaction; the release as a whole does
not. A failed file leaves an already committed exact prefix and no dirty flag.
Cancellation stops new work. Cleanup uses bounded contexts detached from the
canceled caller. A commit error wraps `ErrCommitOutcomeUnknown` and is never
retried. The caller uses `errors.Is`, then inspects the ledger and schema before
deciding what happened.

`ReleaseVerifier` is deliberately read-only and lock-free. It may observe an
in-progress release between file transactions and fail exact-head readiness.

Trust boundary: migration SQL is trusted release code, not sandboxed input. It
must not control transactions, manipulate the package ledger, or manipulate
advisory locks; the package intentionally does not embed a PostgreSQL parser.
Database contents and preexisting ledger shape are untrusted until attested. No
identifiers or SQL fragments come from runtime strings. Package-owned catalog
and advisory-lock functions are schema-qualified.
