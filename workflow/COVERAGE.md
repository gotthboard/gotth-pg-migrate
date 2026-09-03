# Coverage

The canonical behavior map is `workflow/artifacts/global-coverage-map.md`.
`make verify` covers the sole public package at `pkg/migration`; PostgreSQL
behavior is additionally exercised by `make verify-integration`. Current
statement coverage is 99.6%, with only unreachable defensive driver outcomes
outside the pinned pgx contracts remaining.
