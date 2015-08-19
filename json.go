package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/codegangsta/cli"
)

// Try to JSON decode the bytes
func isValidJSON(b []byte) bool {
	var v interface{}
	err := json.Unmarshal(b, &v)
	return err == nil
}

func copyJSONRows(i *Import, reader *bufio.Reader, ignoreErrors bool) (error, int, int) {
	success := 0
	failed := 0

	for {
		// We use ReadBytes because it can deal with very long lines
		// which happens often with big JSON objects
		line, err := reader.ReadBytes('\n')

		if err == io.EOF {
			err = nil
			break
		}

		//todo: Better error handling so that db can close
		failOnError(err, "Could not read line")

		valid := isValidJSON(line)
		if valid {
			err = i.AddRow(string(line))
		}

		if !valid || err != nil {
			failed++
			if ignoreErrors {
				os.Stderr.WriteString(string(line))
			} else {
				return err, success, failed
			}
		} else {
			success++
		}
	}

	return nil, success, failed
}

func importJSON(c *cli.Context) {
	cli.CommandHelpTemplate = strings.Replace(cli.CommandHelpTemplate, "[arguments...]", "<json-file>", -1)

	filename := c.Args().First()
	if filename == "" {
		cli.ShowCommandHelp(c, "json")
		os.Exit(1)
	}

	schema := c.GlobalString("schema")
	tableName := parseTableName(c, filename)

	file, err := os.Open(filename)
	exitOnError(err, fmt.Sprintf("Cannot open %s", filename))
	defer file.Close()

	connStr := parseConnStr(c)
	db, err := connect(connStr, schema)
	exitOnError(err, fmt.Sprintf("Cannot connect to database %s", connStr))
	defer db.Close()

	bar := NewProgressBar(file)
	i, err := NewJSONImport(db, schema, tableName, "data")
	//handle error
	reader := bufio.NewReader(io.TeeReader(file, bar))
	err, _, _ = copyJSONRows(i, reader, c.GlobalBool("ignore-errors"))

	bar.Finish()
	if err != nil {

	} else {

		// handle error
		err = i.Commit()
		failOnError(err, "Could not commit")
	}
}
