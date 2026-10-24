//go:build integration

package migration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func TestWithMigrationLockOwnsOnePostgreSQLSessionLock(t *testing.T) {
	databaseURL := os.Getenv("GOTTH_PG_MIGRATE_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("GOTTH_PG_MIGRATE_TEST_DATABASE_URL is required for integration tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	owner, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect lock owner: %v", err)
	}
	defer owner.Close(context.Background())
	observer, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect lock observer: %v", err)
	}
	defer observer.Close(context.Background())

	const shadowSQL = `DROP SCHEMA IF EXISTS gotth_lock_shadow CASCADE;
CREATE SCHEMA gotth_lock_shadow;
CREATE FUNCTION gotth_lock_shadow.pg_advisory_lock(bigint) RETURNS void
LANGUAGE sql AS 'SELECT NULL';
CREATE FUNCTION gotth_lock_shadow.pg_advisory_unlock(bigint) RETURNS boolean
LANGUAGE sql AS 'SELECT true';
SET search_path TO gotth_lock_shadow, pg_catalog`
	if _, err := owner.Exec(ctx, shadowSQL); err != nil {
		t.Fatalf("install hostile advisory-lock functions: %v", err)
	}
	defer func() {
		cleanupContext, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		if _, err := observer.Exec(cleanupContext, "DROP SCHEMA IF EXISTS gotth_lock_shadow CASCADE"); err != nil {
			t.Errorf("drop hostile advisory-lock schema: %v", err)
		}
	}()

	if err := withMigrationLock(ctx, owner, func() error {
		var acquired bool
		if err := observer.QueryRow(ctx, "SELECT pg_try_advisory_lock($1)", migrationAdvisoryLockKey).Scan(&acquired); err != nil {
			return err
		}
		if acquired {
			t.Fatal("second PostgreSQL session acquired the held migration lock")
		}
		return nil
	}); err != nil {
		t.Fatalf("withMigrationLock() returned error: %v", err)
	}

	var acquired bool
	if err := observer.QueryRow(ctx, "SELECT pg_try_advisory_lock($1)", migrationAdvisoryLockKey).Scan(&acquired); err != nil {
		t.Fatalf("try lock after release: %v", err)
	}
	if !acquired {
		t.Fatal("migration lock remained held after withMigrationLock returned")
	}
	var released bool
	if err := observer.QueryRow(ctx, releaseMigrationLockSQL, migrationAdvisoryLockKey).Scan(&released); err != nil || !released {
		t.Fatalf("release observer lock = (%t, %v), want (true, nil)", released, err)
	}
}
