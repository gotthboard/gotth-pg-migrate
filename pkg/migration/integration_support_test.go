//go:build integration

package migration

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

func requirePostgreSQL17(t *testing.T, ctx context.Context, connection *pgx.Conn) {
	t.Helper()
	var serverVersion int
	if err := connection.QueryRow(ctx, "SELECT current_setting('server_version_num')::integer").Scan(&serverVersion); err != nil {
		t.Fatalf("query PostgreSQL version: %v", err)
	}
	if serverVersion < 170000 || serverVersion >= 180000 {
		t.Fatalf("PostgreSQL server_version_num = %d, want major version 17", serverVersion)
	}
}
