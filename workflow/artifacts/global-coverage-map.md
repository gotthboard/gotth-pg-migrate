# Global coverage map

| Subsystem | Risk | Required harnesses | Current evidence | Known gaps |
| --- | --- | --- | --- | --- |
| File loading and hashing | High | Unit, malformed input | `loader_test.go` | None |
| Ledger attestation and drift | Critical | Unit, PostgreSQL integration | `attest*_test.go`, `drift_test.go` | None |
| Lock and connection ownership | Critical | Unit, race, PostgreSQL integration | `lock*_test.go`, `owner_test.go` | None |
| Transaction application | Critical | Unit, PostgreSQL integration | `apply_one*_test.go` | None |
| Public consumer contract | High | External-package compile, unit | `public_api_test.go` | None |
| Readiness verification | High | Unit, PostgreSQL integration | `verify_test.go`, coordinator integration | None after PGM-04 evidence |
