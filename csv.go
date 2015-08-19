package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/cheggaaa/pb"
	"github.com/codegangsta/cli"
	"github.com/lib/pq"
)

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

func importCsv(c *cli.Context) {
	cli.CommandHelpTemplate = strings.Replace(cli.CommandHelpTemplate, "[arguments...]", "<csv-file>", -1)

	filename := c.Args().First()
	if filename == "" {
		cli.ShowCommandHelp(c, "csv")
		os.Exit(1)
	}

	schema := c.GlobalString("schema")
	tableName := parseTableName(c, filename)

	db, err := connect(parseConnStr(c), schema)
	failOnError(err, "Could not connect to db")
	defer db.Close()

	file, err := os.Open(filename)
	failOnError(err, "Cannot open file")
	defer file.Close()

	fi, err := file.Stat()
	failOnError(err, "Could not find out file size of file")
	total := fi.Size()
	bar := pb.New64(total)
	bar.SetUnits(pb.U_BYTES)
	bar.Start()

	reader := csv.NewReader(io.TeeReader(file, bar))
	reader.Comma = rune(c.String("delimiter")[0])
	reader.LazyQuotes = true
	reader.TrailingComma = c.Bool("trailing-comma")

	columns := parseColumns(c, reader)
	reader.FieldsPerRecord = len(columns)

	table, err := createTable(db, schema, tableName, columns)
	failOnError(err, "Could not create table statement")

	_, err = table.Exec()
	failOnError(err, "Could not create table")

	txn, err := db.Begin()
	failOnError(err, "Could not start transaction")

	stmt, err := txn.Prepare(pq.CopyInSchema(schema, tableName, columns...))
	failOnError(err, "Could not prepare copy in statement")

	successCount := 0
	failCount := 0

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
		successCount++
	}

	_, err = stmt.Exec()
	failOnError(err, "Could not exec the bulk copy")

	err = stmt.Close()
	failOnError(err, "Could not close")

	err = txn.Commit()
	failOnError(err, "Could not commit transaction")

	//print report
	fmt.Println(fmt.Sprintf("Successfully copied %d rows into %s"))

	if c.GlobalBool("ignore-errors") {
		fmt.Println(fmt.Sprintf("%d rows could not be imported into %s and have been written to stderr."))
	} else {
		fmt.Println(fmt.Sprintf("%d rows could not be imported into %s."))
	}

	if failCount > 0 && c.GlobalBool("ignore-errors") {
		fmt.Println("You can specify the --ignore-errors flag to write errors to stderr, instead of aborting the transcation.")
	}

	bar.Finish()
}
