package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/cheggaaa/pb"
	"github.com/codegangsta/cli"
	"github.com/lib/pq"
)

func lineCounter(r io.Reader) (int, error) {
	buf := make([]byte, 8196)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		if err != nil && err != io.EOF {
			return count, err
		}

		count += bytes.Count(buf[:c], lineSep)

		if err == io.EOF {
			break
		}
	}

	return count, nil
}

func isValidJSON(b []byte) bool {
	var v interface{}
	err := json.Unmarshal(b, &v)
	return err == nil
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

	db, err := connect(parseConnStr(c), schema)
	failOnError(err, "Could not connect to db")
	defer db.Close()

	columns := []string{"data"}
	table, err := createJSONTable(db, schema, tableName, columns[0])
	failOnError(err, "Could not create table statement")

	_, err = table.Exec()
	failOnError(err, "Could not create table")

	txn, err := db.Begin()
	failOnError(err, "Could not start transaction")

	stmt, err := txn.Prepare(pq.CopyInSchema(schema, tableName, columns...))
	failOnError(err, "Could not prepare copy in statement")

	file, err := os.Open(filename)
	failOnError(err, "Cannot open file")
	defer file.Close()

	fi, err := file.Stat()
	failOnError(err, "Could not find out file size of file")
	total := fi.Size()
	bar := pb.New64(total)
	bar.SetUnits(pb.U_BYTES)
	bar.Start()

	reader := bufio.NewReader(io.TeeReader(file, bar))
	for {
		line, err := reader.ReadBytes('\n')

		if err == io.EOF {
			err = nil
			break
		}
		failOnError(err, "Could not read line")

		handleError := func() {
			if c.GlobalBool("ignore-errors") {
				os.Stderr.WriteString(string(line))
			} else {
				msg := fmt.Sprintf("Invalid JSON %s: %s", err, line)
				log.Fatalln(msg)
				panic(msg)
			}
		}

		if !isValidJSON(line) {
			handleError()
		}

		_, err = stmt.Exec(string(line))
		if err != nil {
			handleError()
		}

		failOnError(err, "Could add bulk insert")
	}

	_, err = stmt.Exec()
	failOnError(err, "Could not exec the bulk copy")

	err = stmt.Close()
	failOnError(err, "Could not close")

	err = txn.Commit()
	failOnError(err, "Could not commit transaction")

	bar.Finish()
}
