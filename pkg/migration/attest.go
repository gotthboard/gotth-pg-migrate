package migration

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

const attestHistoryTableSQL = `WITH target AS (
    SELECT class.oid
    FROM pg_catalog.pg_class AS class
    JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid OPERATOR(pg_catalog.=) class.relnamespace
    JOIN pg_catalog.pg_am AS table_access_method ON table_access_method.oid OPERATOR(pg_catalog.=) class.relam
    WHERE namespace.nspname OPERATOR(pg_catalog.=) 'public'
      AND class.relname OPERATOR(pg_catalog.=) 'gotth_schema_migrations'
      AND class.relkind OPERATOR(pg_catalog.=) 'r'
      AND class.relpersistence OPERATOR(pg_catalog.=) 'p'
      AND table_access_method.amname OPERATOR(pg_catalog.=) 'heap'
      AND class.reloptions IS NULL AND class.relreplident OPERATOR(pg_catalog.=) 'd'
      AND NOT class.relispartition
      AND COALESCE(NULLIF(class.reltablespace, 0), (
          SELECT database_value.dattablespace
          FROM pg_catalog.pg_database AS database_value
          WHERE database_value.datname OPERATOR(pg_catalog.=) pg_catalog.current_database()))
          OPERATOR(pg_catalog.=) (
              SELECT tablespace_value.oid
              FROM pg_catalog.pg_tablespace AS tablespace_value
              WHERE tablespace_value.spcname OPERATOR(pg_catalog.=) 'pg_default')
      AND NOT class.relrowsecurity
      AND NOT class.relforcerowsecurity
)
SELECT
    (SELECT pg_catalog.count(*) OPERATOR(pg_catalog.=) 1 FROM target)
    AND (SELECT pg_catalog.count(*) OPERATOR(pg_catalog.=) 4
         FROM pg_catalog.pg_attribute AS attribute
         JOIN target ON target.oid OPERATOR(pg_catalog.=) attribute.attrelid
         WHERE attribute.attnum OPERATOR(pg_catalog.>) 0)
    AND EXISTS (
        SELECT 1 FROM pg_catalog.pg_attribute AS attribute JOIN target ON target.oid OPERATOR(pg_catalog.=) attribute.attrelid
        WHERE attribute.attnum OPERATOR(pg_catalog.=) 1 AND attribute.attname OPERATOR(pg_catalog.=) 'version'
          AND attribute.atttypid OPERATOR(pg_catalog.=) 'pg_catalog.int8'::pg_catalog.regtype
          AND attribute.atttypmod OPERATOR(pg_catalog.=) -1
          AND attribute.attcollation OPERATOR(pg_catalog.=) (
              SELECT type_value.typcollation FROM pg_catalog.pg_type AS type_value
              WHERE type_value.oid OPERATOR(pg_catalog.=) attribute.atttypid)
          AND attribute.attnotnull
          AND attribute.attidentity OPERATOR(pg_catalog.=) '' AND attribute.attgenerated OPERATOR(pg_catalog.=) '')
    AND EXISTS (
        SELECT 1 FROM pg_catalog.pg_attribute AS attribute JOIN target ON target.oid OPERATOR(pg_catalog.=) attribute.attrelid
        WHERE attribute.attnum OPERATOR(pg_catalog.=) 2 AND attribute.attname OPERATOR(pg_catalog.=) 'name'
          AND attribute.atttypid OPERATOR(pg_catalog.=) 'pg_catalog.text'::pg_catalog.regtype
          AND attribute.atttypmod OPERATOR(pg_catalog.=) -1
          AND attribute.attcollation OPERATOR(pg_catalog.=) (
              SELECT type_value.typcollation FROM pg_catalog.pg_type AS type_value
              WHERE type_value.oid OPERATOR(pg_catalog.=) attribute.atttypid)
          AND attribute.attnotnull
          AND attribute.attidentity OPERATOR(pg_catalog.=) '' AND attribute.attgenerated OPERATOR(pg_catalog.=) '')
    AND EXISTS (
        SELECT 1 FROM pg_catalog.pg_attribute AS attribute JOIN target ON target.oid OPERATOR(pg_catalog.=) attribute.attrelid
        WHERE attribute.attnum OPERATOR(pg_catalog.=) 3 AND attribute.attname OPERATOR(pg_catalog.=) 'sha256'
          AND attribute.atttypid OPERATOR(pg_catalog.=) 'pg_catalog.bytea'::pg_catalog.regtype
          AND attribute.atttypmod OPERATOR(pg_catalog.=) -1
          AND attribute.attcollation OPERATOR(pg_catalog.=) (
              SELECT type_value.typcollation FROM pg_catalog.pg_type AS type_value
              WHERE type_value.oid OPERATOR(pg_catalog.=) attribute.atttypid)
          AND attribute.attnotnull
          AND attribute.attidentity OPERATOR(pg_catalog.=) '' AND attribute.attgenerated OPERATOR(pg_catalog.=) '')
    AND EXISTS (
        SELECT 1 FROM pg_catalog.pg_attribute AS attribute JOIN target ON target.oid OPERATOR(pg_catalog.=) attribute.attrelid
        WHERE attribute.attnum OPERATOR(pg_catalog.=) 4 AND attribute.attname OPERATOR(pg_catalog.=) 'applied_at'
          AND attribute.atttypid OPERATOR(pg_catalog.=) 'pg_catalog.timestamptz'::pg_catalog.regtype
          AND attribute.atttypmod OPERATOR(pg_catalog.=) -1
          AND attribute.attcollation OPERATOR(pg_catalog.=) (
              SELECT type_value.typcollation FROM pg_catalog.pg_type AS type_value
              WHERE type_value.oid OPERATOR(pg_catalog.=) attribute.atttypid)
          AND attribute.attnotnull
          AND attribute.attidentity OPERATOR(pg_catalog.=) '' AND attribute.attgenerated OPERATOR(pg_catalog.=) '')
    AND (SELECT pg_catalog.count(*) OPERATOR(pg_catalog.=) 1
         FROM pg_catalog.pg_attrdef AS default_value
         JOIN target ON target.oid OPERATOR(pg_catalog.=) default_value.adrelid)
    AND EXISTS (
        SELECT 1 FROM pg_catalog.pg_attrdef AS default_value JOIN target ON target.oid OPERATOR(pg_catalog.=) default_value.adrelid
        WHERE default_value.adnum OPERATOR(pg_catalog.=) 4
          AND pg_catalog.replace(
                  pg_catalog.pg_get_expr(default_value.adbin, default_value.adrelid, false),
                  'pg_catalog.clock_timestamp', 'clock_timestamp')
              OPERATOR(pg_catalog.=) 'clock_timestamp()')
    AND NOT EXISTS (
        SELECT 1
        FROM pg_catalog.pg_depend AS dependency
        JOIN pg_catalog.pg_attrdef AS default_value
          ON dependency.classid OPERATOR(pg_catalog.=) 'pg_catalog.pg_attrdef'::pg_catalog.regclass
         AND dependency.objid OPERATOR(pg_catalog.=) default_value.oid
        JOIN target ON target.oid OPERATOR(pg_catalog.=) default_value.adrelid
        LEFT JOIN pg_catalog.pg_proc AS referenced_function
          ON dependency.refclassid OPERATOR(pg_catalog.=) 'pg_catalog.pg_proc'::pg_catalog.regclass
         AND dependency.refobjid OPERATOR(pg_catalog.=) referenced_function.oid
        LEFT JOIN pg_catalog.pg_operator AS referenced_operator
          ON dependency.refclassid OPERATOR(pg_catalog.=) 'pg_catalog.pg_operator'::pg_catalog.regclass
         AND dependency.refobjid OPERATOR(pg_catalog.=) referenced_operator.oid
        JOIN pg_catalog.pg_namespace AS referenced_namespace
          ON referenced_namespace.oid OPERATOR(pg_catalog.=)
             COALESCE(referenced_function.pronamespace, referenced_operator.oprnamespace)
        WHERE referenced_namespace.nspname OPERATOR(pg_catalog.<>) 'pg_catalog')
    AND (SELECT pg_catalog.count(*) OPERATOR(pg_catalog.=) 3
         FROM pg_catalog.pg_constraint AS constraint_value
         JOIN target ON target.oid OPERATOR(pg_catalog.=) constraint_value.conrelid)
    AND (SELECT pg_catalog.count(*) OPERATOR(pg_catalog.=) 1
         FROM pg_catalog.pg_index AS index_value
         JOIN target ON target.oid OPERATOR(pg_catalog.=) index_value.indrelid)
    AND EXISTS (
        SELECT 1
        FROM pg_catalog.pg_constraint AS constraint_value
        JOIN target ON target.oid OPERATOR(pg_catalog.=) constraint_value.conrelid
        JOIN pg_catalog.pg_index AS index_value ON index_value.indexrelid OPERATOR(pg_catalog.=) constraint_value.conindid
        JOIN pg_catalog.pg_class AS index_class ON index_class.oid OPERATOR(pg_catalog.=) index_value.indexrelid
        JOIN pg_catalog.pg_am AS access_method ON access_method.oid OPERATOR(pg_catalog.=) index_class.relam
        WHERE constraint_value.conname OPERATOR(pg_catalog.=) 'gotth_schema_migrations_pkey' AND constraint_value.contype OPERATOR(pg_catalog.=) 'p'
          AND constraint_value.convalidated AND index_value.indisprimary AND index_value.indisunique
          AND index_value.indisvalid AND index_value.indisready AND index_value.indimmediate
          AND index_value.indislive AND NOT index_value.indcheckxmin AND NOT index_value.indnullsnotdistinct
          AND NOT index_value.indisclustered AND NOT index_value.indisexclusion AND NOT index_value.indisreplident
          AND index_value.indnkeyatts OPERATOR(pg_catalog.=) 1 AND index_value.indnatts OPERATOR(pg_catalog.=) 1
          AND index_value.indkey OPERATOR(pg_catalog.=) '1'::pg_catalog.int2vector
          AND index_value.indcollation OPERATOR(pg_catalog.=) '0'::pg_catalog.oidvector
          AND index_value.indoption OPERATOR(pg_catalog.=) '0'::pg_catalog.int2vector
          AND index_value.indclass[0] OPERATOR(pg_catalog.=) (
              SELECT operator_class.oid
              FROM pg_catalog.pg_opclass AS operator_class
              JOIN pg_catalog.pg_am AS operator_access_method
                ON operator_access_method.oid OPERATOR(pg_catalog.=) operator_class.opcmethod
              WHERE operator_access_method.amname OPERATOR(pg_catalog.=) 'btree'
                AND operator_class.opcintype OPERATOR(pg_catalog.=) 'pg_catalog.int8'::pg_catalog.regtype
                AND operator_class.opcdefault)
          AND index_value.indexprs IS NULL AND index_value.indpred IS NULL
          AND access_method.amname OPERATOR(pg_catalog.=) 'btree'
          AND COALESCE(NULLIF(index_class.reltablespace, 0), (
              SELECT database_value.dattablespace
              FROM pg_catalog.pg_database AS database_value
              WHERE database_value.datname OPERATOR(pg_catalog.=) pg_catalog.current_database()))
              OPERATOR(pg_catalog.=) (
                  SELECT tablespace_value.oid
                  FROM pg_catalog.pg_tablespace AS tablespace_value
                  WHERE tablespace_value.spcname OPERATOR(pg_catalog.=) 'pg_default')
          AND index_class.reloptions IS NULL
          AND pg_catalog.replace(
                  pg_catalog.pg_get_constraintdef(constraint_value.oid, false),
                  ' USING INDEX TABLESPACE pg_default', '')
              OPERATOR(pg_catalog.=) 'PRIMARY KEY (version)')
    AND EXISTS (
        SELECT 1 FROM pg_catalog.pg_constraint AS constraint_value JOIN target ON target.oid OPERATOR(pg_catalog.=) constraint_value.conrelid
        WHERE constraint_value.conname OPERATOR(pg_catalog.=) 'gotth_schema_migrations_version_positive' AND constraint_value.contype OPERATOR(pg_catalog.=) 'c'
          AND constraint_value.convalidated
          AND pg_catalog.replace(
                  pg_catalog.pg_get_constraintdef(constraint_value.oid, false),
                  'OPERATOR(pg_catalog.>)', '>')
              OPERATOR(pg_catalog.=) 'CHECK ((version > 0))')
    AND EXISTS (
        SELECT 1 FROM pg_catalog.pg_constraint AS constraint_value JOIN target ON target.oid OPERATOR(pg_catalog.=) constraint_value.conrelid
        WHERE constraint_value.conname OPERATOR(pg_catalog.=) 'gotth_schema_migrations_sha256_length' AND constraint_value.contype OPERATOR(pg_catalog.=) 'c'
          AND constraint_value.convalidated
          AND pg_catalog.replace(
                  pg_catalog.replace(
                      pg_catalog.pg_get_constraintdef(constraint_value.oid, false),
                      'pg_catalog.octet_length', 'octet_length'),
                  'OPERATOR(pg_catalog.=)', '=')
              OPERATOR(pg_catalog.=) 'CHECK ((octet_length(sha256) = 32))')
    AND NOT EXISTS (
        SELECT 1
        FROM pg_catalog.pg_depend AS dependency
        JOIN pg_catalog.pg_constraint AS constraint_value
          ON dependency.classid OPERATOR(pg_catalog.=) 'pg_catalog.pg_constraint'::pg_catalog.regclass
         AND dependency.objid OPERATOR(pg_catalog.=) constraint_value.oid
        JOIN target ON target.oid OPERATOR(pg_catalog.=) constraint_value.conrelid
        LEFT JOIN pg_catalog.pg_proc AS referenced_function
          ON dependency.refclassid OPERATOR(pg_catalog.=) 'pg_catalog.pg_proc'::pg_catalog.regclass
         AND dependency.refobjid OPERATOR(pg_catalog.=) referenced_function.oid
        LEFT JOIN pg_catalog.pg_operator AS referenced_operator
          ON dependency.refclassid OPERATOR(pg_catalog.=) 'pg_catalog.pg_operator'::pg_catalog.regclass
         AND dependency.refobjid OPERATOR(pg_catalog.=) referenced_operator.oid
        JOIN pg_catalog.pg_namespace AS referenced_namespace
          ON referenced_namespace.oid OPERATOR(pg_catalog.=)
             COALESCE(referenced_function.pronamespace, referenced_operator.oprnamespace)
        WHERE referenced_namespace.nspname OPERATOR(pg_catalog.<>) 'pg_catalog')
    AND NOT EXISTS (
        SELECT 1
        FROM pg_catalog.pg_inherits AS inheritance
        JOIN target
          ON target.oid OPERATOR(pg_catalog.=) inheritance.inhrelid
          OR target.oid OPERATOR(pg_catalog.=) inheritance.inhparent)
    AND NOT EXISTS (
        SELECT 1 FROM pg_catalog.pg_trigger AS trigger_value JOIN target ON target.oid OPERATOR(pg_catalog.=) trigger_value.tgrelid
        WHERE NOT trigger_value.tgisinternal)
    AND NOT EXISTS (
        SELECT 1 FROM pg_catalog.pg_rewrite AS rewrite_value JOIN target ON target.oid OPERATOR(pg_catalog.=) rewrite_value.ev_class)`

type migrationRowQuerier interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

// attestHistoryTable proves the existing ledger matches the exact PostgreSQL
// 17 catalog shape created by this release before any history is trusted.
//
// Complexity: local time and space are tight Theta(1). Total time is the
// delegated fixed-catalog query cost C: O(C), Omega(1), with no tighter Theta
// bound because PostgreSQL catalog access is external.
func attestHistoryTable(ctx context.Context, querier migrationRowQuerier) error {
	if ctx == nil {
		return fmt.Errorf("migration attestation context is required")
	}
	if querier == nil {
		return fmt.Errorf("migration attestation connection is required")
	}
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("migration attestation canceled: %w", err)
	}
	var valid bool
	if err := querier.QueryRow(ctx, attestHistoryTableSQL).Scan(&valid); err != nil {
		return fmt.Errorf("attest migration history table: %w", err)
	}
	if !valid {
		return fmt.Errorf("migration history table does not match the release contract")
	}
	return nil
}
