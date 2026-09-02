# Verification

Required local gate: `make verify`.

With `GOTTH_PG_MIGRATE_TEST_DATABASE_URL`, integration tests prove atomic failure,
concurrent serialization, exact ledger attestation, release-prefix rejection,
and bounded cleanup on PostgreSQL 17. Coverage must account for every public
function and every fail-closed branch; generated or external database behavior
is not counted as local statement coverage.

The extraction is acceptable only when GOTTH Board can consume a pinned module
without changing its migration files, database ledger, readiness semantics, or
rollback boundary.
