package main

import (
	"bufio"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	_ "github.com/lib/pq"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func guessSeparator(file *os.File) string {
	scanner := bufio.NewScanner(file)
	separators := []string{",", ";", " ", "\t"}
	separatorCounts := make(map[string]int)
	for scanner.Scan() {
		line := scanner.Text()
		for _, sep := range separators {
			separatorCounts[sep] += strings.Count(line, sep)
		}
	}

	err := scanner.Err()
	failOnError(err, "Could not scan file")

	maxSep := separators[0]
	maxCount := separatorCounts[maxSep]
	for sep, count := range separatorCounts {
		if count > maxCount {
			maxCount = count
			maxSep = sep
		}
	}
	return maxSep
}

func main() {
	app := cli.NewApp()
	app.Name = "pgfutter"
	app.Usage = "Imports anything into PostgreSQL"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "dbname, db",
			Value:  "postgres",
			Usage:  "Database to connect to",
			EnvVar: "DB_NAME",
		},
		cli.StringFlag{
			Name:   "host",
			Value:  "localhost",
			Usage:  "Host name",
			EnvVar: "DB_HOST",
		},
		cli.StringFlag{
			Name:   "port",
			Value:  "5432",
			Usage:  "Port",
			EnvVar: "DB_PORT",
		},
		cli.StringFlag{
			Name:   "username, user",
			Value:  "postgres",
			Usage:  "Username",
			EnvVar: "DB_USER",
		},
		cli.StringFlag{
			Name:   "pass, pw",
			Value:  "",
			Usage:  "Password",
			EnvVar: "DB_PASS",
		},
	}

	app.Action = func(c *cli.Context) {
		filename := c.Args().First()
		if filename == "" {
			fmt.Println("Please provide name of file to import")
			os.Exit(1)
		}

		otherParams := "sslmode=disable connect_timeout=5"
		connStr := fmt.Sprintf("user=%s dbname=%s password='%s' host=%s port=%s %s",
			c.String("username"),
			c.String("dbname"),
			c.String("pass"),
			c.String("host"),
			c.String("port"),
			otherParams,
		)

		db, err := sql.Open("postgres", connStr)
		failOnError(err, "Could not prepare connection to database")
		defer db.Close()

		err = db.Ping()
		failOnError(err, "Could not reach the database")

		createSchema, err := db.Prepare("CREATE SCHEMA IF NOT EXISTS import")
		failOnError(err, "Could not create statement")

		_, err = createSchema.Exec()
		failOnError(err, "Could not create schema")

		file, err := os.Open(filename)
		failOnError(err, "Cannot open file")
		defer file.Close()

		reader := csv.NewReader(file)
		reader.Comma = rune(guessSeparator(file)[0])
		file.Seek(0, 0)

		columnLengths := make(map[int]int)
		for {
			record, err := reader.Read()

			if err == io.EOF {
				break
			}
			failOnError(err, "Could not read csv")

			for i, column := range record {
				if len(column) > columnLengths[i] {
					columnLengths[i] = len(column)
				}
			}
		}
		columnTypes := make(map[int](string))

		for colIndex, maxLength := range columnLengths {
			columnTypes[colIndex] = fmt.Sprintf("VARCHAR (%d)", maxLength)
		}

		for _, columnType := range columnTypes {
			fmt.Println(columnType)
		}
	}

	app.Run(os.Args)
}
