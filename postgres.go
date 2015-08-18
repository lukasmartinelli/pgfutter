package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/codegangsta/cli"
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

//Parse table to copy to from given filename or passed flags
func parseTableName(c *cli.Context, filename string) string {
	tableName := c.GlobalString("table")
	if tableName == "" {
		base := filepath.Base(filename)
		ext := filepath.Ext(filename)
		tableName = strings.TrimSuffix(base, ext)
	}
	return postgresify(tableName)
}

//Makes sure that a string is a valid PostgreSQL identifier
func postgresify(identifier string) string {
	str := strings.ToLower(identifier)
	str = strings.TrimSpace(str)
	str = strings.Replace(str, " ", "_", -1)
	str = strings.Replace(str, "/", "_", -1)
	str = strings.Replace(str, ".", "_", -1)
	str = strings.Replace(str, ":", "_", -1)
	str = strings.Replace(str, ";", "_", -1)
	str = strings.Replace(str, "-", "_", -1)
	str = strings.Replace(str, ",", "_", -1)
	str = strings.Replace(str, "?", "", -1)
	str = strings.Replace(str, "!", "", -1)
	str = strings.Replace(str, "$", "", -1)
	str = strings.Replace(str, "%", "", -1)

	str = strings.Replace(str, "é", "e", -1)
	str = strings.Replace(str, "ê", "e", -1)
	str = strings.Replace(str, "è", "e", -1)
	str = strings.Replace(str, "ü", "u", -1)
	str = strings.Replace(str, "ä", "a", -1)
	str = strings.Replace(str, "à", "a", -1)
	str = strings.Replace(str, "ö", "o", -1)
	str = strings.Replace(str, "ô", "o", -1)

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
