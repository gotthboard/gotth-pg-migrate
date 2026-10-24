// Package migration applies an immutable, contiguous PostgreSQL 17 migration
// history and rejects checksum drift, gaps, rewrites, and partial releases.
//
// Version 1 uses the fixed public.gotth_schema_migrations ledger, one fixed
// session advisory-lock key, one transaction per migration file, and no dirty
// flag or automatic retry. ErrCommitOutcomeUnknown is the only stable package
// error identity; other error text and wrapping are not API.
package migration
