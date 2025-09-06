// test/integration/main_test.go
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

	"github.com/aarondl/sqlboiler/v4/boil"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	// migrate
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var testdb *sql.DB

func TestMain(m *testing.M) {
	ctx := context.Background()

	// 1) Start Postgres container
	req := testcontainers.ContainerRequest{
		Image:        "postgres:17",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		fmt.Printf("Cannot start container: %v\n", err)
		os.Exit(1)
	}
	// ensure terminated
	defer func() {
		_ = container.Terminate(ctx)
	}()

	// 2) Get mapped host & port
	host, _ := container.Host(ctx)
	p, _ := container.MappedPort(ctx, "5432")
	port := p.Port()

	// 3) Connect DB (pool)
	dsn := fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable", host, port)
	testdb, err = sql.Open("postgres", dsn)
	if err != nil {
		fmt.Printf("Cannot open DB: %v\n", err)
		os.Exit(1)
	}
	// small backoff wait for DB ready
	for i := 0; i < 10; i++ {
		if err = testdb.Ping(); err == nil {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if err != nil {
		fmt.Printf("Cannot ping DB: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Connected to ephemeral test DB:", dsn)

	// 4) Run migrations with golang-migrate (programmatically)
	root := findProjectRoot()
	migrationsPath := "file://" + filepath.Join(root, "migrations")
	// create migrate DB driver from existing *sql.DB
	driver, err := postgres.WithInstance(testdb, &postgres.Config{})
	if err != nil {
		fmt.Printf("migrate: cannot create driver: %v\n", err)
		os.Exit(1)
	}
	migr, err := migrate.NewWithDatabaseInstance(migrationsPath, "postgres", driver)
	if err != nil {
		fmt.Printf("migrate: cannot create migrator: %v\n", err)
		os.Exit(1)
	}
	if err := migr.Up(); err != nil && err != migrate.ErrNoChange {
		fmt.Printf("migrate: Up failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Migrations applied from:", migrationsPath)

	// 5) Ensure default search_path for new connections (affects new sessions)
	_, _ = testdb.Exec(`ALTER DATABASE testdb SET search_path TO public`)

	// Also set for current session (optional)
	_, _ = testdb.Exec(`SET search_path TO public`)

	// 6) Bind to SQLBoiler global DB handle
	boil.SetDB(testdb)
	// optional: boil.DebugMode = true

	// 7) Run tests
	code := m.Run()

	// 8) Cleanup
	_ = testdb.Close()
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

