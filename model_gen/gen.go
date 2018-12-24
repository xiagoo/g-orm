package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os/exec"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mijia/modelq/drivers"
	"github.com/xiagoo/g-orm/generator"
)

func main() {
	var db, tableNames, packageName string
	var driver, schemaName string
	flag.StringVar(&db, "db", "", "Database source string: root@tcp(127.0.0.1:3306)/test?parseTime=true&loc=Local")
	flag.StringVar(&tableNames, "tables", "", "Create tables, e.g. \"user,books\"")
	flag.StringVar(&packageName, "pkg", "", "GO package name")
	flag.StringVar(&driver, "driver", "mysql", "Current supported drivers include mysql, postgres")
	flag.StringVar(&schemaName, "schema", "", "Schema for postgresql, database name for mysql")
	flag.Parse()

	if db == "" {
		fmt.Println("Please provide the target database source.")
		fmt.Println("Usage:")
		flag.PrintDefaults()
		return
	}
	if packageName == "" {
		printUsages("Please provide the go source code package name for generated models.")
		return
	}
	if driver != "mysql" && driver != "postgres" {
		printUsages("Current supported drivers include mysql, postgres.")
		return
	}
	if schemaName == "" {
		printUsages("Please provide the schema name.")
		return
	}

	dbSchema, err := drivers.LoadDatabaseSchema(driver, db, schemaName, tableNames)
	if err != nil {
		panic(err)
	}
	codeConfig := generator.CodeConfig{
		PackageName: packageName,
	}
	generator.GenerateModels(schemaName, dbSchema, codeConfig)
	formatCodes(packageName)
}

func formatCodes(pkg string) {
	log.Println("Running gofmt *.go")
	var out bytes.Buffer
	cmd := exec.Command("gofmt", "-w", pkg)
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		log.Println(out.String())
		log.Fatalf("Fail to run gofmt package, %s", err)
	}
}

func printUsages(message ...interface{}) {
	for _, x := range message {
		fmt.Println(x)
	}
	fmt.Println("\nUsage:")
	flag.PrintDefaults()
}
