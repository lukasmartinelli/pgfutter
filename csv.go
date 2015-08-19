package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"unicode/utf8"
)

// Parse columns from first header row or from flags
func parseColumns(reader *csv.Reader, skipHeader bool, fields string) []string {
	var err error
	var columns []string
	if skipHeader {
		columns = strings.Split(fields, ",")
	} else {
		columns, err = reader.Read()
		failOnError(err, "Could not read header row")
	}

	for i, column := range columns {
		columns[i] = postgresify(column)
	}

	return columns
}

func importCSV(filename string, connStr string, schema string, tableName string, ignoreErrors bool, skipHeader bool, fields string, delimiter string) error {

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	db, err := connect(connStr, schema)
	if err != nil {
		return err
	}
	defer db.Close()

	success := 0
	failed := 0
	bar := NewProgressBar(file)
	bar.Start()

	reader := csv.NewReader(io.TeeReader(file, bar))
	reader.Comma, _ = utf8.DecodeRuneInString(delimiter)
	reader.LazyQuotes = true

	columns := parseColumns(reader, skipHeader, fields)
	reader.FieldsPerRecord = len(columns)

	i, err := NewCSVImport(db, schema, tableName, columns)
	if err != nil {
		return err
	}

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

		if err != nil {
			failed++
			line := strings.Join(record, delimiter)

			if ignoreErrors {
				os.Stderr.WriteString(line)
			} else {
				msg := fmt.Sprintf("Could not parse %s: %s", err, line)
				break
			}
		} else {
			success++
		}

		err = i.AddRow(cols...)

		if err != nil {
			failed++
			line := strings.Join(record, delimiter)

			if ignoreErrors {
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

	bar.Finish()

	if err != nil {
		i.Rollback()
		lineNumber := success + failed
		if !skipHeader {
			lineNumber++
		}
		return errors.New(fmt.Sprintf("line %d: %s", lineNumber, err))
	} else {
		fmt.Println(fmt.Sprintf("%d rows have successfully been copied into %s.%s", success, schema, tableName))

		if ignoreErrors && failed > 0 {
			fmt.Println(fmt.Sprintf("%d rows could not be imported into %s.%s and have been written to stderr.", failed, schema, tableName))
		}

		return i.Commit()
	}
}
