package lib

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/codegangsta/cli"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func ImportJson(c *cli.Context) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var record map[string]interface{}
		value := scanner.Text()
		err := json.Unmarshal([]byte(value), &record)
		failOnError(err, "Could not unmarshal")
	}
	failOnError(scanner.Err(), "Could not parse")
}
