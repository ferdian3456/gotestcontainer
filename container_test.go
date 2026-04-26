package gotestcontainer_test

import (
	"context"
	"fmt"
	"net/http"
	"net/smtp"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	tcminio "github.com/testcontainers/testcontainers-go/modules/minio"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

func TestPostgresContainer(t *testing.T) {
	ctx := context.Background()
	var err error
	pgContainer, err := postgres.Run(ctx,
		"postgres:15-alpine", // nama image yang akan digunakan
		postgres.WithDatabase("test_db"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		// wait postgresnya sampai postgres output database system is ready .. 2 kali
		// tiap database punya cara wait yang beda beda sampai containernya siap digunakan
		postgres.BasicWaitStrategies(),
		// init script
		// postgres.WithInitScripts("file:///home/ferdian/Documents/virdan/gotestcontainer/init.sql"),
	)

	if err != nil {
		t.Fatal(err) // t fatal akan menghentikan testnya dan print errornya
	}

	defer func() {
		//
		err = testcontainers.TerminateContainer(pgContainer)
		if err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}()

	// dapatin connection string postgresnya, karena test container ketika jalanin
	// containernya(postgresnya) itu selalu diassign ke port yang beda beda tiap kali
	// jalanin testnya misal 12345, nanti test kedua 54321,lalu 99999,dst
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Connection string: %s", connStr)

	// untuk connect ke db dengan conn string yang sudah didapat sebelumnya
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatal(err)
	}

	defer pool.Close()

	var result int
	err = pool.QueryRow(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		t.Fatal(err)
	}

	if result != 1 {
		t.Fatalf("expected 1, got %d", result)
	}

	t.Logf("Sucessfully query to postgres, result: %d", result)
}

func TestRedisContainer(t *testing.T) {
	ctx := context.Background()
	var err error

	rdsContainer, err := tcredis.Run(ctx,
		"redis:7-alpine",
	)

	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err = testcontainers.TerminateContainer(rdsContainer)
		if err != nil {
			t.Logf("failed to terminate container %v", err)
		}
	}()

	// Gk bisa gini karena redis.client itu expect addr seperti localhost:9090
	// bukan redis://localhost:9090
	// connStr, err := rdsContainer.ConnectionString(ctx)
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// dapatin host dan portnya
	host, err := rdsContainer.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// cari tahu extrernal port berapa yang di assign ke port 6379 (docker internal)
	port, err := rdsContainer.MappedPort(ctx, "6379")
	if err != nil {
		t.Fatal(err)
	}

	connStr := fmt.Sprintf("%s:%s", host, port.Port())

	t.Logf("Connection string: %s", connStr)

	rdb := redis.NewClient(&redis.Options{
		Addr: connStr,
	})

	defer rdb.Close()

	err = rdb.Set(ctx, "username", "ferdian", 0).Err()
	if err != nil {
		t.Fatal(err)
	}

	val, err := rdb.Get(ctx, "username").Result()
	if err != nil {
		t.Fatal(err)
	}

	if val != "ferdian" {
		t.Fatalf("expected 'ferdian', got %s", val)
	}

	t.Logf("Successfully query to redis, result: %s", val)
}

func TestMinioContainer(t *testing.T) {
	ctx := context.Background()
	var err error

	minioContainer, err := tcminio.Run(ctx,
		"minio/minio:RELEASE.2024-01-16T16-07-38Z",
	)

	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err = testcontainers.TerminateContainer(minioContainer)
		if err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}()

	connStr, err := minioContainer.ConnectionString(ctx)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Connection string: %s", connStr)

	mc, err := minio.New(connStr, &minio.Options{
		Creds:  credentials.NewStaticV4("minioadmin", "minioadmin", ""),
		Secure: false,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = mc.MakeBucket(ctx, "my-bucket", minio.MakeBucketOptions{})
	if err != nil {
		// ketika jalanin test container minio itu tidak mungkin bucketnya sudah ada
		// karena terminate container itu bakal menghapus container dan volumenya
		t.Fatal(err)
	}

	t.Logf("Sucessfully create bucket: my-bucket")
}

func TestMailHogContainer(t *testing.T) {
	ctx := context.Background()
	var err error

	// pakai generic container karena belum disupport sama test container
	mailhogContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "mailhog/mailhog:latest",
			ExposedPorts: []string{"1025/tcp", "8025/tcp"},
			WaitingFor:   wait.ForListeningPort("1025/tcp"),
		},
		Started: true, // auto start
	})

	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err = testcontainers.TerminateContainer(mailhogContainer)
		if err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}()

	host, err := mailhogContainer.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// cari tahu extrernal port berapa yang di assign ke port 1025 (docker internal)
	smtpPort, err := mailhogContainer.MappedPort(ctx, "1025")
	if err != nil {
		t.Fatal(err)
	}

	apiPort, err := mailhogContainer.MappedPort(ctx, "8025")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("MailHog Host: %s", host)
	t.Logf("MailHog API Port: %s", apiPort.Port())
	t.Logf("MailHog SMTP Port: %s", smtpPort.Port())

	smtpAddr := fmt.Sprintf("%s:%s", host, smtpPort.Port())
	apiURL := fmt.Sprintf("http://%s:%s/api/v1/messages", host, apiPort.Port())

	t.Logf("MailHog SMTP Address: %s", smtpAddr)
	t.Logf("MailHog API URL: %s", apiURL)

	msg := []byte("Subject: Halo Testcontainers\r\n\r\nIni adalah isi email test.")
	err = smtp.SendMail(smtpAddr, nil, "pengirim@test.com", []string{"penerima@test.com"}, msg)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.Get(apiURL)
	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	t.Logf("Sucessfully send email to MailHog")
}
