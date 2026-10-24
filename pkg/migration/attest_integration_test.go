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
		if _, err := conn.Exec(ctx, "DROP TABLE IF EXISTS public.gotth_schema_migrations CASCADE; DROP TYPE IF EXISTS public.gotth_ledger_type CASCADE"); err != nil {
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
		"extra index":               "CREATE INDEX gotth_schema_migrations_name_idx ON public.gotth_schema_migrations (name)",
		"timestamp precision":       "ALTER TABLE public.gotth_schema_migrations ALTER COLUMN applied_at TYPE pg_catalog.timestamptz(0)",
		"text collation":            "ALTER TABLE public.gotth_schema_migrations ALTER COLUMN name TYPE pg_catalog.text COLLATE \"C\"",
		"dropped column residue":    "ALTER TABLE public.gotth_schema_migrations ADD COLUMN residue pg_catalog.text; ALTER TABLE public.gotth_schema_migrations DROP COLUMN residue",
		"primary index options":     "ALTER INDEX public.gotth_schema_migrations_pkey SET (fillfactor=50)",
		"table options":             "ALTER TABLE public.gotth_schema_migrations SET (fillfactor=50)",
		"replica identity":          "ALTER TABLE public.gotth_schema_migrations REPLICA IDENTITY NOTHING",
		"typed table":               "CREATE TYPE public.gotth_ledger_type AS (version pg_catalog.int8, name pg_catalog.text, sha256 pg_catalog.bytea, applied_at pg_catalog.timestamptz); ALTER TABLE public.gotth_schema_migrations OF public.gotth_ledger_type",
		"column storage":            "ALTER TABLE public.gotth_schema_migrations ALTER COLUMN name SET STORAGE EXTERNAL",
		"column compression":        "ALTER TABLE public.gotth_schema_migrations ALTER COLUMN sha256 SET COMPRESSION pglz",
		"column statistics":         "ALTER TABLE public.gotth_schema_migrations ALTER COLUMN name SET STATISTICS 100",
		"column options":            "ALTER TABLE public.gotth_schema_migrations ALTER COLUMN name SET (n_distinct=0.5)",
		"fast default residue":      "DROP TABLE public.gotth_schema_migrations; CREATE TABLE public.gotth_schema_migrations (version pg_catalog.int8 NOT NULL, CONSTRAINT gotth_schema_migrations_pkey PRIMARY KEY (version) USING INDEX TABLESPACE pg_default) USING heap TABLESPACE pg_default; INSERT INTO public.gotth_schema_migrations (version) VALUES (1); ALTER TABLE public.gotth_schema_migrations ADD COLUMN name pg_catalog.text NOT NULL DEFAULT ''::pg_catalog.text; ALTER TABLE public.gotth_schema_migrations ALTER COLUMN name DROP DEFAULT; ALTER TABLE public.gotth_schema_migrations ADD COLUMN sha256 pg_catalog.bytea NOT NULL DEFAULT pg_catalog.decode(pg_catalog.repeat('00', 32), 'hex'); ALTER TABLE public.gotth_schema_migrations ALTER COLUMN sha256 DROP DEFAULT; ALTER TABLE public.gotth_schema_migrations ADD COLUMN applied_at pg_catalog.timestamptz; UPDATE public.gotth_schema_migrations SET applied_at = pg_catalog.clock_timestamp(); ALTER TABLE public.gotth_schema_migrations ALTER COLUMN applied_at SET NOT NULL; ALTER TABLE public.gotth_schema_migrations ALTER COLUMN applied_at SET DEFAULT pg_catalog.clock_timestamp(); ALTER TABLE public.gotth_schema_migrations ADD CONSTRAINT gotth_schema_migrations_version_positive CHECK (version OPERATOR(pg_catalog.>) 0), ADD CONSTRAINT gotth_schema_migrations_sha256_length CHECK (pg_catalog.octet_length(sha256) OPERATOR(pg_catalog.=) 32)",
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
