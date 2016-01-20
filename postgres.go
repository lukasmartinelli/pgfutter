package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/kennygrant/sanitize"
)

//tries to create the schema and ignores failures to do so.
//versions after Postgres 9.3 support the "IF NOT EXISTS" sql syntax
func tryCreateSchema(db *sql.DB, importSchema string) {
	createSchema, err := db.Prepare(fmt.Sprintf("CREATE SCHEMA %s", importSchema))

	if err == nil {
		createSchema.Exec()
	}
}

//setup a database connection and create the import schema
func connect(connStr string, importSchema string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return db, err
	}

	err = db.Ping()
	if err != nil {
		return db, err
	}

	tryCreateSchema(db, importSchema)
	return db, nil
}

//Makes sure that a string is a valid PostgreSQL identifier
func postgresify(identifier string) string {
	str := sanitize.BaseName(identifier)
	str = strings.ToLower(identifier)
	str = strings.TrimSpace(str)

	replacements := map[string]string{
		" ": "_",
		"/": "_",
		".": "_",
		":": "_",
		";": "_",
		"|": "_",
		"-": "_",
		",": "_",

		"[":  "",
		"]":  "",
		"{":  "",
		"}":  "",
		"(":  "",
		")":  "",
		"?":  "",
		"!":  "",
		"$":  "",
		"%":  "",
		"*":  "",
		"\"": "",
	}
	for oldString, newString := range replacements {
		str = strings.Replace(str, oldString, newString, -1)
	}

	if len(str) == 0 {
		str = fmt.Sprintf("_col%d", rand.Intn(10000))
	} else {
		firstLetter := string(str[0])
		if _, err := strconv.Atoi(firstLetter); err == nil {
			str = "_" + str
		}
	}

	return str
}

//parse sql connection string from cli flags
func parseConnStr(c *cli.Context) string {
	otherParams := "sslmode=disable connect_timeout=5"
	if c.GlobalBool("ssl") {
		otherParams = "sslmode=require connect_timeout=5"
	}
	return fmt.Sprintf("user=%s dbname=%s password='%s' host=%s port=%s %s",
		c.GlobalString("username"),
		c.GlobalString("dbname"),
		c.GlobalString("pass"),
		c.GlobalString("host"),
		c.GlobalString("port"),
		otherParams,
	)
}

//create table with a single JSON column data
func createJSONTable(db *sql.DB, schema string, tableName string, column string) (*sql.Stmt, error) {
	fullyQualifiedTable := fmt.Sprintf("%s.%s", schema, tableName)
	tableSchema := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s JSON)", fullyQualifiedTable, column)

	statement, err := db.Prepare(tableSchema)
	if err == nil {
		return statement, err
	}

	return statement, nil
}

//create table with TEXT columns
func createTable(db *sql.DB, schema string, tableName string, columns []string) (*sql.Stmt, error) {
	columnTypes := make([]string, len(columns))
	for i, col := range columns {
		columnTypes[i] = fmt.Sprintf("%s TEXT", col)
	}
	columnDefinitions := strings.Join(columnTypes, ",")
	fullyQualifiedTable := fmt.Sprintf("%s.%s", schema, tableName)
	tableSchema := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", fullyQualifiedTable, columnDefinitions)

	statement, err := db.Prepare(tableSchema)
	if err == nil {
		return statement, err
	}

	return statement, nil
}
