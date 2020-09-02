package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
)

//Parse table to copy to from given filename or passed flags
func parseTableName(c *cli.Context, filename string) string {
	tableName := c.String("table")
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
	if c.Bool("jsonb") {
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
		&cli.StringFlag{
			Name:   "dbname",
			Aliases:[]string{"db"},
			Value:  "postgres",
			Usage:  "database to connect to",
			EnvVars: []string{"DB_NAME"},
		},
		&cli.StringFlag{
			Name:   "host",
			Value:  "localhost",
			Usage:  "host name",
			EnvVars: []string{"DB_HOST"},
		},
		&cli.StringFlag{
			Name:   "port",
			Value:  "5432",
			Usage:  "port",
			EnvVars: []string{"DB_PORT"},
		},
		&cli.StringFlag{
			Name:   "username",
			Aliases:[]string{"user"},
			Value:  "postgres",
			Usage:  "username",
			EnvVars: []string{"DB_USER"},
		},
		&cli.BoolFlag{
			Name:  "ssl",
			Usage: "require ssl mode",
		},
		&cli.StringFlag{
			Name:   "pass",
			Aliases:[]string{"pw"},
			Value:  "",
			Usage:  "password",
			EnvVars: []string{"DB_PASS"},
		},
		&cli.StringFlag{
			Name:   "schema",
			Value:  "import",
			Usage:  "database schema",
			EnvVars: []string{"DB_SCHEMA"},
		},
		&cli.StringFlag{
			Name:   "table",
			Usage:  "destination table",
			EnvVars: []string{"DB_TABLE"},
		},
		&cli.BoolFlag{
			Name:  "jsonb",
			Usage: "use JSONB data type",
		},
		&cli.BoolFlag{
			Name:  "ignore-errors",
			Usage: "halt transaction on inconsistencies",
		},
	}

	app.Commands = []*cli.Command{
		{
			Name:  "json",
			Usage: "Import newline-delimited JSON objects into database",
			Action: func(c *cli.Context) error {
				cli.CommandHelpTemplate = strings.Replace(cli.CommandHelpTemplate, "[arguments...]", "<json-file>", -1)

				filename := c.Args().First()

				ignoreErrors := c.Bool("ignore-errors")
				schema := c.String("schema")
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
				&cli.BoolFlag{
					Name:  "excel",
					Usage: "support problematic Excel 2008 and Excel 2011 csv line endings",
				},
				&cli.BoolFlag{
					Name:  "skip-header",
					Usage: "skip header row",
				},
				&cli.StringFlag{
					Name:  "fields",
					Usage: "comma separated field names if no header row",
				},
				&cli.StringFlag{
					Name:       "delimiter",
					Aliases:    []string{"d"},
					Value:      ",",
					Usage:      "field delimiter",
				},
				&cli.StringFlag{
					Name:       "null-delimiter",
					Aliases:    []string{"nd"},
					Value:      "\\N",
					Usage:      "null delimiter",
				},
				&cli.BoolFlag{
					Name:  "skip-parse-delimiter",
					Usage: "skip parsing escape sequences in the given delimiter",
				},
			},
			Action: func(c *cli.Context) error {
				cli.CommandHelpTemplate = strings.Replace(cli.CommandHelpTemplate, "[arguments...]", "<csv-file>", -1)

				filename := c.Args().First()

				ignoreErrors := c.Bool("ignore-errors")
				schema := c.String("schema")
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

	err := app.Run(os.Args)
    if err != nil {
        log.SetFlags(0)
        log.Fatal(err)
    }
}
