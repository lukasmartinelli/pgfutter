package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

// Try to JSON decode the bytes
func tryUnmarshal(b []byte) error {
	var v interface{}
	err := json.Unmarshal(b, &v)
	return err
}

//Copy JSON Rows and return list of errors
func copyJSONRows(i *Import, reader *bufio.Reader, ignoreErrors bool) (error, int, int) {
	success := 0
	failed := 0

	for {
		// ReadBytes instead of a Scanner because it can deal with very long lines
		// which happens often with big JSON objects
		line, err := reader.ReadBytes('\n')

		if err == io.EOF {
			err = nil
			break
		}

		if err != nil {
			err = errors.New(fmt.Sprintf("%s: %s", err, line))
			return err, success, failed
		}

		err = tryUnmarshal(line)
		if err != nil {
			failed++
			if ignoreErrors {
				os.Stderr.WriteString(string(line))
				continue
			} else {
				err = errors.New(fmt.Sprintf("%s: %s", err, line))
				return err, success, failed
			}
		}

		err = i.AddRow(string(line))
		if err != nil {
			failed++
			if ignoreErrors {
				os.Stderr.WriteString(string(line))
				continue
			} else {
				err = errors.New(fmt.Sprintf("%s: %s", err, line))
				return err, success, failed
			}
		}

		success++
	}

	return nil, success, failed
}

func importJSONObject(filename string, connStr string, schema string, tableName string, dataType string) error {
	db, err := connect(connStr, schema)
	if err != nil {
		return err
	}
	defer db.Close()

	// The entire file is read into memory because we need to add
	// it into the PostgreSQL transaction, this will hit memory limits
	// for big JSON objects
	var bytes []byte
	if filename == "" {
		bytes, err = ioutil.ReadAll(os.Stdin)
	} else {
		bytes, err = ioutil.ReadFile(filename)
	}
	if err != nil {
		return err
	}

	i, err := NewJSONImport(db, schema, tableName, "data", dataType)
	if err != nil {
		return err
	}

	// The JSON file is not validated at client side
	// it is just copied into the database
	// If the JSON file is corrupt PostgreSQL will complain when querying
	err = i.AddRow(string(bytes))
	if err != nil {
		return err
	}

	return i.Commit()
}

func importJSON(filename string, connStr string, schema string, tableName string, ignoreErrors bool, dataType string) error {

	db, err := connect(connStr, schema)
	if err != nil {
		return err
	}
	defer db.Close()

	i, err := NewJSONImport(db, schema, tableName, "data", dataType)
	if err != nil {
		return err
	}

	var success, failed int
	if filename == "" {
		reader := bufio.NewReader(os.Stdin)
		err, success, failed = copyJSONRows(i, reader, ignoreErrors)
	} else {
		file, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer file.Close()

		bar := NewProgressBar(file)
		reader := bufio.NewReader(io.TeeReader(file, bar))
		bar.Start()
		err, success, failed = copyJSONRows(i, reader, ignoreErrors)
		bar.Finish()
	}

	if err != nil {
		lineNumber := success + failed
		return errors.New(fmt.Sprintf("line %d: %s", lineNumber, err))
	} else {
		fmt.Println(fmt.Sprintf("%d rows imported into %s.%s", success, schema, tableName))

		if ignoreErrors && failed > 0 {
			fmt.Println(fmt.Sprintf("%d rows could not be imported into %s.%s and have been written to stderr.", failed, schema, tableName))
		}

		return i.Commit()
	}
}
