# GOTTH PostgreSQL Migrate

> **Distribution:** GitHub is the public clone, Go import, and future release endpoint.
> Forgejo remains canonical development. Report public bugs through GitHub
> Issues and security vulnerabilities through GitHub private reporting.
> See [the distribution contract](docs/distribution.md).


`gotth-pg-migrate` is a small forward-only PostgreSQL migration library for
immutable application releases. It validates a contiguous embedded migration
set, serializes application with a session advisory lock, attests its ledger,
rejects drift, applies each file atomically with its ledger record, and refuses
to guess after an unknown commit outcome.

Canonical Go package:
`github.com/gotthboard/gotth-pg-migrate/pkg/migration`. The module root
contains repository governance only. New consumers use the canonical package.

Repository layout:

- `pkg/migration/` — public implementation plus unit and PostgreSQL tests;
- module root — module metadata and repository governance;
- `docs/` — requirements, runtime contracts, and admission records;
- `workflow/` — canonical feature state, review, and evidence.

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
Those constraints prevent two library instances from silently claiming
independent truth in one database. The lock does not serialize legacy migrators that use a
different key; cutovers require a stopped/drained writer boundary. Migration
SQL, grants, rollout, backup, and rollback remain consumer-owned. The exact v1
semantics are frozen in [the version 1 contract](docs/contract-v1.md).

Run `make verify` with the pinned Go toolchain. PostgreSQL integration tests run
with `make verify-integration` when `GOTTH_PG_MIGRATE_TEST_DATABASE_URL` points
to a disposable PostgreSQL 17 database.

Extracted from the migration engine admitted in GOTTH Board 1.0.0-alpha.2 at
commit `47c3a17d8b147e648ad8c57fbc80cec06076e89b`.

## Installation, compatibility, and support

The local candidate is `v1.0.0`; no tag or hosted Release has been published.
Its stable runtime boundary is Go 1.26.6, pgx/v5, and PostgreSQL 17.x. After the
immutable tag is published, install the sole public package with:

```sh
go get github.com/gotthboard/gotth-pg-migrate/pkg/migration@v1.0.0
```

Until that tag exists, do not treat `main` as the stable release. See the
[candidate release notes](docs/release-notes-v1.0.0.md).

The owner-authored repository contents are licensed under the
[MIT License](LICENSE). Third-party dependencies and assets retain their own
licenses. No long-term support promise is implied. Versioning, release
admission, security reporting, and contribution details are in
[the release policy](docs/RELEASING.md), [security policy](SECURITY.md), and
[contribution guide](CONTRIBUTING.md).
