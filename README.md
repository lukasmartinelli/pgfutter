# pgfutter

<img align="right" alt="elephant" src="elephant.jpg" />

Import CSV and JSON into PostgreSQL the easy way.
This small tool abstract all the hassles and swearing you normally
have to deal with when you just want to dump some data into the database.

## Install

You can download a single binary for Linux, OSX or Windows.

**OSX**

```bash
wget -O pgfutter https://github.com/lukasmartinelli/pgfutter/releases/download/v0.1-alpha/pgfutter_darwin_amd64
chmod +x pgfutter
./pgfutter --help
```

**Linux**

```bash
wget -O pgfutter https://github.com/lukasmartinelli/pgfutter/releases/download/v0.1-alpha/pgfutter_linux_amd64
chmod +x pgfutter
./pgfutter --help
```

If you are using Windows or 32-Bit architectures you need to [download the appropriate binary
yourself](https://github.com/lukasmartinelli/pgfutter/releases/latest).

## Import CSV

`pgfutter` is great to take a quick look at open data sets in the database.

Let's import all traffic violations of Montgomery, Alabama.

```bash
wget -nc -O traffic_violations.csv https://data.montgomerycountymd.gov/api/views/4mse-ku6q/rows.csv
```

Because header rows are already provided `pgfutter` will create the appropriate
table and copy the rows.

```bash
pgfutter csv traffic_violations.csv
```

Database connection details can be provided via environment variables
or as separate flags.

name        | default     | description
------------|-------------|------------------------------
`DB_NAME`   | `postgres`  | host name
`DB_HOST`   | `localhost` | port
`DB_PORT`   | `5432`      | username
`DB_SCHEMA` | `import`    | schema to create tables for
`DB_USER`   | `postgres`  | database user
`DB_PASS`   |             | password (or empty if none)

### Dealing with different CSV formats

`pgfutter` will only deal with CSV files conforming to RFC 4180.
Most often you want to specify a custom delimiter (default: `,`)
or custom encoding (default: `utf-8`).

```bash
pgfutter csv -d "\t" traffic_violations.csv
```

Specifying quote characters other than `"` is not supported.

### Custom header fields

If you want to specify the field names explicitly you can
skip the header row and pass a comma separated field name list.

```bash
pgfutter csv --skip-header --fields "name,state,year" traffic_violations.csv
```

### Dealing with invalid input

A lot of CSV files don't confirm to proper CSV standards. If you want
to ignore errors you can pass the `--ignore-errors` flags which will
commit the transaction even if some rows cannot be imported.
The failed rows will be written to stdout so you can clean them up with other tools.

```bash
pgfutter --ignore-errors csv traffic_violations.csv 2> traffic_violations_errors.csv
```

### Custom Table

`pgfutter` will take the sanitized filename as the table name. If you want to specify a custom table name or import into your predefined table schema you can specify the table explicitly.

```bash
pgfutter csv --table violations traffic_violations.csv
```

## JSON

Line based JSON files are more and more common.
Each line should contain an individual JSON object.

```json
{"name": "Lukas", "age": 21, "friends": ["Alfred"]}
{"name": "Alfred", "age": 25, "friends": []}
```

Per default your JSON object will be stored in a single plain-text column.

json                                                |
----------------------------------------------------|
{"name": "Lukas", "age": 21, "friends": ["Alfred"]} |
{"name": "Alfred", "age": 25, "friends": []}        |

### Flattening objects

This example is from my [repostruct](http://github.com/lukasmartinelli/repostruct)
project, where I collect the filepaths of every GitHub repo.

Assuming we have a file containing one JSON object per line for each Github repository.

```json
{
    "repo":"sn4kebite/bin"
    "social_counts": {
        "forks":"0",
        "watchers":"1",
        "stars":"1"
    },
    "summary":{
        "releases":"0",
        "contributors":"0",
        "commits":"7",
        "branches":"1"
    },
    "language_statistics":[
        ["Python", "65.1"],
        ["Shell", "34.9"]
    ]
}
```

We can tell `pgfutter` to flatten the JSON graph directly and store
the different fields as separate columns.

```
pgfutter json repos.json --flatten
```

This will flatten all single fields into columns and store
array fields as plain text.

```sql
CREATE TABLE import.repos (
    repo TEXT,
    social_counts_forks TEXT,
    social_counts_watchers TEXT,
    social_counts_stars TEXT,
    summary_releases TEXT,
    summary_contributors TEXT,
    summary_commits TEXT,
    summary_branches TEXT,
    language_statistics TEXT
)
```

## Alternatives

For more sophisticated needs you should use [pgloader](http://pgloader.io).

## Advanced Use Cases

You can follow through some more advanced use cases to learn
how to best import data and clean it up.

## Regression Tests

The program is tested with open data sets from around the world.

Download all samples into the folder `samples`.

```bash
./download-samples.sh
```

Run import regression tests against the samples.

```bash
./test.sh
```

## Cross-compiling

We use [gox](https://github.com/mitchellh/gox) to create distributable
binaries for Windows, OSX and Linux.

```bash
docker run --rm -v "$(pwd)":/usr/src/pgfutter -w /usr/src/pgfutter tcnksm/gox:1.4.2-light
```
