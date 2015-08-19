package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/cheggaaa/pb"
	"github.com/codegangsta/cli"
)

func isValidJSON(b []byte) bool {
	var v interface{}
	err := json.Unmarshal(b, &v)
	return err == nil
}

// NewProgressBar initializes new progress bar based on size of file
func NewProgressBar(file *os.File) *pb.ProgressBar {
	fi, err := file.Stat()

	total := int64(0)
	if err == nil {
		total = fi.Size()
	}

	bar := pb.New64(total)
	bar.SetUnits(pb.U_BYTES)
	bar.Start()
	return bar
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
	failOnError(err, "Cannot open file")
	defer file.Close()

	db, err := connect(parseConnStr(c), schema)
	failOnError(err, "Could not connect to db")
	defer db.Close()

	success := 0
	failed := 0
	bar := NewProgressBar(file)

	i, err := NewJSONImport(db, schema, tableName, "data")

	reader := bufio.NewReader(io.TeeReader(file, bar))
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

		//todo: not so happy with this part
		handleError := func() {
			failed++
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

		err = i.AddRow(string(line))
		if err != nil {
			handleError()
		} else {
			success++
		}

	}

	i.Commit()
	bar.Finish()
}
