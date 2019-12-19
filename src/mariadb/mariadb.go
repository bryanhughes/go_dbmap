package mariadb

import (
	"database/sql"
	"dbmap"
	"fmt"
	_ "github.com/lib/pq"
	"log"
)

const selectTables = "SELECT * FROM information_schema.tables WHERE table_schema = $1 AND table_name NOT IN ($2)"

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
    	a.attname AS column_name
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
    	t.relname,
    	i.relname;`

type Provider struct {
	dbmap.Config
}

func (provider *Provider) ReadDatabase() dbmap.Database {
	schemaNames := provider.Generator.Schemas
	schemas := make([]dbmap.Schema, len(schemaNames))

	db := initDB(provider)

	for i, schemaName := range schemaNames {
		schema := dbmap.Schema{SchemaName: schemaName}
		schemas[i] = schema
		readSchema(provider, &schema)
	}

	database := dbmap.Database{DB: db, Schemas: schemas}
	return database
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

	if err = db.Ping(); err != nil {
		log.Panic(err)
	}

	return db
}

func readSchema(provider *Provider, schema *dbmap.Schema) {
	fmt.Printf("[%s] %s\n", provider.Database.Provider, schema.SchemaName)
}
