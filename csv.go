package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/codegangsta/cli"
)

// Parse columns from first header row or from flags
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

	file, err := os.Open(filename)
	failOnError(err, "Cannot open file")
	defer file.Close()

	db, err := connect(parseConnStr(c), schema)
	failOnError(err, "Could not connect to db")
	defer db.Close()

	success := 0
	failed := 0
	bar := NewProgressBar(file)
	bar.Start()

	reader := csv.NewReader(io.TeeReader(file, bar))
	reader.Comma = rune(c.String("delimiter")[0])
	reader.LazyQuotes = true
	reader.TrailingComma = c.Bool("trailing-comma")

	columns := parseColumns(c, reader)
	reader.FieldsPerRecord = len(columns)

	i, err := NewCSVImport(db, schema, tableName, columns)
	failOnError(err, "Could not prepare import")

	for {
		cols := make([]interface{}, len(columns))
		record, err := reader.Read()

		//Loop ensures we don't insert too many values and that
		//values are properly converted into empty interfaces
		for i, col := range record {
			cols[i] = col
		}

		if err == io.EOF {
			break
		}

		//Todo: better error handling
		failOnError(err, "Could not read csv")

		err = i.AddRow(cols...)
		if err != nil {
			failed++
			line := strings.Join(record, c.GlobalString("delimiter"))

			if c.GlobalBool("ignore-errors") {
				os.Stderr.WriteString(line)
			} else {
				msg := fmt.Sprintf("Could not import %s: %s", err, line)
				log.Fatalln(msg)
				panic(msg)
			}
		} else {
			success++
		}
	}

	err = i.Commit()
	failOnError(err, "Could not commit")
	bar.Finish()

	//refactore whole reporting stuff
	//print report
	fmt.Println(fmt.Sprintf("Successfully copied %d rows into %s"))

	if c.GlobalBool("ignore-errors") {
		fmt.Println(fmt.Sprintf("%d rows could not be imported into %s and have been written to stderr."))
	} else {
		fmt.Println(fmt.Sprintf("%d rows could not be imported into %s."))
	}

	if failed > 0 && c.GlobalBool("ignore-errors") {
		fmt.Println("You can specify the --ignore-errors flag to write errors to stderr, instead of aborting the transcation.")
	}
}
