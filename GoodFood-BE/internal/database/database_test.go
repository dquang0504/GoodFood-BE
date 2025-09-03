package database

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	_ "github.com/lib/pq"
)

var db *sql.DB

func TestMain(m *testing.M){
	var (
		dbName = os.Getenv("GOODFOOD_DB_DATABASE")
		dbPwd  = os.Getenv("GOODFOOD_DB_PASSWORD")
		dbUser = os.Getenv("GOODFOOD_DB_USERNAME")
		dbHost = os.Getenv("GOODFOOD_DB_HOST")
		dbPort = os.Getenv("GOODFOOD_DB_PORT")
	)

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPwd, dbHost, dbPort, dbName)

	var err error
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		fmt.Printf("❌ cannot open db: %v\n", err)
		os.Exit(1)
	}
	if err = db.Ping(); err != nil {
		fmt.Printf("❌ cannot ping db: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Connected to test database")

	code := m.Run()

	_ = db.Close()
	os.Exit(code)
}

func TestNew(t *testing.T) {
	srv := New()
	if srv == nil {
		t.Fatal("New() returned nil")
	}
}

func TestHealth(t *testing.T) {
	srv := New()

	stats := srv.Health()

	if stats["status"] != "up" {
		t.Fatalf("expected status to be up, got %s", stats["status"])
	}

	if _, ok := stats["error"]; ok {
		t.Fatalf("expected error not to be present")
	}

	if stats["message"] != "It's healthy" {
		t.Fatalf("expected message to be 'It's healthy', got %s", stats["message"])
	}
}

func TestClose(t *testing.T) {
	srv := New()

	if srv.Close() != nil {
		t.Fatalf("expected Close() to return nil")
	}
}
