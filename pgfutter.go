package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/lib/pq"
	"github.com/lukasmartinelli/pgfutter/lib"
)

func postgresify(identifier string) string {
	str := strings.ToLower(identifier)
	str = strings.Replace(str, " ", "_", -1)
	str = strings.Replace(str, "-", "_", -1)
	str = strings.Replace(str, ",", "_", -1)
	str = strings.Replace(str, "?", "", -1)
	str = strings.Replace(str, "!", "", -1)
	return str
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func connect(connStr string, importSchema string) *sql.DB {
	db, err := sql.Open("postgres", connStr)
	failOnError(err, "Could not prepare connection to database")

	err = db.Ping()
	failOnError(err, "Could not reach the database")

	createSchema, err := db.Prepare(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", importSchema))
	failOnError(err, "Could not create schema statement")

	_, err = createSchema.Exec()
	failOnError(err, fmt.Sprintf("Could not create schema %s", importSchema))

	return db
}

func createConnStr(c *cli.Context) string {
	otherParams := "sslmode=disable connect_timeout=5"
	return fmt.Sprintf("user=%s dbname=%s password='%s' host=%s port=%s %s",
		c.GlobalString("username"),
		c.GlobalString("dbname"),
		c.GlobalString("pass"),
		c.GlobalString("host"),
		c.GlobalString("port"),
		otherParams,
	)
}

func createTableStatement(db *sql.DB, schema string, tableName string, columns []string) *sql.Stmt {
	columnTypes := make([]string, len(columns))
	for i, col := range columns {
		columnTypes[i] = fmt.Sprintf("%s TEXT", col)
	}
	columnDefinitions := strings.Join(columnTypes, ",")
	fullyQualifiedTable := fmt.Sprintf("%s.%s", schema, tableName)
	tableSchema := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", fullyQualifiedTable, columnDefinitions)

	statement, err := db.Prepare(tableSchema)
	failOnError(err, fmt.Sprintf("Could not create table schema %s", tableSchema))

	return statement
}

func parseColumns(c *cli.Context, reader *csv.Reader) []string {
	var err error
	var columns []string
	if c.Bool("skip-header") {
		columns = strings.Split(c.String("fields"), ",")
	} else {
		columns, err = reader.Read()
		failOnError(err, "Could not read header row")
	}

	for i, column := range columns {
		columns[i] = postgresify(column)
	}

	return columns
}

func parseTableName(c *cli.Context, filename string) string {
	tableName := c.GlobalString("table")
	if tableName == "" {
		base := filepath.Base(filename)
		ext := filepath.Ext(filename)
		tableName = strings.TrimSuffix(base, ext)
	}
	return postgresify(tableName)
}

func importCsv(c *cli.Context) {
	filename := c.Args().First()
	if filename == "" {
		fmt.Println("Please provide name of file to import")
		os.Exit(1)
	}

	schema := c.GlobalString("schema")
	tableName := parseTableName(c, filename)

	db := connect(createConnStr(c), schema)
	defer db.Close()

	file, err := os.Open(filename)
	failOnError(err, "Cannot open file")
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = rune(c.String("delimiter")[0])
	reader.LazyQuotes = true

	columns := parseColumns(c, reader)
	reader.FieldsPerRecord = len(columns)

	createTable := createTableStatement(db, schema, tableName, columns)
	_, err = createTable.Exec()
	failOnError(err, "Could not create table")

	txn, err := db.Begin()
	failOnError(err, "Could not start transaction")

	stmt, err := txn.Prepare(pq.CopyInSchema(schema, tableName, columns...))
	failOnError(err, "Could not prepare copy in statement")

	for {
		cols := make([]interface{}, len(columns))
		record, err := reader.Read()
		for i, col := range record {
			cols[i] = col
		}

		if err == io.EOF {
			break
		}
		failOnError(err, "Could not read csv")
		_, err = stmt.Exec(cols...)
		failOnError(err, "Could add bulk insert")
	}

	_, err = stmt.Exec()
	failOnError(err, "Could not exec the bulk copy")

	err = stmt.Close()
	failOnError(err, "Could not close")

	err = txn.Commit()
	failOnError(err, "Could not commit transaction")
}

func main() {
	app := cli.NewApp()
	app.Name = "pgfutter"
	app.Usage = "Import JSON and CSV into PostgreSQL the easy way"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "dbname, db",
			Value:  "postgres",
			Usage:  "database to connect to",
			EnvVar: "DB_NAME",
		},
		cli.StringFlag{
			Name:   "host",
			Value:  "localhost",
			Usage:  "host name",
			EnvVar: "DB_HOST",
		},
		cli.StringFlag{
			Name:   "port",
			Value:  "5432",
			Usage:  "port",
			EnvVar: "DB_PORT",
		},
		cli.StringFlag{
			Name:   "username, user",
			Value:  "postgres",
			Usage:  "username",
			EnvVar: "DB_USER",
		},
		cli.StringFlag{
			Name:   "pass, pw",
			Value:  "",
			Usage:  "password",
			EnvVar: "DB_PASS",
		},
		cli.StringFlag{
			Name:   "schema",
			Value:  "import",
			Usage:  "database schema",
			EnvVar: "DB_SCHEMA",
		},
		cli.StringFlag{
			Name:   "table",
			Usage:  "destination table",
			EnvVar: "DB_TABLE",
		},
		cli.BoolFlag{
			Name:  "abort",
			Usage: "halt transaction on inconsistencies",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:   "json",
			Usage:  "Import JSON objects into database",
			Action: lib.ImportJson,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "flatten-graph, flatten",
					Usage: "flatten fields into columns",
				},
			},
		},
		{
			Name:   "csv",
			Usage:  "Import CSV into database",
			Action: importCsv,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "skip-header",
					Usage: "skip header row",
				},
				cli.StringFlag{
					Name:  "fields",
					Usage: "comma separated field names if no header row",
				},
				cli.StringFlag{
					Name:  "delimiter, d",
					Value: ",",
					Usage: "field delimiter",
				},
			},
		},
	}

	app.Run(os.Args)
}
