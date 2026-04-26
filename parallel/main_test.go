package parallel

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var (
	infra *TestInfra
)

type TestInfra struct {
	PgContainer *postgres.PostgresContainer
	BaseConnStr string
}

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

// Helper untuk buat database baru per test sehingga test bisa jalan parallel dan tidak bentrok data datanya
func SetupTestDatabase(t *testing.T) *pgxpool.Pool {
	t.Log(">>>>> Setup test database")

	ctx := context.Background()
	var err error

	dbName := fmt.Sprintf("test_db_%d", time.Now().UnixNano())

	tempPool, _ := pgxpool.New(ctx, infra.BaseConnStr)
	defer tempPool.Close()

	_, err = tempPool.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", dbName))
	if err != nil {
		t.Fatal(err)
	}

	newConnStr := strings.Replace(infra.BaseConnStr, "/test_db", "/"+dbName, 1)

	RunMigrations(t, newConnStr)

	pool, _ := pgxpool.New(ctx, newConnStr)

	// hanya dijalankan ketika test atau sub test yang panggil function ini tuh selesai.
	t.Cleanup(func() {
		pool.Close()

		tempPool, err := pgxpool.New(ctx, infra.BaseConnStr)
		if err != nil {
			// tidak t.Fatal karena kalau testnya sukses tapi ketika ingin cleanup itu gagal
			// maka test tetap dianggap sukses dan juga toh diakhir test containernya akan dihapus
			// dan ketika testnya dijalanin ulang akan buat container baru
			t.Logf("cleanup: failed to connect to db: %v", err)
		}
		defer tempPool.Close()

		// pakai force agar tidak error cannot drop database because its being accessed by other users
		// ingat 1 test itu seharusnya 1 database sendiri di postgres
		_, err = tempPool.Exec(ctx, fmt.Sprintf("DROP DATABASE %s WITH (FORCE)", dbName))
		if err != nil {
			t.Logf("cleanup: failed to drop db %s: %v", dbName, err)
		}
	})

	t.Log(">>>>> Setup test database finished")

	return pool
}

func RunMigrations(t *testing.T, connStr string) {
	t.Log(">>>>> Start running migrations")

	migrationPath := "file://migrations"

	m, err := migrate.New(migrationPath, connStr)
	if err != nil {
		t.Fatalf("failed to create migrate instance: %v", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		t.Fatalf("failed to run migrations: %v", err)
	}

	t.Log(">>>>> Finished running migrations")
}

func run(m *testing.M) int {
	ctx := context.Background()
	var err error

	pgContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		// perlu diisi karena kita tidak bisa create database kalau belum connect ke database manapun
		postgres.WithDatabase("test_db"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		postgres.BasicWaitStrategies(),
	)

	if err != nil {
		panic(err)
	}

	defer func() {
		err = testcontainers.TerminateContainer(pgContainer)
		if err != nil {
			panic(err)
		}
	}()

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic(err)
	}

	infra = &TestInfra{
		PgContainer: pgContainer,
		BaseConnStr: connStr,
	}

	return m.Run()
}
