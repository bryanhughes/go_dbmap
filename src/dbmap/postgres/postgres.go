package postgres

import (
	"database/sql"
	"dbmap"
	"fmt"
	"github.com/lib/pq"
	"log"
	"time"
)

const selectTables = "SELECT table_schema, table_name FROM information_schema.tables WHERE table_type = 'BASE TABLE' AND table_schema = $1 AND table_name NOT IN ($2) ORDER BY table_name"

const selectColumns = `SELECT
		c.column_name, 
		c.ordinal_position,
		c.data_type,
		c.udt_name::regtype::text,
		c.column_default,
		c.is_nullable,
		CASE WHEN pa.attname is null THEN false ELSE true END is_pkey,
		CASE WHEN pg_get_serial_sequence(table_schema || '.' || table_name, column_name) is null THEN false ELSE true END is_seq
	 FROM
		pg_namespace ns
		JOIN pg_class t ON
			t.relnamespace = ns.oid
			AND t.relkind = 'r'
			AND t.relname = $2
		JOIN information_schema.columns c ON
			c.table_schema = ns.nspname
			AND c.table_name = t.relname
		LEFT OUTER JOIN pg_index pi ON
			pi.indrelid = t.oid AND pi.indisprimary = true
		LEFT OUTER JOIN pg_attribute pa ON
			pa.attrelid = pi.indrelid
			AND pa.attnum = ANY(pi.indkey)
			AND pa.attname = c.column_name
	 WHERE
		ns.nspname = $1
	 ORDER BY table_schema, table_name, ordinal_position`

const selectIndexes = `SELECT
    i.relname AS index_name,
    a.attname AS column_name,
    tc.constraint_type AS constraint_type
FROM
    pg_class t,
    pg_class i,
    pg_index ix,
    pg_attribute a,
    pg_namespace ns,
    information_schema.table_constraints tc
WHERE
    t.oid = ix.indrelid
    AND i.oid = ix.indexrelid
    AND a.attrelid = t.oid
    AND a.attnum = ANY(ix.indkey)
    AND t.relkind = 'r'
    AND t.relname = $2
    AND t.relnamespace = ns.oid
    AND ns.nspname = $1
    AND tc.constraint_name = i.relname
    AND tc.table_name = t.relname
    AND tc.table_schema = ns.nspname
ORDER BY
    t.relname,
    i.relname`

type Provider struct {
	dbmap.Config
}

func (provider *Provider) ReadDatabase() *dbmap.Database {
	schemaNames := provider.Generator.Schemas
	schemas := make([]dbmap.Schema, len(schemaNames))

	db := initDB(provider)

	for i, schemaName := range schemaNames {
		schema := dbmap.Schema{SchemaName: schemaName}
		schemas[i] = schema
		if err := readTables(db, provider, &schema); err != nil {
			fmt.Printf("[FAILED] Reading schema %s - %s", schemaName, err)
			return nil
		}
	}

	database := dbmap.Database{DB: db, Schemas: schemas}
	return &database
}

func initDB(provider *Provider) *sql.DB {
	fmt.Printf("Connecting to %s://user=%s:%s/%s\n",
		provider.Database.Provider, provider.Database.Username, provider.Database.Host, provider.Database.Database)

	dataSource := "host=" + provider.Database.Host + " port=" + provider.Database.Port +
		" user=" + provider.Database.Username + " password=" + provider.Database.Password +
		" dbname=" + provider.Database.Database + " sslmode=disable"

	var err error
	db, err := sql.Open("postgres", dataSource)
	if err != nil {
		log.Panic(err)
	}

	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	if err = db.Ping(); err != nil {
		log.Panic(err)
	}

	return db
}

func readTables(db *sql.DB, provider *Provider, schema *dbmap.Schema) (err error) {
	if db == nil || provider == nil || schema == nil {
		return dbmap.InvalidArguments
	}
	fmt.Printf("[%s] %s", provider.Database.Provider, schema.SchemaName)

	rows, err := db.Query(selectTables, schema.SchemaName, pq.Array(provider.Generator.ExcludedTables))
	if err != nil {
		log.Print(err)
		return err
	}

	var tables []dbmap.Table
	for rows.Next() {
		table := dbmap.Table{}
		if err := rows.Scan(&table.TableSchema, &table.TableName); err != nil {
			fmt.Printf("[%s] FAILED reading tables in schema: %s\n", provider.Database.Provider, schema.SchemaName)
			return err
		}

		fmt.Printf("[%s] %s.%s", provider.Database.Provider, table.TableSchema, table.TableName)

		if err := readColumns(db, provider, &table); err != nil {
			fmt.Printf("[%s] FAILED reading columns for table: %s\n", provider.Database.Provider, table.TableName)
			return err
		}

		if err := readIndexes(db, provider, &table); err != nil {
			fmt.Printf("[%s] FAILED reading indexes for table: %s\n", provider.Database.Provider, table.TableName)
			return err
		}

		tables = append(tables, table)
	}
	schema.Tables = tables
	return nil
}

func readColumns(db *sql.DB, provider *Provider, table *dbmap.Table) (err error) {
	if db == nil || provider == nil || table == nil {
		return dbmap.InvalidArguments
	}

	rows, err := db.Query(selectColumns, table.TableSchema, table.TableName)
	if err != nil {
		log.Print(err)
		return err
	}

	var columns []dbmap.Column
	for rows.Next() {
		column := dbmap.Column{}
		if err := rows.Scan(&column.ColumnName, &column.OrdinalPosition, &column.DataType, &column.UdtName,
			&column.ColumnDefault, &column.IsNullable, &column.IsPrimaryKey, &column.IsSequence); err != nil {
			fmt.Printf("[%s] FAILED reading columns for table: %s\n", provider.Database.Provider, table.TableName)
			return err
		}
		columns = append(columns, column)
	}
	table.Columns = columns
	return nil
}

func readIndexes(db *sql.DB, provider *Provider, table *dbmap.Table) (err error) {
	if db == nil || provider == nil || table == nil {
		return dbmap.InvalidArguments
	}

	rows, err := db.Query(selectIndexes, table.TableSchema, table.TableName)
	if err != nil {
		log.Print(err)
		return err
	}

	var index dbmap.Index
	var indexes []dbmap.Index
	var indexName string
	var indexType string
	var workingIndex string
	var columns []string
	var columnName string
	for rows.Next() {
		if err := rows.Scan(&indexName, &columnName, &indexType); err != nil {
			fmt.Printf("[%s] FAILED reading indexes for table: %s\n", provider.Database.Provider, table.TableName)
			return err
		}

		// New index, so add the working and start a new one
		if indexName != workingIndex {
			index.Columns = columns
			indexes = append(indexes, index)

			index = dbmap.Index{}
			index.TableSchema = table.TableSchema
			index.TableName = table.TableName
			index.IndexName = indexName
			index.IndexType = indexType
			columns = make([]string, 1)
		}
		columns = append(columns, columnName)
	}
	index.Columns = columns
	indexes = append(indexes, index)

	table.Indexes = indexes
	return nil
}