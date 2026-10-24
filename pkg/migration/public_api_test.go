package migration_test

import (
	"context"
	"errors"
	"io/fs"
	"testing"
	"testing/fstest"

	migration "github.com/gotthboard/gotth-pg-migrate/pkg/migration"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	_ migration.VerificationQuerier                       = (*pgx.Conn)(nil)
	_ migration.VerificationQuerier                       = (*pgxpool.Pool)(nil)
	_ func(context.Context, *pgx.ConnConfig, fs.FS) error = migration.Apply
	_ interface {
		Verify(context.Context, migration.VerificationQuerier) error
	} = (*migration.ReleaseVerifier)(nil)
)

func TestExternalConsumerCanConstructReleaseVerifier(t *testing.T) {
	t.Parallel()

	verifier, err := migration.NewReleaseVerifier(fstest.MapFS{
		"000001_initial.sql": {Data: []byte("SELECT 1;")},
	})
	if err != nil || verifier == nil {
		t.Fatalf("NewReleaseVerifier() = (%v, %v)", verifier, err)
	}
	if migration.ErrCommitOutcomeUnknown == nil || !errors.Is(migration.ErrCommitOutcomeUnknown, migration.ErrCommitOutcomeUnknown) {
		t.Fatal("ErrCommitOutcomeUnknown lost stable error identity")
	}
}
