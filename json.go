package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/lib/pq"
)

func importJSON(c *cli.Context) {
	cli.CommandHelpTemplate = strings.Replace(cli.CommandHelpTemplate, "[arguments...]", "<json-file>", -1)

	filename := c.Args().First()
	if filename == "" {
		cli.ShowCommandHelp(c, "json")
		os.Exit(1)
	}

	schema := c.GlobalString("schema")
	tableName := parseTableName(c, filename)

	db, err := connect(parseConnStr(c), schema)
	failOnError(err, "Could not connect to db")
	defer db.Close()

	columns := []string{"data"}
	createTable, err := createJSONTable(db, schema, tableName, columns[0])
	failOnError(err, "Could not create table statement")

	_, err = createTable.Exec()
	failOnError(err, "Could not create table")

	txn, err := db.Begin()
	failOnError(err, "Could not start transaction")

	stmt, err := txn.Prepare(pq.CopyInSchema(schema, tableName, columns...))
	failOnError(err, "Could not prepare copy in statement")

	file, err := os.Open(filename)
	failOnError(err, "Cannot open file")
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var record map[string]interface{}
		value := scanner.Text()
		err := json.Unmarshal([]byte(value), &record)

		if err != nil {
			if c.GlobalBool("ignore-errors") {
				os.Stderr.WriteString(value)
			} else {
				msg := fmt.Sprintf("Invalid JSON: %s", value)
				log.Fatalln(msg)
				panic(msg)
			}
		} else {
			row, err := json.Marshal(record)
			failOnError(err, "Can not deserialize")

			_, err = stmt.Exec(row)
			failOnError(err, "Could add bulk insert")
		}
	}
	failOnError(scanner.Err(), "Could not parse")

	_, err = stmt.Exec()
	failOnError(err, "Could not exec the bulk copy")

	err = stmt.Close()
	failOnError(err, "Could not close")

	err = txn.Commit()
	failOnError(err, "Could not commit transaction")
}
