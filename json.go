package main

import (
	"bufio"
	"encoding/json"
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

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var record map[string]interface{}
		value := scanner.Text()
		err := json.Unmarshal([]byte(value), &record)
		failOnError(err, "Could not unmarshal")
	}
	failOnError(scanner.Err(), "Could not parse")
}
