package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/codegangsta/cli"
)

func exitOnError(err error) {
	log.SetFlags(0)
	if err != nil {
		log.Fatalln(err)
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
	app.Version = "0.3.2"
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
		cli.BoolFlag{
			Name:  "ssl",
			Usage: "require ssl mode",
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
			Name:  "json",
			Usage: "Import JSON objects into database",
			Action: func(c *cli.Context) {
				cli.CommandHelpTemplate = strings.Replace(cli.CommandHelpTemplate, "[arguments...]", "<json-file>", -1)

				filename := c.Args().First()
				if filename == "" {
					cli.ShowCommandHelp(c, "json")
					os.Exit(1)
				}

				ignoreErrors := c.GlobalBool("ignore-errors")
				schema := c.GlobalString("schema")
				tableName := parseTableName(c, filename)

				connStr := parseConnStr(c)
				err := importJSON(filename, connStr, schema, tableName, ignoreErrors)
				exitOnError(err)
			},
		},
		{
			Name:  "csv",
			Usage: "Import CSV into database",
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
			},
			Action: func(c *cli.Context) {
				cli.CommandHelpTemplate = strings.Replace(cli.CommandHelpTemplate, "[arguments...]", "<csv-file>", -1)

				filename := c.Args().First()
				if filename == "" {
					cli.ShowCommandHelp(c, "csv")
					os.Exit(1)
				}

				ignoreErrors := c.GlobalBool("ignore-errors")
				schema := c.GlobalString("schema")
				tableName := parseTableName(c, filename)

				skipHeader := c.Bool("skip-header")
				fields := c.String("fields")
				delimiter := c.String("delimiter")

				connStr := parseConnStr(c)
				err := importCSV(filename, connStr, schema, tableName, ignoreErrors, skipHeader, fields, delimiter)
				exitOnError(err)
			},
		},
	}

	app.Run(os.Args)
}
