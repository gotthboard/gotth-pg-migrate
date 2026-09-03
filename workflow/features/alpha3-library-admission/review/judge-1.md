# Judge pass 1 — rejected and repaired

The first cold review rejected a module-root compatibility facade. No released
tag or consumer pin established that import path, so the facade created a
second public API without preserving userspace.

Repair: remove the facade, retain exactly one public library package at
`pkg/migration`, preserve sentinel identity in that package, and keep the
module root for governance.
