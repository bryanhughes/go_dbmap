package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	_ "gopkg.in/yaml.v3"
	"os"
	"strings"
)

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
	} `yaml:"output"`
	Proto struct {
		Path        string `yaml:"path"`
		JavaPackage string `yaml:"java_package"`
		Version     string `yaml:"proto2"`
	} `yaml:"proto"`
	Generator struct {
		Schemas        []string `yaml:"schemas"`
		ExcludedTables []string `yaml:"excluded_tables"`
		Lookup         []struct {
			Tablename string     `yaml:"table"`
			Columns   [][]string `yaml:"columns"`
		} `yaml:"lookup"`
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

func main() {
	fmt.Println("Go DB Mapping Code Generator")
	fmt.Println("============================")

	args := os.Args[1:]
	if len(args) < 1 {
		showUsage()
	}

	var cfg Config
	readFile(&cfg, args)

	readSchemas(cfg)
}

func readSchemas(config Config) {
	if strings.ToLower(config.Database.Provider) == "postgres" {
		schemas := config.Generator.Schemas
		for _, schema := range schemas {
			readSchema(config, schema)
		}
	} else {
		fmt.Println("Unsupported provider: ", config.Database.Provider)
		fmt.Println("Try back later...")
		os.Exit(0)
	}
}

func readSchema(config Config, schema string) {
	fmt.Println("Reading schema: ", schema)
}

func showUsage() {
	fmt.Println("Invalid usage!")
	fmt.Println("usage: go_dbmap <config-file>")
	os.Exit(-1)
}

func processError(err error) {
	fmt.Println(err)
	os.Exit(2)
}

func readFile(cfg *Config, args []string) {
	configFile := args[0]
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
