package dbmap

import (
	"fmt"
	"github.com/iancoleman/strcase"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func GenerateProto(cfg Config, database *Database) error {
	for _, schema := range database.Schemas {
		dir, _ := os.Getwd()
		path := filepath.Join(dir, cfg.Proto.Path, schema.SchemaName)
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			fmt.Printf("FAILED to create output path with permission 0755 - %s : %s\n", path, err)
			return err
		}

		for _, table := range schema.Tables {
			fmt.Printf("%s/%s.proto\n", table.TableSchema, table.TableName)
			if err := writeProto(cfg, table); err != nil {
				fmt.Printf("Failed to write proto for table %s in path %s : %s\n", table.TableName, path, err)
				return err
			}
		}
	}

	return nil
}

func writeProto(cfg Config, table Table) error {
	filename := filepath.Join(cfg.Proto.Path, table.TableSchema, table.TableName+".proto")
	f, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Failed to create file %s : %s\n", filename, err)
		return err
	}
	defer f.Close()

	_, _ = fmt.Fprintf(f, "//-------------------------------------------------------------------\n")
	_, _ = fmt.Fprintf(f, "// This file is automatically generated from the database schema.\n")
	_, _ = fmt.Fprintf(f, "// ---- DO NOT MAKE CHANGES DIRECTLY TO THIS FILE! ----\n\n")

	_, _ = fmt.Fprintf(f, "syntax = \"%s\";\n\n", cfg.Proto.Version)

	_, _ = fmt.Fprintf(f, "package %s;\n\n", table.TableSchema)
	_, _ = fmt.Fprintf(f, "option cc_enable_arenas = true;\n")
	_, _ = fmt.Fprintf(f, "option java_package = \"%s.%s\";\n", cfg.Proto.JavaPackage, table.TableSchema)
	_, _ = fmt.Fprintf(f, "option java_outer_classname = \"%sProto\";\n", table.TableName)
	_, _ = fmt.Fprintf(f, "option objc_class_prefix = \"%s\";\n\n", cfg.Proto.ObjCPrefix)

	if cfg.EmbedRelationships {
		writeImports(f, table)
	}

	maybeWriteOtherImports(f, table)

	_, _ = fmt.Fprintf(f, "message %s {\n", strcase.ToCamel(table.TableName))

	writeFields(f, cfg, table)

	_, _ = fmt.Fprint(f, "}\n")

	return nil
}

func maybeWriteOtherImports(f *os.File, table Table) {
	tsFlag := false
	commentFlag := false
	for _, column := range table.Columns {
		if strings.HasPrefix(column.UdtName, "timestamp") && ! tsFlag {
			tsFlag = true
			if ! commentFlag {
				commentFlag = true
				_, _ = fmt.Fprintf(f, "// Other Datatype Imports\n\n")
			}

			_, _ = fmt.Fprintf(f, "import \"google/protobuf/timestamp.proto\";\n")
		}
	}

	if commentFlag {
		_, _ = fmt.Fprintf(f, "\n")
	}
}

func writeImports(f *os.File, table Table) {
	if len(table.Relations) == 0 {
		return
	}
	_, _ = fmt.Fprintf(f, "// Foreign Key Imports\n\n")

	for _, relation := range table.Relations {
		_, _ = fmt.Fprintf(f, "import \"%s/%s.proto\";\n", relation.ForeignSchema, relation.ForeignTable)
	}
	_, _ = fmt.Fprint(f, "\n")
}

func writeFields(f *os.File, cfg Config, table Table) {
	// If we are writing the protos with embedded messages, we need to build a column map that will handle the use
	// cases: 1) two tables with the same name from different schemas, and 2) two of the same tables with different
	// referencing column names
	if cfg.EmbedRelationships {
		fields := buildFieldList(table)
		counter := 0
		for column, rel := range fields {
			counter += 1
			if rel != nil {
				_, _ = fmt.Fprintf(f, "    optional %s.%s %s = %d; // => %s\n", rel.ForeignSchema,
					strcase.ToCamel(rel.ForeignTable), rel.MapName, counter, getLocalKeys(rel))
			} else {
				if column.DataType == "ARRAY" {
					_, _ = fmt.Fprintf(f, "    repeated ")
				} else if cfg.Proto.Version == "proto2" {
					_, _ = fmt.Fprintf(f, "    optional ")
				}
				_, _ = fmt.Fprintf(f, "%s %s = %d;\n", sqlToProto(column.UdtName), column.ColumnName, counter)
			}
		}
	} else {
		for i, column := range table.Columns {
			if column.DataType == "ARRAY" {
				_, _ = fmt.Fprintf(f, "    repeated ")
			} else if cfg.Proto.Version == "proto2" {
				_, _ = fmt.Fprintf(f, "    optional ")
			}
			_, _ = fmt.Fprintf(f, "%s %s = %d;\n", sqlToProto(column.UdtName), column.ColumnName, i + 1)
		}
	}
}

func getLocalKeys(relation *ForeignRelation) string {
	list := make([]string,0)
	for _, fcols := range relation.Columns {
		list = append(list, fcols.LocalColumn)
	}
	return strings.Join(list, ", ")
}

func buildFieldList(table Table) map[Column]*ForeignRelation {
	columns := table.Columns   // Make a copy
	for _, relation := range table.Relations {
		removeCompositeColumns(relation.Columns, &columns)
	}

	fields := make(map[Column]*ForeignRelation)
	fieldCounter := make(map[string]int64)

	for _, column := range columns {
		if rel := getRelation(column, table.Relations); rel != nil {
			counter := fieldCounter[rel.ForeignTable]
			if counter > 0 {
				rel.MapName = rel.ForeignTable + strconv.FormatInt(counter + 1, 10)
				fields[column] = rel
			} else {
				rel.MapName = rel.ForeignTable
				fields[column] = rel
			}
			fieldCounter[rel.ForeignTable] = counter + 1
		} else {
			fields[column] = nil
		}
	}

	return fields
}

func removeCompositeColumns(fcolumns []ForeignColumns, columns *[]Column){
	set := make(map[string]Column)
	firstFlag := true
	for _, col := range *columns {
		keep := true
		for _, fcol := range fcolumns {
			if fcol.LocalColumn == col.ColumnName && firstFlag {
				firstFlag = false
				break
			} else if fcol.LocalColumn == col.ColumnName {
				keep = false
			}
		}

		if keep {
			set[col.ColumnName] = col
		}
	}

	*columns = make([]Column, 0)
	for _, col := range set {
		*columns = append(*columns, col)
	}
}

func getRelation(column Column, relations []ForeignRelation) *ForeignRelation {
	for _, relation := range relations {
		for _, rcol := range relation.Columns {
			if rcol.LocalColumn == column.ColumnName {
				return &relation
			}
		}
	}
	return nil
}

func sqlToProto(sType string) string {
	if sType == "bigint" || sType == "bigint[]" || sType == "bigserial" || sType == "serial8" {
		return "int64"
	} else if strings.HasPrefix(sType, "int") || strings.HasPrefix(sType, "bit") ||
		strings.HasPrefix(sType, "smallint") || sType == "int2" || sType == "smallserial" || sType == "serial" {
		return "int32"
	} else if strings.HasPrefix(sType, "bool") {
		return "bool"
	} else if sType == "jsonb" {
		return "bytes"
	} else if strings.HasPrefix(sType, "json") {
		return "string"
	} else if strings.HasPrefix(sType, "char") || strings.HasPrefix(sType, "varchar") ||
		strings.HasPrefix(sType, "text") || sType == "xml" || sType == "uuid" {
		return "string"
	} else if sType == "money" || strings.HasPrefix(sType, "number") || sType == "numeric" ||
		strings.HasPrefix(sType, "decimal") || sType == "float8" || sType == "double precision" {
		return "double"
	} else if sType == "float" || sType == "real" {
		return "float64"
	} else if strings.HasPrefix(sType, "time") || sType == "date" {
		return "int64"
	} else if sType == "bytea" {
		return "bytes"
	} else if strings.HasPrefix(sType, "timestamp") {
		return "google.protobuf.Timestamp"
	} else {
		fmt.Printf("[warning] Failed to map postgres datatype to protobuf: %s. Using \"bytes\"\n", sType)
		return "bytes"
	}
}