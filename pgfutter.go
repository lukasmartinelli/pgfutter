package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli"
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
		if filename == "" {
			// if no filename is not set, we reading stdin
			filename = "stdin"
		}
		base := filepath.Base(filename)
		ext := filepath.Ext(filename)
		tableName = strings.TrimSuffix(base, ext)
	}
	return postgresify(tableName)
}

func getDataType(c *cli.Context) string {
	dataType := "json"
	if c.GlobalBool("jsonb") {
		dataType = "jsonb"
	}

	return dataType
}

func main() {
	app := cli.NewApp()
	app.Name = "pgfutter"
	app.Version = "1.2"
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
			Name:  "jsonb",
			Usage: "use JSONB data type",
		},
		cli.BoolFlag{
			Name:  "ignore-errors",
			Usage: "halt transaction on inconsistencies",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:  "json",
			Usage: "Import newline-delimited JSON objects into database",
			Action: func(c *cli.Context) error {
				cli.CommandHelpTemplate = strings.Replace(cli.CommandHelpTemplate, "[arguments...]", "<json-file>", -1)

				filename := c.Args().First()

				ignoreErrors := c.GlobalBool("ignore-errors")
				schema := c.GlobalString("schema")
				tableName := parseTableName(c, filename)
				dataType := getDataType(c)

				connStr := parseConnStr(c)
				err := importJSON(filename, connStr, schema, tableName, ignoreErrors, dataType)
				return err
			},
		},
		{
			Name:  "csv",
			Usage: "Import CSV into database",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "excel",
					Usage: "support problematic Excel 2008 and Excel 2011 csv line endings",
				},
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
				cli.StringFlag{
					Name:  "null-delimiter, nd",
					Value: "\\N",
					Usage: "null delimiter",
				},
				cli.BoolFlag{
					Name:  "skip-parse-delimiter",
					Usage: "skip parsing escape sequences in the given delimiter",
				},
			},
			Action: func(c *cli.Context) error {
				cli.CommandHelpTemplate = strings.Replace(cli.CommandHelpTemplate, "[arguments...]", "<csv-file>", -1)

				filename := c.Args().First()

				ignoreErrors := c.GlobalBool("ignore-errors")
				schema := c.GlobalString("schema")
				tableName := parseTableName(c, filename)

				skipHeader := c.Bool("skip-header")
				fields := c.String("fields")
				nullDelimiter := c.String("null-delimiter")
				skipParseheader := c.Bool("skip-parse-delimiter")
				excel := c.Bool("excel")
				delimiter := parseDelimiter(c.String("delimiter"), skipParseheader)
				connStr := parseConnStr(c)
				err := importCSV(filename, connStr, schema, tableName, ignoreErrors, skipHeader, fields, delimiter, excel, nullDelimiter)
				return err
			},
		},
	}

	app.Run(os.Args)
}
