//go:build integration

package postgres

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	ctr, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase("test"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		log.Fatalf("start container: %v", err)
	}
	dsn, err := ctr.ConnectionString(ctx, "sslmode=disable")

	if err != nil {
		log.Fatalf("connection string: %v", err)
	}
	testPool, err = NewDB(ctx, dsn)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	applyMigration(ctx)

	code := m.Run()
	testPool.Close()
	ctr.Terminate(ctx)
	os.Exit(code)
}

func applyMigration(ctx context.Context) {
	files, err := filepath.Glob("../../../migrations/*.up.sql")
	if err != nil {
		log.Fatalf("glop migrations: %v", err)
	}
	sort.Strings(files)

	for _, f := range files {
		sql, err := os.ReadFile(f)
		if err != nil {
			log.Fatalf("read %s:%v", f, err)
		}
		if _, err := testPool.Exec(ctx, string(sql)); err != nil {
			log.Fatalf("apply %s: %v", f, err)
		}
	}

}

func resetTables(t *testing.T) {
	t.Helper()
	if _, err := testPool.Exec(t.Context(), "TRUNCATE projects, systems, segments CASCADE"); err != nil {
		t.Fatalf("reset tables: %v", err)
	}
}
