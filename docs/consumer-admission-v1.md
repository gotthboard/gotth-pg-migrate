# Representative consumer admission

## Scope

GOTTH Board is the representative consumer because this library was extracted
from its migration engine and its current history uses the v1 six-digit file
contract. This admission reads an exact Board checkout and copies fixtures only
into disposable verification space. It does not modify Board, Mail, WorldWeave,
or a shared database and does not claim that any product has switched imports.

- Board source: `e075ecb96b252a92f150cad70bde5f632987ac06`.
- Candidate base: `bafb20b5aee66222fb7dd0d6abe7ee970cdb7878`.
- Board migration count/head: 14 / `000014_administrator_smtp.sql`.
- Sorted migration checksum-manifest SHA-256:
  `92e7d0ff32d2b184df3aa776b6845f30cfeb04ff5876e7c8346b17e5b4c878fc`.

## Mechanism comparison

At the exact Board revision, `api.go`, `attest.go`, `bootstrap.go`,
`coordinator.go`, `drift.go`, `owner.go`, `query.go`, and `rows.go` were
byte-identical to the library baseline. The v1 candidate deliberately hardens
that baseline with regular-file validation, schema-qualified built-ins and
types, exact modifier/collation/index/expression/inheritance attestation,
`ONLY` history reads, complete connection-error causes, and qualified advisory
functions. Its public differences are an exported `ErrCommitOutcomeUnknown`,
an exported readiness query interface usable by direct connections and pools,
and the library-wide lock key instead of Board's product-local key.

The exact Board ledger and files pass the stricter candidate unchanged;
checksum, prefix/drift, per-file transaction, and connection-ownership semantics
remain compatible. Advisory-lock identity is not
compatible. A future Board cutover must stop/drain the legacy migrator before
starting the library and must not run both concurrently.

## Reused exact evidence

Historical product evidence is context, not fresh candidate verification:

- `docs/evidence/beta1-03-upgrade-restore-3c9a184.txt` in the Board checkout,
  SHA-256 `50f07e7b799367db2d8d416f3010c4157377c0a7620331f8286257dedcf11a40`,
  records packaged fresh/upgrade/restore behavior on PostgreSQL 17.10.
- `docs/evidence/beta1-09-05-pre-live-admission.txt`, SHA-256
  `e03b6a547cebeab4b0edab39a4d65c4d25e3c5421d0230bc748f111e6d10a30d`,
  records later exact migration/readiness admission.

Fresh candidate admission must independently compile an external module,
construct a verifier from the 14 exact Board files, apply them to an isolated
PostgreSQL 17 database, verify exact head, reject one changed copied byte, and
exercise concurrent v1 callers. Results belong in revision-bound external
evidence; they must not rewrite product records.

## Exclusions

Mail's current four-digit migration filenames and different migration mechanism
are not v1 inputs. WorldWeave has no admitted migration input in the available
consumer evidence. Neither is mutated or represented as compatible.
