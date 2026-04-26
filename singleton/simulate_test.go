package singleton

import (
	"context"
	"testing"
)

func TestQuerySelect1(t *testing.T) {
	var n int
	err := dbPool.QueryRow(context.Background(), "SELECT 1").Scan(&n)
	if err != nil {
		t.Fatal(err)
	}

	if n != 1 {
		t.Fatalf("expected 1, got %d", n)
	}

	t.Logf("Sucessfully query to postgres, result: %d", n)
}

func TestQuerySelect2(t *testing.T) {
	var n int
	err := dbPool.QueryRow(context.Background(), "SELECT 2").Scan(&n)
	if err != nil {
		t.Fatal(err)
	}

	if n != 2 {
		t.Fatalf("expected 2, got %d", n)
	}

	t.Logf("Sucessfully query to postgres, result: %d", n)
}
