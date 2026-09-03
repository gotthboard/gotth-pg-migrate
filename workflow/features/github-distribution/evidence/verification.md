# GitHub Distribution Verification

Status: complete

## Identity and scope

- Pinned base: `3d843b88514730a5ef5372271a81af5e89cb1a78`
- Publicly verified candidate: `430659d47a12695e3681efdf4365af76c91ef26c`
- Declared module: `github.com/gotthboard/gotth-pg-migrate`
- Runtime/API behavior: unchanged; this is a module-identity and distribution
  contract migration.

Exact stale-prefix searches found no old module or import identity in Go source,
`go.mod`, examples, or fixtures. Canonical Forgejo URLs remain only where the
development, issue, contribution, and security-reporting endpoints are stated.

## Verification

- Local `go mod tidy` produced no dependency drift.
- Local `go vet -mod=readonly ./...` passed.
- Local `go test -mod=readonly ./...` passed.
- On `development`, `make verify` passed with race coverage 99.6%.
- On `development`, `go test -mod=readonly -race -count=50 ./...` passed.
- PostgreSQL 17.10 integration tests passed in a disposable loopback-only container.
- A fresh public GitHub clone of `feature/github-distribution` resolved exact
  commit `430659d47a12695e3681efdf4365af76c91ef26c` and passed `go test -mod=readonly ./...`.
- A fresh external consumer compiled the public package through both direct VCS
  resolution and `https://proxy.golang.org,direct` at
  `v0.0.0-20260903060720-430659d47a12`.
- Complete Forgejo and GitHub advertised head/tag ref sets matched after the
  candidate push.
- A fresh public GitHub `main` clone resolved
  `8d18b4fd5a362b5ac9c940f85c4eb94b021192cc`, produced no `go mod tidy` drift, and passed
  `go test -mod=readonly ./...`.
- Fresh external consumers resolved `@main` through direct VCS and
  `https://proxy.golang.org,direct`, then compiled at
  `v0.0.0-20260903062630-8d18b4fd5a36`.

The slash-containing feature ref is accepted by direct VCS resolution but is
not a legal version query for the module proxy. The pre-promotion proxy lane
therefore used the exact candidate pseudo-version above; both final `@main`
lanes passed after promotion.

## Impact graph

Graphify recorded 215 nodes / 484 edges at implementation commit. Graph SHA-256:
`ec278a9cf19e5e115ac1fd2ec30e2ab9b9f788503136e37f7993a13a650624ed`.
Subsequent commits before this record changed documentation only. The merged
suite graph had 4,116 nodes and 11,415 edges, with no
cross-repository module dependency edge.

## Admission and residual gates

Two cold Judge passes reviewed the completed candidate before promotion. This
completion update changes evidence and workflow state only and receives two
fresh cold passes before commit. No performance benchmark applies because
executable paths and data flow are unchanged.

No license was selected. Release tags remain blocked until Danny closes that
decision gate. GitHub metadata mutation lacks authentication. Forgejo is still
private, so unauthenticated public contribution and private vulnerability
reporting remain unresolved. Account conversion and ownership changes were not
performed.
