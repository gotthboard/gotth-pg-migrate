//go:build integration

package migration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func TestAttestHistoryTableRejectsCatalogDriftOnPostgreSQL17(t *testing.T) {
	databaseURL := os.Getenv("GOTTH_PG_MIGRATE_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("GOTTH_PG_MIGRATE_TEST_DATABASE_URL is required for integration tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatalf("pgx.Connect() returned error: %v", err)
	}
	t.Cleanup(func() {
		if err := conn.Close(context.Background()); err != nil {
			t.Errorf("close PostgreSQL connection: %v", err)
		}
	})
	drop := func(ctx context.Context) {
		if _, err := conn.Exec(ctx, "DROP TABLE IF EXISTS public.gotth_schema_migrations CASCADE"); err != nil {
			t.Fatalf("drop history table: %v", err)
		}
	}
	drop(ctx)
	t.Cleanup(func() {
		cleanupContext, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		drop(cleanupContext)
	})

	requirePostgreSQL17(t, ctx, conn)
	drifts := map[string]string{
		"missing digest constraint": "ALTER TABLE public.gotth_schema_migrations DROP CONSTRAINT gotth_schema_migrations_sha256_length",
		"extra column":              "ALTER TABLE public.gotth_schema_migrations ADD COLUMN unexpected text",
		"changed timestamp default": "ALTER TABLE public.gotth_schema_migrations ALTER COLUMN applied_at SET DEFAULT statement_timestamp()",
		"row security":              "ALTER TABLE public.gotth_schema_migrations ENABLE ROW LEVEL SECURITY",
		"insert rewrite rule":       "CREATE RULE gotth_schema_migrations_ignore_insert AS ON INSERT TO public.gotth_schema_migrations DO INSTEAD NOTHING",
		"inherited child":           "CREATE TABLE public.gotth_schema_migrations_child () INHERITS (public.gotth_schema_migrations)",
	}
	for name, statement := range drifts {
		name, statement := name, statement
		t.Run(name, func(t *testing.T) {
			drop(ctx)
			if err := ensureHistoryTable(ctx, conn); err != nil {
				t.Fatalf("ensureHistoryTable() returned error: %v", err)
			}
			if err := attestHistoryTable(ctx, conn); err != nil {
				t.Fatalf("attestHistoryTable() exact table: %v", err)
			}
			if _, err := conn.Exec(ctx, statement); err != nil {
				t.Fatalf("alter history table: %v", err)
			}
			if err := attestHistoryTable(ctx, conn); err == nil {
				t.Fatal("attestHistoryTable() accepted catalog drift")
			}
		})
	}
}

func TestHistoryTableExpressionsIgnoreAndRejectHostileSearchPath(t *testing.T) {
	databaseURL := os.Getenv("GOTTH_PG_MIGRATE_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("GOTTH_PG_MIGRATE_TEST_DATABASE_URL is required for integration tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatalf("pgx.Connect() returned error: %v", err)
	}
	defer conn.Close(context.Background())
	defer func() {
		cleanupContext, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		if _, err := conn.Exec(cleanupContext, `DROP TABLE IF EXISTS public.gotth_schema_migrations;
DROP SCHEMA IF EXISTS gotth_catalog_shadow CASCADE`); err != nil {
			t.Errorf("clean hostile catalog fixture: %v", err)
		}
	}()

	const installShadowSQL = `DROP TABLE IF EXISTS public.gotth_schema_migrations;
DROP SCHEMA IF EXISTS gotth_catalog_shadow CASCADE;
CREATE SCHEMA gotth_catalog_shadow;
CREATE FUNCTION gotth_catalog_shadow.clock_timestamp() RETURNS timestamp with time zone
LANGUAGE sql IMMUTABLE AS 'SELECT ''2000-01-01 UTC''::timestamp with time zone';
CREATE FUNCTION gotth_catalog_shadow.octet_length(bytea) RETURNS integer
LANGUAGE sql IMMUTABLE AS 'SELECT 32';
CREATE FUNCTION gotth_catalog_shadow.int8_greater(bigint, bigint) RETURNS boolean
LANGUAGE sql IMMUTABLE AS 'SELECT true';
CREATE OPERATOR gotth_catalog_shadow.> (
    LEFTARG = bigint, RIGHTARG = bigint, FUNCTION = gotth_catalog_shadow.int8_greater
);
CREATE FUNCTION gotth_catalog_shadow.int4_equal(integer, integer) RETURNS boolean
LANGUAGE sql IMMUTABLE AS 'SELECT true';
CREATE OPERATOR gotth_catalog_shadow.= (
    LEFTARG = integer, RIGHTARG = integer, FUNCTION = gotth_catalog_shadow.int4_equal
);
SET search_path TO gotth_catalog_shadow, pg_catalog`
	if _, err := conn.Exec(ctx, installShadowSQL); err != nil {
		t.Fatalf("install hostile catalog objects: %v", err)
	}
	if err := ensureHistoryTable(ctx, conn); err != nil {
		t.Fatalf("ensureHistoryTable() under hostile search_path: %v", err)
	}
	if err := attestHistoryTable(ctx, conn); err != nil {
		t.Fatalf("attestHistoryTable() rejected qualified built-ins: %v", err)
	}
	if _, err := conn.Exec(ctx, "DROP TABLE public.gotth_schema_migrations"); err != nil {
		t.Fatalf("drop qualified ledger: %v", err)
	}

	const hostileLedgerSQL = `CREATE TABLE public.gotth_schema_migrations (
    version bigint PRIMARY KEY,
    name text NOT NULL,
    sha256 bytea NOT NULL,
    applied_at timestamp with time zone NOT NULL DEFAULT clock_timestamp(),
    CONSTRAINT gotth_schema_migrations_version_positive CHECK (version > 0),
    CONSTRAINT gotth_schema_migrations_sha256_length CHECK (octet_length(sha256) = 32)
)`
	if _, err := conn.Exec(ctx, hostileLedgerSQL); err != nil {
		t.Fatalf("create hostile-expression ledger: %v", err)
	}
	if err := attestHistoryTable(ctx, conn); err == nil {
		t.Fatal("attestHistoryTable() accepted non-pg_catalog expression dependencies")
	}
}
