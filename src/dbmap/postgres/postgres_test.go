package postgres

import (
	"database/sql"
	"github.com/bryanhughes/go_dbmap/src/dbmap"
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

func TestReadRelations(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	table := dbmap.Table{
		TableName:   "example_b",
		TableSchema: "public",
		Columns:     nil,
		Indexes:     nil,
		Relations:   nil,
	}

	if err := readForeignRelationships(db, &provider, &table); err != nil {
		t.Fatalf("Got an error ; %s", err)
	}

	// We are expecting 4 foreign relationships
	if len(table.Relations) != 4 {
		t.Fatal("Expected 4 relationships")
	}

	relation := table.Relations[0]
	if relation.ForeignSchema != "public" {
		t.Fatal("Expected public")
	}

	if relation.ForeignTable != "example_a" {
		t.Fatal("Expected example_a")
	}

	if len(relation.Columns) != 2 {
		t.Fatal("Expected 2 columns")
	}

	column := relation.Columns[0]
	if column.ForeignColumn != "column_a" {
		t.Fatal("Expected column_a")
	}

	if column.LocalColumn != "column_a" {
		t.Fatal("Expected column_a")
	}

	if column.OrdinalPosition != 1 {
		t.Fatal("Expected ordinal position 1")
	}

	column = relation.Columns[1]
	if column.ForeignColumn != "column_b" {
		t.Fatal("Expected column_b")
	}

	if column.LocalColumn != "column_b1" {
		t.Fatal("Expected column_b1")
	}

	if column.OrdinalPosition != 2 {
		t.Fatal("Expected ordinal position 1")
	}

	// Do a table with no foreign keys
	table = dbmap.Table{
		TableName:   "example_a",
		TableSchema: "public",
		Columns:     nil,
		Indexes:     nil,
		Relations:   nil,
	}

	if err := readForeignRelationships(db, &provider, &table); err != nil {
		t.Fatalf("Got an error ; %s", err)
	}

	// We are expecting no relationships
	if len(table.Relations) != 0 {
		t.Fatal("Expected no relationships")
	}

	// Negative testing, we are expecting errors
	if err := readForeignRelationships(nil, &provider, &table); err == nil {
		t.Fatalf("Got an error ; %s", err)
	} else if err != dbmap.InvalidArguments {
		t.Fatalf("Unexpected error: %s", err)
	}

	if err := readForeignRelationships(db, nil, &table); err == nil {
		t.Fatalf("Got an error ; %s", err)
	} else if err != dbmap.InvalidArguments {
		t.Fatalf("Unexpected error: %s", err)
	}

	if err := readForeignRelationships(db, &provider, nil); err == nil {
		t.Fatalf("Got an error ; %s", err)
	} else if err != dbmap.InvalidArguments {
		t.Fatalf("Unexpected error: %s", err)
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
		Relations:   nil,
	}

	if err := readIndexes(db, &provider, &table); err != nil {
		t.Fatalf("Got an error ; %s", err)
	}

	// We are expecting 3 index
	if len(table.Indexes) != 3 {
		t.Fatal("Expected 3 indexes")
	}

	index := table.Indexes[0]
	if index.IndexName != "idx_example_a" {
		t.Fatal("Expected idx_example_a name")
	}

	if index.IndexType != dbmap.NonUnique {
		t.Fatal("Expected non unique key")
	}

	if len(index.Columns) != 1 {
		t.Fatal("Expected 1 columns")
	}

	if index.Columns[0] != "column_c" {
		t.Fatal("Expected column_c")
	}

	index = table.Indexes[1]
	if index.IndexName != "pk_example_a" {
		t.Fatal("Expected pk_example_a name")
	}

	if index.IndexType != dbmap.PrimaryKey {
		t.Fatal("Expected primary key")
	}

	if len(index.Columns) != 2 {
		t.Fatal("Expected 2 columns")
	}

	if index.Columns[0] != "column_a" {
		t.Fatal("Expected column_a")
	}

	if index.Columns[1] != "column_b" {
		t.Fatal("Expected column_b")
	}

	index = table.Indexes[2]
	if index.IndexName != "unq_example_a" {
		t.Fatal("Expected unq_example_a name")
	}

	if index.IndexType != dbmap.Unique {
		t.Fatal("Expected unique key")
	}

	if len(index.Columns) != 1 {
		t.Fatal("Expected 1 columns")
	}

	if index.Columns[0] != "column_i" {
		t.Fatal("Expected column_i")
	}

	// Negative testing, we are expecting errors
	if err := readIndexes(nil, &provider, &table); err == nil {
		t.Fatalf("Got an error ; %s", err)
	} else if err != dbmap.InvalidArguments {
		t.Fatalf("Unexpected error: %s", err)
	}

	if err := readIndexes(db, nil, &table); err == nil {
		t.Fatalf("Got an error ; %s", err)
	} else if err != dbmap.InvalidArguments {
		t.Fatalf("Unexpected error: %s", err)
	}

	if err := readIndexes(db, &provider, nil); err == nil {
		t.Fatalf("Got an error ; %s", err)
	} else if err != dbmap.InvalidArguments {
		t.Fatalf("Unexpected error: %s", err)
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

	var database *dbmap.Database
	if database = provider.ReadDatabase(); database == nil {
		t.Fatal("Expected a non nil database")
	}

	if len(database.Schemas) != 2 {
		t.Fatal("Expected 2 schemas")
	}

	schema := database.Schemas[0]
	if schema.SchemaName != "public" {
		t.Fatal("Expecting public schema")
	}

	// We are excluding two of our tables
	if len(schema.Tables) != 8 {
		t.Fatal("Expected 8 tables in schema")
	}

	schema = database.Schemas[1]
	if schema.SchemaName != "test_schema" {
		t.Fatal("Expecting test_schema schema")
	}

	if len(schema.Tables) != 6 {
		t.Fatal("Expected 6 tables in schema")
	}
}

func TestIsColumnExcluded(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	column := dbmap.Column{
		TableName:       "user",
		TableSchema:     "test_schema",
		ColumnName:      "geog",
		OrdinalPosition: 5,
		DataType:        "geography",
		UdtName:         "geography",
		ColumnDefault:   "",
		IsNullable:      false,
		IsSequence:      false,
		IsPrimaryKey:    false,
	}

	if !isColumnExcluded(column, &provider) {
		t.Fatal("Column should be exclulded")
	}
}
