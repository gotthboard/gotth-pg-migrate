# Feature plan

1. `PGM-01` — extraction baseline: product-neutral module, provenance, tests.
2. `PGM-02` — public API and exact ledger contract.
3. `PGM-03` — PostgreSQL 17 integration and concurrent-run verification.
4. `PGM-04` — independent consumer integration and first tagged release.

Each feature requires scoped tests, race verification, coverage evidence, and
an explicit compatibility ruling. Application migration files never enter this
repository.
