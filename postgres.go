package main

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/codegangsta/cli"
)

func connect(connStr string, importSchema string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return db, err
	}

	err = db.Ping()
	if err != nil {
		return db, err
	}

	createSchema, err := db.Prepare(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", importSchema))
	if err != nil {
		return db, err
	}

	_, err = createSchema.Exec()
	if err != nil {
		return db, err
	}

	return db, nil
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

func CreateTableStatement(db *sql.DB, schema string, tableName string, columns []string) *sql.Stmt {
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

//Parse table to copy to from given filename or passed flags
func ParseTableName(c *cli.Context, filename string) string {
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
	str = strings.Replace(str, " ", "_", -1)
	str = strings.Replace(str, "/", "_", -1)
	str = strings.Replace(str, ".", "_", -1)
	str = strings.Replace(str, ":", "_", -1)
	str = strings.Replace(str, "-", "_", -1)
	str = strings.Replace(str, ",", "_", -1)
	str = strings.Replace(str, "?", "", -1)
	str = strings.Replace(str, "!", "", -1)

	first_letter := string(str[0])
	if _, err := strconv.Atoi(first_letter); err == nil {
		str = "_" + str
	}

	return str
}
