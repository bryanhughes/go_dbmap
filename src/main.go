package main

import (
	"dbmap"
	"dbmap/mariadb"
	"dbmap/postgres"
	"fmt"
	"os"
)


func main() {
	fmt.Println("Go DB Code Mapping")
	fmt.Println("=========================================================================")

	args := os.Args[1:]
	if len(args) < 1 {
		showUsage()
	}

	var cfg dbmap.Config
	configFile := args[0]
	dbmap.ReadFile(&cfg, configFile)

	var provider dbmap.Provider
	if cfg.Database.Provider == "postgres" {
		provider = &postgres.Provider{cfg}
	} else {
		provider = &mariadb.Provider{cfg}
	}
	fmt.Println("\nReading Schemas")
	fmt.Println("=========================================================================")
	database := provider.ReadDatabase()

	fmt.Println("\nGenerating Protos")
	fmt.Println("=========================================================================")
	if err := dbmap.GenerateProto(cfg, database); err != nil {
		os.Exit(-1)
	}

	fmt.Println("\nGenerating Code")
	fmt.Println("=========================================================================")
	if err := dbmap.GenerateCode(cfg, database); err != nil {
		os.Exit(-1)
	}
}

func showUsage() {
	fmt.Println("Invalid usage!")
	fmt.Println("usage: go_dbmap <config-file>")
	os.Exit(-1)
}


