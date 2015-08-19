package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/codegangsta/cli"
)

func exitOnError(err error, msg string) {
	log.SetFlags(0)
	if err != nil {
		log.Fatalln(msg)
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

//Parse table to copy to from given filename or passed flags
func parseTableName(c *cli.Context, filename string) string {
	tableName := c.GlobalString("table")
	if tableName == "" {
		base := filepath.Base(filename)
		ext := filepath.Ext(filename)
		tableName = strings.TrimSuffix(base, ext)
	}
	return postgresify(tableName)
}

func main() {
	app := cli.NewApp()
	app.Name = "pgfutter"
	app.Usage = "Import JSON and CSV into PostgreSQL the easy way"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "dbname, db",
			Value:  "postgres",
			Usage:  "database to connect to",
			EnvVar: "DB_NAME",
		},
		cli.StringFlag{
			Name:   "host",
			Value:  "localhost",
			Usage:  "host name",
			EnvVar: "DB_HOST",
		},
		cli.StringFlag{
			Name:   "port",
			Value:  "5432",
			Usage:  "port",
			EnvVar: "DB_PORT",
		},
		cli.StringFlag{
			Name:   "username, user",
			Value:  "postgres",
			Usage:  "username",
			EnvVar: "DB_USER",
		},
		cli.StringFlag{
			Name:   "pass, pw",
			Value:  "",
			Usage:  "password",
			EnvVar: "DB_PASS",
		},
		cli.StringFlag{
			Name:   "schema",
			Value:  "import",
			Usage:  "database schema",
			EnvVar: "DB_SCHEMA",
		},
		cli.StringFlag{
			Name:   "table",
			Usage:  "destination table",
			EnvVar: "DB_TABLE",
		},
		cli.BoolFlag{
			Name:  "ignore-errors",
			Usage: "halt transaction on inconsistencies",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:   "json",
			Usage:  "Import JSON objects into database",
			Action: importJSON,
		},
		{
			Name:   "csv",
			Usage:  "Import CSV into database",
			Action: importCsv,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "skip-header",
					Usage: "skip header row",
				},
				cli.StringFlag{
					Name:  "fields",
					Usage: "comma separated field names if no header row",
				},
				cli.StringFlag{
					Name:  "delimiter, d",
					Value: ",",
					Usage: "field delimiter",
				},
				cli.BoolFlag{
					Name:  "trailing-comma",
					Usage: "extra delimiter at end of line",
				},
			},
		},
	}

	app.Run(os.Args)
}
