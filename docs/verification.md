# Verification

Required local gate: `make verify`.

With `GOTTH_PG_MIGRATE_TEST_DATABASE_URL`, integration tests prove atomic failure,
concurrent serialization, exact ledger attestation, release-prefix rejection,
and bounded cleanup on a disposable PostgreSQL 17 instance. Coverage must account for every public
function and every fail-closed branch; generated or external database behavior
is not counted as local statement coverage.

The reusable contract is admissible only after an external-package compile
test, disposable PostgreSQL 17 integration, race repetition, and clean-clone
verification pass. A tag remains a separate consumer-pinning decision.
