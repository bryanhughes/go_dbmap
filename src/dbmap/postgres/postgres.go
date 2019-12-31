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
		CASE WHEN c.is_nullable is null THEN false ELSE true END is_nullable,
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
    ix.indisunique,
    ix.indisprimary
FROM
    pg_class t,
    pg_class i,
    pg_index ix,
    pg_attribute a,
    pg_namespace ns
WHERE
    t.oid = ix.indrelid
    AND i.oid = ix.indexrelid
    AND a.attrelid = t.oid
    AND a.attnum = ANY(ix.indkey)
    AND t.relkind = 'r'
    AND t.relname = $2
    AND t.relnamespace = ns.oid
    AND ns.nspname = $1
ORDER BY
    i.relname`

const selectForeignRelationships = `SELECT DISTINCT
	f_kcu.table_schema AS foreign_schema,
	f_kcu.table_name AS foreign_table,
	f_kcu.column_name AS foreign_column,
	kcu.column_name AS local_column,
	f_kcu.ordinal_position
FROM
	information_schema.key_column_usage kcu
	JOIN information_schema.referential_constraints rc ON
	    rc.constraint_schema = kcu.constraint_schema
	    AND rc.constraint_name = kcu.constraint_name
	JOIN information_schema.key_column_usage f_kcu ON
	    f_kcu.constraint_schema = rc.unique_constraint_schema
	    AND f_kcu.constraint_name = rc.unique_constraint_name
	    AND f_kcu.ordinal_position = kcu.position_in_unique_constraint
WHERE
	kcu.table_schema = $1
	AND kcu.table_name = $2
	AND kcu.position_in_unique_constraint IS NOT NULL
ORDER BY
	foreign_schema, foreign_table`

type Provider struct {
	dbmap.Config
}

func (provider *Provider) ReadDatabase() *dbmap.Database {
	schemaNames := provider.Generator.Schemas
	schemas := make([]dbmap.Schema, len(schemaNames))

	db := initDB(provider)

	var schema dbmap.Schema
	for i, schemaName := range schemaNames {
		schema = dbmap.Schema{SchemaName: schemaName}
		if err := readTables(db, provider, &schema); err != nil {
			fmt.Printf("[FAILED] Reading schema %s - %s", schemaName, err)
			return nil
		}
		schemas[i] = schema
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
	fmt.Printf("[%s] %s\n", provider.Database.Provider, schema.SchemaName)

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

		if isTableExcluded(table, provider) {
			fmt.Printf("[%s] %s.%s (excluding)\n", provider.Database.Provider, table.TableSchema, table.TableName)
		} else {
			fmt.Printf("[%s] %s.%s\n", provider.Database.Provider, table.TableSchema, table.TableName)

			if err := readColumns(db, provider, &table); err != nil {
				fmt.Printf("[%s] FAILED reading columns for table: %s\n", provider.Database.Provider, table.TableName)
				return err
			}

			if err := readIndexes(db, provider, &table); err != nil {
				fmt.Printf("[%s] FAILED reading indexes for table: %s\n", provider.Database.Provider, table.TableName)
				return err
			}

			if err := readForeignRelationships(db, provider, &table); err != nil {
				fmt.Printf("[%s] FAILED reading foreign relationships for table: %s\n",
					provider.Database.Provider, table.TableName)
				return err
			}

			tables = append(tables, table)
		}
	}
	schema.Tables = tables
	return nil
}

func isTableExcluded(table dbmap.Table, provider *Provider) bool {
	for _, tableName := range provider.Generator.ExcludedTables {
		if table.TableName == tableName {
			return true
		}
	}
	return false
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

	var columnDefault sql.NullString
	var columns []dbmap.Column
	for rows.Next() {
		column := dbmap.Column{}
		column.TableSchema = table.TableSchema
		column.TableName = table.TableName

		if err := rows.Scan(&column.ColumnName, &column.OrdinalPosition, &column.DataType, &column.UdtName,
			&columnDefault, &column.IsNullable, &column.IsPrimaryKey, &column.IsSequence); err != nil {
			fmt.Printf("[%s] FAILED reading columns for table: %s\n", provider.Database.Provider, table.TableName)
			return err
		}

		if isColumnExcluded(column, provider) {
			fmt.Printf("   Excluding column: %s\n", column.ColumnName)
		} else {
			if columnDefault.Valid {
				column.ColumnDefault = columnDefault.String
			}
			columns = append(columns, column)
		}
	}
	table.Columns = columns
	return nil
}

func isColumnExcluded(column dbmap.Column, provider *Provider) bool {
	for _, excludedColumn := range provider.Generator.ExcludedColumns {
		if excludedColumn.Tablename == column.TableSchema + "." + column.TableName {
			for _, c := range excludedColumn.Columns {
				if c == column.ColumnName {
					return true
				}
			}
		}
	}
	return false
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
	var isUnique bool
	var isPrimaryKey bool
	var workingIndex string
	var columns []string
	var columnName string
	var firstTime = true
	for rows.Next() {
		if err := rows.Scan(&indexName, &columnName, &isUnique, &isPrimaryKey); err != nil {
			fmt.Printf("[%s] FAILED reading indexes for table: %s\n", provider.Database.Provider, table.TableName)
			return err
		}

		// New index, so add the working and start a new one
		if indexName != workingIndex {
			index.Columns = columns
			if firstTime {
				firstTime = false
			} else {
				indexes = append(indexes, index)
			}

			index = dbmap.Index{TableSchema: table.TableSchema, TableName: table.TableName, IndexName: indexName}
			if isPrimaryKey {
				index.IndexType = dbmap.PrimaryKey
			} else if isUnique {
				index.IndexType = dbmap.Unique
			} else {
				index.IndexType = dbmap.NonUnique
			}

			columns = make([]string, 0)
		}
		workingIndex = indexName
		columns = append(columns, columnName)
	}
	if ! firstTime {
		index.Columns = columns
		indexes = append(indexes, index)
	}
	table.Indexes = indexes
	return nil
}

func readForeignRelationships(db *sql.DB, provider *Provider, table *dbmap.Table) (err error) {
	if db == nil || provider == nil || table == nil {
		return dbmap.InvalidArguments
	}

	rows, err := db.Query(selectForeignRelationships, table.TableSchema, table.TableName)
	if err != nil {
		log.Print(err)
		return err
	}

	var relation dbmap.ForeignRelation
	var relations []dbmap.ForeignRelation
	var fSchema string
	var fTable string
	var fColumn string
	var lColumn string
	var columns []dbmap.ForeignColumns
	var oPos int32
	var workingTable string
	var workingSchema string
	var firstTime = true
	for rows.Next() {
		if err := rows.Scan(&fSchema, &fTable, &fColumn, &lColumn, &oPos); err != nil {
			fmt.Printf("[%s] FAILED reading indexes for table: %s\n", provider.Database.Provider, table.TableName)
			return err
		}

		if fTable != workingTable || fSchema != workingSchema {
			relation.Columns = columns
			if firstTime {
				firstTime = false
			} else {
				relations = append(relations, relation)
			}

			relation = dbmap.ForeignRelation{
				ForeignSchema: fSchema,
				ForeignTable:  fTable,
				Columns:       nil,
				RelationType:  dbmap.ZeroOneOrMore,
			}

			columns = make([]dbmap.ForeignColumns, 0)
		}
		workingTable = fTable
		workingSchema = fSchema
		columns = append(columns, dbmap.ForeignColumns{
			ForeignColumn:   fColumn,
			LocalColumn:     lColumn,
			OrdinalPosition: oPos,
		})
	}
	if ! firstTime {
		relation.Columns = columns
		relations = append(relations, relation)
	}
	table.Relations = relations
	return nil
}
