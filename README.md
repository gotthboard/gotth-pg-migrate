# GOTTH PostgreSQL Migrate

`gotth-pg-migrate` is a small forward-only PostgreSQL migration library for
immutable application releases. It validates a contiguous embedded migration
set, serializes application with a session advisory lock, attests its ledger,
rejects drift, applies each file atomically with its ledger record, and refuses
to guess after an unknown commit outcome.

The library owns no application schema and has no ORM, daemon, down-migration,
or configuration framework. Applications pass a parsed `pgx.ConnConfig` and an
`fs.FS` containing files named `000001_description.sql`.

```go
if err := migration.Apply(ctx, connectionConfig, migrations); err != nil {
	return err
}
```

Run `make verify` with the pinned Go toolchain. PostgreSQL integration tests run
when `GOTTH_TEST_DATABASE_URL` is set.

Extracted from the migration engine admitted in GOTTH Board 1.0.0-alpha.2 at
commit `47c3a17d8b147e648ad8c57fbc80cec06076e89b`.
