package parallel

import (
	"context"
	"testing"
)

func TestUserA(t *testing.T) {
	t.Parallel() // <--- Jalanin secara parallel biar cepat

	dbPool := SetupTestDatabase(t) // Dapet database unik 1
	ctx := context.Background()

	// Insert data di database 1
	_, err := dbPool.Exec(ctx, "INSERT INTO users (name) VALUES ('Ferdian')")
	if err != nil {
		t.Fatal(err)
	}

	// Cek jumlah data, harusnya cuma 1
	var count int
	dbPool.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&count)

	if count != 1 {
		t.Fatalf("TestUserA: expected 1 user, got %d", count)
	}
}

func TestUserB(t *testing.T) {
	t.Parallel() // <--- Jalan barengan sama TestUserA

	dbPool := SetupTestDatabase(t) // Dapet database unik 2
	ctx := context.Background()

	// Insert data di database 2
	_, err := dbPool.Exec(ctx, "INSERT INTO users (name) VALUES ('Budi')")
	if err != nil {
		t.Fatal(err)
	}

	// Cek jumlah data, harusnya tetep cuma 1 (Data Ferdian tidak boleh kelihatan di sini)
	var count int
	dbPool.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&count)

	if count != 1 {
		t.Fatalf("TestUserB: expected 1 user, got %d", count)
	}
}

func TestUserC(t *testing.T) {
	t.Parallel()

	dbPool := SetupTestDatabase(t)
	ctx := context.Background()

	_, err := dbPool.Exec(ctx, "INSERT INTO users (name) VALUES ('anton')")
	if err != nil {
		t.Fatal(err)
	}

	var count int
	_ = dbPool.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&count)

	if count != 1 {
		t.Fatalf("TestUserC: expected 1 user, got %d", count)
	}
}
