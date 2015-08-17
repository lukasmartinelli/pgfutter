package main

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/codegangsta/cli"
)

func importJSON(c *cli.Context) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var record map[string]interface{}
		value := scanner.Text()
		err := json.Unmarshal([]byte(value), &record)
		failOnError(err, "Could not unmarshal")
	}
	failOnError(scanner.Err(), "Could not parse")
}
