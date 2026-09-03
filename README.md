# GOTTH PostgreSQL Migrate

> **Distribution:** GitHub is the public clone, Go import, and future release endpoint.
> Forgejo remains canonical development and the issue/contribution location.
> See [the distribution contract](docs/distribution.md).


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
	if errors.Is(err, migration.ErrCommitOutcomeUnknown) {
		// Inspect the ledger before deciding whether any retry is safe.
	}
	return err
}
```

Readiness can validate the exact embedded release without taking the migration
lock or writing to PostgreSQL:

```go
verifier, err := migration.NewReleaseVerifier(migrations)
if err == nil {
	err = verifier.Verify(ctx, database)
}
```

The explicit compatibility boundary is PostgreSQL 17, one application-owned
migration history per database, the fixed
`public.gotth_schema_migrations` ledger, and the fixed package advisory lock.
Those constraints prevent two libraries from silently claiming independent
truth in one database. Migration SQL, grants, rollout, backup, and rollback
remain consumer-owned.

Run `make verify` with the pinned Go toolchain. PostgreSQL integration tests run
with `make verify-integration` when `GOTTH_PG_MIGRATE_TEST_DATABASE_URL` points
to a disposable PostgreSQL 17 database.

Extracted from the migration engine admitted in GOTTH Board 1.0.0-alpha.2 at
commit `47c3a17d8b147e648ad8c57fbc80cec06076e89b`.

## Installation, compatibility, and support

Unreleased. The Go API is pre-1.0 and its current runtime compatibility target
is PostgreSQL 17.

No post-migration version has been tagged. To inspect the current source
before the first admitted release:

```sh
go get github.com/gotthboard/gotth-pg-migrate@main
```

The repository has no selected license and no long-term support promise.
Versioning, release admission, security reporting, and contribution details are
in [the release policy](docs/RELEASING.md), [security policy](SECURITY.md), and
[contribution guide](CONTRIBUTING.md).
