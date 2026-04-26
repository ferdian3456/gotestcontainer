package singleton

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var dbPool *pgxpool.Pool

// TestMain atau testing.M itu ketika kita jalanin testnya maka file ini yang akan dijalnin dahulu sebelum testnya jalan
// Jadi cocok untuk setup database yang akan digunakan bersamaan oleh semua test
// Jadinya tidak setiap test akan buat dan jalanin container baru.
func TestMain(m *testing.M) {
	// kenapa buat function run untuk jalanin singletonnya bukan di TestMain ya karena kalau
	// ada panic di run itu bisa didefer dengan close db connection lalu terminate containernya
	// kalau di TestMain itu kan ada proses os.Exit, os.Exit dengan defer itu gk bisa
	os.Exit(run(m))
}

func run(m *testing.M) int {
	ctx := context.Background()
	var err error

	// tidak pakai semacam t.Logf karena ini bukan file testnya tapi file mainnya
	log.Println(">>>>> Bootstrapping Singleton Postgres")

	pgContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
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
		// tidak pakai semacam t.Fatal karena ini kan file main bukan test
		// yakali mau t.Fatal
		panic(err)
	}

	log.Println("Connection string: ", connStr)

	dbPool, err = pgxpool.New(ctx, connStr)
	if err != nil {
		panic(err)
	}

	defer dbPool.Close()

	// jalankan semua test di dalam 1 folder ini, return 0 kalau success semua testnya
	code := m.Run()

	log.Println(">>>>> Stopping Singleton Postgres")

	return code
}
