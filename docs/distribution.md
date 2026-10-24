# Distribution Contract

## Endpoints

- Canonical development source:
  <https://git.dannyhunn.com/gotthboard/gotth-pg-migrate>
- Public clone, Go import, and future releases:
  <https://github.com/gotthboard/gotth-pg-migrate>
- Public bug tracker:
  <https://github.com/gotthboard/gotth-pg-migrate/issues>
- Private vulnerability reports:
  <https://github.com/gotthboard/gotth-pg-migrate/security/advisories/new>

Forgejo pushes one way to GitHub. GitHub does not feed commits or tags back to
Forgejo. A ref is distributed only when the exact object ID is visible at both
endpoints.

## Maturity and compatibility

Current status: local `v1.0.0` candidate, not tagged or published. Its frozen
database compatibility target is PostgreSQL 17.x.

## Current source use

No stable tag exists yet. The admitted candidate package path is:

```go
import migration "github.com/gotthboard/gotth-pg-migrate/pkg/migration"
```

After publication, install it with:

```sh
go get github.com/gotthboard/gotth-pg-migrate/pkg/migration@v1.0.0
```

Before publication, an exact candidate commit may be used only for isolated
admission. Do not mistake the moving `main` branch for the v1 contract.

The repository pins Go 1.26.6 where a Go module exists. Supported protocol,
runtime, database, and tool versions remain the ones stated in the README and
project verification documents; this distribution change does not widen those
contracts.

## License

The maintainer selected the MIT license for the owner-authored contents of
this repository. The standard grant is in `LICENSE`; third-party dependencies
and assets retain their own licenses. This decision does not publish a release
or create a support or compatibility promise.

## Migration traceability

| Requirement | Repository implementation | Verification |
| --- | --- | --- |
| DIST-001 | Existing history, tags, worktrees, and mirror direction remain unchanged | pinned ref and worktree inventory |
| DIST-002 | Module directive, exact self-imports, fixtures, and examples use the GitHub identity | stale-prefix search, tidy, vet, test, and clean public import |
| DIST-003/004 | README, contribution, security, changelog, and release contracts describe public use and support | documentation audit |
| DIST-006 | MIT license is present and scoped to owner-authored contents | license and dependency inventory |
| DIST-008 | Forgejo remains source and GitHub remains the one-way mirror target | push-mirror configuration and exact ref comparison |
