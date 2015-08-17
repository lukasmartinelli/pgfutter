# pgfutter

<img align="right" alt="elephant" src="elephant.jpg" />

Import CSV and JSON into PostgreSQL the easy way.
This small tool abstract all the hassles and swearing you normally
have to deal with when you just want to dump some data into the database.

`pgfutter` will only help you to get the data into the database. After that
you should sanitize and normalize the data according to your desired
database schema.

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

If you are using Windows or 32-bit architectures you need to [download the appropriate binary
yourself](https://github.com/lukasmartinelli/pgfutter/releases/latest).

## Import CSV

Let's import all traffic violations of Montgomery, Alabama.

```bash
wget -nc -O traffic_violations.csv https://data.montgomerycountymd.gov/api/views/4mse-ku6q/rows.csv
```

Because header rows are already provided `pgfutter` will create the appropriate
table and copy the rows.

```bash
pgfutter csv traffic_violations.csv
```

### Dealing with different CSV formats

`pgfutter` will only deal with CSV files conforming to RFC 4180.
Most often you want to specify a custom delimiter (default: `,`).

```bash
pgfutter csv -d "\t" traffic_violations.csv
```

You have to use `"` as a quoting character and `\` as escape character.
You might omit the quoting character if it is not necessary.

### Custom header fields

If you want to specify the field names explicitly you can
skip the header row and pass a comma separated field name list.

```bash
pgfutter csv --skip-header --fields "name,state,year" traffic_violations.csv
```

### Encoding

All CSV files need to be `utf-8` encoded. No other encoding is supported.
Encoding is a nasty topic and you should deal with it before it enters
the database.

### Dealing with invalid input

A lot of CSV files don't confirm to proper CSV standards. If you want
to ignore errors you can pass the `--ignore-errors` flag which will
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

## Import JSON

Line based JSON files are more and more common.
Each line should contain an individual JSON object.

```json
{"name": "Lukas", "age": 21, "friends": ["Alfred"]}
{"name": "Alfred", "age": 25, "friends": []}
```

Per default your JSON objects will be stored in a single plain-text column.

json                                                  |
------------------------------------------------------|
`{"name": "Lukas", "age": 21, "friends": ["Alfred"]}` |
`{"name": "Alfred", "age": 25, "friends": []}`        |

[PostgreSQL has excellent JSON support](http://www.postgresql.org/docs/9.3/static/functions-json.html) which means you can then start
normalizing your data.

```sql
WITH imported_friends AS (
    SELECT to_json(json) as friend FROM import.friends
)
SELECT friend->'name', friend->'age'
INTO public.friends
FROM imported_friends
```

## Database Connection

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

## Alternatives

For more sophisticated needs you should use [pgloader](http://pgloader.io).

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
