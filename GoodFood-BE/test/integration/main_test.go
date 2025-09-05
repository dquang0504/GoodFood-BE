package integration

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var db *sql.DB

func TestMain(m *testing.M) {
	ctx := context.Background()

	// 1. Khởi tạo Postgres ephemeral container
	req := testcontainers.ContainerRequest{
		Image:        "postgres:17",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		fmt.Printf("Cannot start container: %v\n", err)
		os.Exit(1)
	}

	defer container.Terminate(ctx)

	// 2. Lấy port động container
	host, _ := container.Host(ctx)
	p, _ := container.MappedPort(ctx, "5432")
	port := p.Port()

	// 3. Kết nối DB
	dsn := fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable", host, port)
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		fmt.Printf("Cannot open DB: %v\n", err)
		os.Exit(1)
	}

	if err = db.Ping(); err != nil {
		fmt.Printf("Cannot ping DB: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Connected to ephemeral test DB")

	// 4. Chạy migration schema
	root := findProjectRoot()
	schemaPath := filepath.Join(root, "migrations", "0001_init.up.sql")
	fmt.Println("Using migration path:", schemaPath)

	schemaSQL, err := os.ReadFile(schemaPath)
	if err != nil {
		fmt.Printf("Cannot read migration: %v\n", err)
		os.Exit(1)
	}

	_, err = db.Exec(string(schemaSQL))
	if err != nil {
		fmt.Printf("Cannot exec migration: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Schema migration applied")

	// 5. Chạy tất cả test
	code := m.Run()

	// 6. Close connection
	db.Close()

	os.Exit(code)
}

// findProjectRoot tìm go.mod để xác định root
func findProjectRoot() string {
	wd, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			panic("could not find go.mod")
		}
		wd = parent
	}
}

