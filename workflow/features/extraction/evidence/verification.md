# Extraction verification

Verified on 2026-09-02 with Go `go1.26.6-X:nodwarf5`:

- `go vet -mod=readonly ./...`
- `go test -mod=readonly -race -cover ./...`
- statement coverage: 99.6%
- generic package and test environment contain no board import
- the fixed `public.gotth_schema_migrations` ledger remains compatible

The PostgreSQL 17 integration suite is retained behind the `integration` build
tag. It requires an isolated database through
`GOTTH_PG_MIGRATE_TEST_DATABASE_URL`; no production database was used during
repository extraction.
