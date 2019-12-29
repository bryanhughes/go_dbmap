package postgres

import (
	"database/sql"
	"dbmap"
	"testing"
	"time"
)

const TestConfig = "src/dbmap/postgres/test_config.yml"

var db *sql.DB = nil
var cfg dbmap.Config
var provider Provider

func setupTestCase(t *testing.T) func(t *testing.T) {
	if db == nil {
		dbmap.ReadFile(&cfg, TestConfig)

		t.Logf("Connecting to %s://user=%s:%s/%s\n", cfg.Database.Provider, cfg.Database.Username, cfg.Database.Host, cfg.Database.Database)

		dataSource := "host=" + cfg.Database.Host + " port=" + cfg.Database.Port +
			" user=" + cfg.Database.Username + " password=" + cfg.Database.Password +
			" dbname=" + cfg.Database.Database + " sslmode=disable"

		var err error
		db, err = sql.Open("postgres", dataSource)
		if err != nil {
			t.Fatal(err)
		}

		db.SetMaxOpenConns(5)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(time.Hour)

		if err = db.Ping(); err != nil {
			t.Fatal(err)
		}

		if cfg.Database.Provider == "postgres" {
			provider = Provider{cfg}
		} else {
			t.Fatalf("%s must be for a postgres database", TestConfig)
		}
	}

	return func(t *testing.T) {
		t.Log("teardown test case")
	}
}

func TestReadIndexes(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	table := dbmap.Table{
		TableName:   "example_a",
		TableSchema: "public",
		Columns:     nil,
		Indexes:     nil,
	}

	if err := readIndexes(db, &provider, &table); err != nil {
		t.Fatalf("Got an error ; %s", err)
	}

	// We are expecting 1 index
	if len(table.Indexes) != 3 {
		t.Fatal("Expected 3 indexes")
	}

	index := table.Indexes[0]
	if index.IndexName != "pk_example_a" {
		t.Fatal("Expected pk_example_a name")
	}

	if index.IndexType != "PRIMARY KEY" {

	}
}

func TestReadSchema(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	// Positive tests
	schema := dbmap.Schema{
		SchemaName: "test_schema",
		Tables:     nil,
	}
	if err := readTables(db, &provider, &schema); err != nil {
		t.Fatalf("Got an error ; %s", err)
	}

	// We are expecting 6 tables
	if len(schema.Tables) != 6 {
		t.Fatal("Expected 6 tables")
	}

	// Tables are ordered...
	if schema.Tables[0].TableName != "address" {
		t.Fatal("Did not get address")
	}

	if schema.Tables[1].TableName != "foo" {
		t.Fatal("Did not get foo")
	}

	if schema.Tables[2].TableName != "test_table_no_pkey" {
		t.Fatal("Did not get test_table_no_pkey")
	}

	if schema.Tables[3].TableName != "test_table_pkey" {
		t.Fatal("Did not get test_table_pkey")
	}

	if schema.Tables[4].TableName != "user" {
		t.Fatal("Did not get user")
	}

	if schema.Tables[5].TableName != "user_product_part" {
		t.Fatal("Did not get user_product_part")
	}

	// Negative testing, we are expecting errors
	if err := readTables(nil, &provider, &schema); err == nil {
		t.Fatalf("Got an error ; %s", err)
	} else if err != dbmap.InvalidArguments {
		t.Fatalf("Unexpected error: %s", err)
	}

	if err := readTables(db, nil, &schema); err == nil {
		t.Fatalf("Got an error ; %s", err)
	} else if err != dbmap.InvalidArguments {
		t.Fatalf("Unexpected error: %s", err)
	}

	if err := readTables(db, &provider, nil); err == nil {
		t.Fatalf("Got an error ; %s", err)
	} else if err != dbmap.InvalidArguments {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestReadDatabase(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	if database := provider.ReadDatabase(); database == nil {
		t.Fatal("Expected a non nil database")
	}
}

