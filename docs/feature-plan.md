# Feature plan

1. `PGM-01` — extraction baseline: product-neutral module, provenance, tests.
2. `PGM-02` — public API and exact ledger contract.
3. `PGM-03` — PostgreSQL 17 integration and concurrent-run verification.
4. `PGM-04` — reusable v0 contract admission: exported readiness seam,
   machine-readable unknown outcome, external-package compile test, disposable
   PostgreSQL 17 verification, and clean-checkout evidence.
5. `PGM-05` — first real consumer pin and tagged release; owned by the consumer
   integration, not fabricated inside this library repository.
6. `PGM-06` — alpha.3 coding-setup admission: canonical package layout,
   requirement/runtime/performance records, clean-clone verification, and two
   clean Judge passes before the consumer pin.

Each feature requires scoped tests, race verification, coverage evidence, and
an explicit compatibility ruling. Application migration files never enter this
repository.
