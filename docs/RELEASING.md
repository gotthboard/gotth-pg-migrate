# Release Policy

`gotth-pg-migrate` versions independently under Semantic Versioning. A version number
in another GOTTH repository creates no release obligation here.

If this repository contains a Go module, release tags use the Go-compatible
`vMAJOR.MINOR.PATCH` form, including any SemVer prerelease suffix. Historical
tags are immutable even when they predate this convention.

No release may be tagged until required verification is complete. The
repository uses the MIT license. Reserved placeholders are not tagged at all.

For an admitted release:

1. Select a version from this repository's actual compatibility change.
2. Verify one exact clean candidate commit with every required local, runtime,
   consumer, and review gate.
3. Record one `SOURCE_DATE_EPOCH` for release-document creation, then build the
   source and metadata artifacts twice from that full commit with
   `scripts/build-release-candidate.sh`; reuse that epoch, require byte-identical checksums, and
   inspect the SPDX, provenance, owner license, and dependency legal-file
   bundle under the exact tool contract in `docs/provenance.md`.
4. Update `docs/CHANGELOG.md` with the version, compatibility effect,
   verification, and known limitations.
5. With separate publication authorization, push the candidate and create a new
   annotated tag on canonical Forgejo. Never move or replace a tag.
6. Verify exact tag-object and peeled-commit parity at GitHub through the
   existing one-way mirror.
7. Rebuild from the exact public tag, require equality with the admitted
   candidate artifacts, publish `SHA256SUMS`, and create the GitHub Release.

Forgejo remains the development source of truth. GitHub is the public clone,
Go import, and release-distribution endpoint. Failure of mirror or release
publication blocks the release; it does not authorize reversing mirror
direction or publishing a different commit.
