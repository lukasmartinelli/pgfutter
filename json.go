package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/codegangsta/cli"
)

func importJSONFromArrayField(c *cli.Context) {
	file, err := ioutil.ReadFile("./config.json")
	failOnError(err, "Cannot open file")

	var record map[string]interface{}
	err = json.Unmarshal(file, &record)
	failOnError(err, "Invalid JSON file")

	arrayField := c.GlobalString("from-array-field")
	array := record[arrayField]
	fmt.Printf("%+v", array)
}

func importJSON(c *cli.Context) {
	cli.CommandHelpTemplate = strings.Replace(cli.CommandHelpTemplate, "[arguments...]", "<json-file>", -1)

	filename := c.Args().First()
	if filename == "" {
		cli.ShowCommandHelp(c, "json")
		os.Exit(1)
	}

	if c.GlobalString("from-array-field") != "" {
		importJSONFromArrayField(c)
		return
	}

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
			fmt.Println(row)
		}
	}
	failOnError(scanner.Err(), "Could not parse")
}
