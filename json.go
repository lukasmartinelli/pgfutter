package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/codegangsta/cli"
)

func importJSON(c *cli.Context) {
	cli.CommandHelpTemplate = strings.Replace(cli.CommandHelpTemplate, "[arguments...]", "<json-file>", -1)

	filename := c.Args().First()
	if filename == "" {
		cli.ShowCommandHelp(c, "json")
		os.Exit(1)
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
		}
	}
	failOnError(scanner.Err(), "Could not parse")
}
