# Architecture

The public boundary consists of `Apply` and `ReleaseVerifier`. `Apply` owns one
dedicated `pgx.Conn`, loads and hashes the supplied `fs.FS`, acquires the fixed
package advisory lock, attests `public.gotth_schema_migrations`, rejects any
non-prefix history, and applies the pending suffix in order. Each SQL file and
ledger insert commit in the same transaction.

The fixed ledger is deliberate: one application database has one canonical
GOTTH migration history. Products that attempt to share a database must first
define ownership instead of adding arbitrary table-name configuration.

Cancellation stops new work. Cleanup uses bounded contexts detached from the
canceled caller. A commit error is reported as outcome-unknown and is never
retried. The caller must inspect the ledger before deciding what happened.

Trust boundary: migration SQL is trusted release input; database contents and
preexisting ledger shape are untrusted until attested. No identifiers or SQL
fragments come from runtime strings.
