package dbmap

import (
	"database/sql"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

var InvalidArguments = errors.New("invalid argument")

// The expected YAML configuration for generating code based on a database schema
type Config struct {
	Database struct {
		Provider string `yaml:"provider"`
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		Database string `yaml:"database"`
		Username string `yaml:"user"`
		Password string `yaml:"password"`
	} `yaml:"database"`
	Output struct {
		Path   string `yaml:"path"`
		Suffix string `yaml:"suffix"`
		Lang   string `yaml:"lang"`
	} `yaml:"output"`
	Proto struct {
		Path        string `yaml:"path"`
		JavaPackage string `yaml:"java_package"`
		Version     string `yaml:"version"`
	} `yaml:"proto"`
	Generator struct {
		Schemas         []string `yaml:"schemas"`
		ExcludedTables  []string `yaml:"excluded_tables"`
		IndexedLookups  bool     `yaml:"indexed_lookups"`
		ExcludedColumns []struct {
			Tablename string   `yaml:"table"`
			Columns   []string `yaml:"columns"`
		} `yaml:"excluded_columns"`
		Mapping []struct {
			Tablename string `yaml:"table"`
			Queries   []struct {
				Name  string `yaml:"name"`
				Query string `yaml:"query"`
			} `yaml:"queries"`
		} `yaml:"mapping"`
		Transforms []struct {
			Tablename string `yaml:"table"`
			Xforms    struct {
				Select []struct {
					Columnname string `yaml:"column"`
					Datatype   string `yaml:"data_type"`
					Xform      string `yaml:"xform"`
				} `yaml:"select"`
				Insert []struct {
					Columnname string `yaml:"column"`
					Datatype   string `yaml:"data_type"`
					Xform      string `yaml:"xform"`
				}
				Update []struct {
					Columnname string `yaml:"column"`
					Datatype   string `yaml:"data_type"`
					Xform      string `yaml:"xform"`
				}
			} `yaml:"xforms"`
		} `yaml:"transforms"`
	} `yaml:"generator"`
}

// The structure of an index needed to generate lookup functions based on alternate keys and indexes
type Index struct {
	TableSchema string
	TableName   string
	IndexName   string
	IndexType   string
	Columns     []string
}

// The structure of a column
type Column struct {
	TableName       string
	TableSchema     string
	ColumnName      string
	OrdinalPosition int
	DataType        string
	UdtName         string
	ColumnDefault   string
	IsNullable      bool
	IsSequence      bool
	IsPrimaryKey    bool
}

// The structure of a table
type Table struct {
	TableName   string
	TableSchema string
	Columns     []Column
	Indexes     []Index
}

// The structure of a schema
type Schema struct {
	SchemaName string
	Tables     []Table
}

// The current database and the schema's we will generate code against
type Database struct {
	DB         *sql.DB
	Schemas		[]Schema
}

type Provider interface {
	ReadDatabase() *Database
}

func ReadFile(cfg *Config, configFile string) {
	fmt.Println("Using configuration: ", configFile)
	f, err := os.Open(configFile)
	if err != nil {
		processError(err)
	}

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		processError(err)
	}
}

func processError(err error) {
	fmt.Println(err)
	os.Exit(2)
}

func ReadResults(rows *sql.Rows, err error) (results []map[string]interface{}, err_out error) {
	if err != nil {
		log.Print(err)
		return nil, err
	}
	defer rows.Close()

	results = make([]map[string]interface{}, 0)
	cols, _ := rows.Columns()
	for rows.Next() {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			log.Print(err)
			return nil, err
		}

		if err := rows.Err(); err != nil {
			log.Print(err)
			return nil, err
		}

		m := make(map[string]interface{})
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			m[colName] = *val
		}
		results = append(results, m)
	}

	return results, nil
}

func GenerateCode(database *Database) {

}