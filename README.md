# pgfutter [![Build Status](https://travis-ci.org/lukasmartinelli/pgfutter.svg?branch=master)](https://travis-ci.org/lukasmartinelli/pgfutter) [![Go Report Card](https://goreportcard.com/badge/github.com/lukasmartinelli/pgfutter)](https://goreportcard.com/report/github.com/lukasmartinelli/pgfutter) ![License](https://img.shields.io/badge/license-MIT%20License-blue.svg)

<img align="right" alt="elephant" src="elephant.jpg" />

Import CSV and line delimited JSON into PostgreSQL the easy way.
This small tool abstract all the hassles and swearing you normally
have to deal with when you just want to dump some data into the database.

Features:

- Generated import tables (`pgfutter csv <file>` and you're done)
- Good performance using the `COPY` streaming protocol
- Easy deployment
- Dealing with import errors
- Import over the network
- Only supports UTF8 encoding

> Check out [pgclimb](https://github.com/lukasmartinelli/pgclimb) for exporting data from PostgreSQL into different data formats.

## Install

You can download a single binary for Linux, OSX or Windows.

**OSX**

```bash
wget -O pgfutter https://github.com/lukasmartinelli/pgfutter/releases/download/v1.2/pgfutter_darwin_amd64
chmod +x pgfutter

./pgfutter --help
```

**Linux**

```bash
wget -O pgfutter https://github.com/lukasmartinelli/pgfutter/releases/download/v1.2/pgfutter_linux_amd64
chmod +x pgfutter

./pgfutter --help
```

**Install from source**

```bash
go get github.com/lukasmartinelli/pgfutter
```

If you are using Windows or 32-bit architectures you need to [download the appropriate binary
yourself](https://github.com/lukasmartinelli/pgfutter/releases/latest).

## Import CSV

`pgfutter` will deal with CSV files conforming to [RFC 4180](https://tools.ietf.org/html/rfc4180#section-2).

Create `friends.csv`.

```csv
name,age,friends
Jacob,26,"Anthony"
Anthony,25,""
Emma,28,"Jacob,Anthony"
```

Import the CSV file.

```bash
pgfutter csv friends.csv
```

Because header rows are already provided `pgfutter` will create the appropriate
table and copy the rows.

name    | age| friends         |
--------|----|-----------------|
Jacob   | 26 | Anthony         |
Anthony | 25 |                 |
Emma    | 28 | Jacob,Anthony   |


`pgfutter` will only help you to get the data into the database. After that
SQL is a great language to sanitize and normalize the data according to your desired database schema.

```sql
CREATE TABLE public.person (
    name VARCHAR(200) PRIMARY KEY,
    age INTEGER
)

CREATE TABLE public.friendship (
    person VARCHAR(200) REFERENCES public.person(name),
    friend VARCHAR(200) REFERENCES public.person(name)
)

INSERT INTO public.person
SELECT name, age::int
FROM import.friends

WITH friends AS
    (SELECT name as person, regexp_split_to_table(friends, E'\\,') AS friend
    FROM import.friends)
INSERT INTO public.friendship
SELECT * FROM
friends WHERE friend <> ''
```

## Import JSON

A lot of event logs contain JSON objects nowadays (e.g. [GitHub Archive](https://www.githubarchive.org/)).
`pgfutter` expects each line to have a valid JSON object. Importing JSON is only supported for Postgres 9.3 and Postgres 9.4 due to the `JSON` type.

Create `friends.json`.

```json
{"name": "Jacob", "age": 26, "friends": ["Anthony"]}
{"name": "Anthony", "age": 25, "friends": []}
{"name": "Emma", "age": 28, "friends": ["Jacob", "Anthony"]}

```

Import the JSON file.

```bash
pgfutter json friends.json
```

Your JSON objects will be stored in a single [JSON](http://www.postgresql.org/docs/9.4/static/datatype-json.html) column called `data`.

data                                                          |
--------------------------------------------------------------|
`{"name": "Jacob", "age": 26, "friends": ["Anthony"]}`        |
`{"name": "Anthony", "age": 25, "friends": []}`               |
`{"name": "Emma", "age": 28, "friends": ["Jacob", "Anthony"]}`|

[PostgreSQL has excellent JSON support](http://www.postgresql.org/docs/9.3/static/functions-json.html) which means you can then start
normalizing your data.

```sql
CREATE TABLE public.person (
    name VARCHAR(200) PRIMARY KEY,
    age INTEGER
)

CREATE TABLE public.friendship (
    person VARCHAR(200) REFERENCES public.person(name),
    friend VARCHAR(200) REFERENCES public.person(name)
)

INSERT INTO public.person
SELECT data->>'name' as name, (data->>'age')::int as age
FROM import.friends

INSERT INTO public.friendship
SELECT data->>'name' as person, json_array_elements_text(data->'friends')
FROM import.friends
```

## Database Connection

Database connection details can be provided via environment variables
or as separate flags.

name        | default     | description
------------|-------------|------------------------------
`DB_NAME`   | `postgres`  | database name
`DB_HOST`   | `localhost` | host name
`DB_PORT`   | `5432`      | port
`DB_SCHEMA` | `import`    | schema to create tables for
`DB_USER`   | `postgres`  | database user
`DB_PASS`   |             | password (or empty if none)

## Advanced Use Cases

### Custom delimiter

Quite often you want to specify a custom delimiter (default: `,`).

```bash
pgfutter csv -d "\t" traffic_violations.csv
```

You have to use `"` as a quoting character and `\` as escape character.
You might omit the quoting character if it is not necessary.


### Using TAB as delimiter

If you want to use tab as delimiter you need to pass `$'\t'` as delimiter
to ensure your shell does not swallow the correct delimiter.

```bash
pgfutter csv -d $'\t' traffic_violations.csv
```

### Custom header fields

If you want to specify the field names explicitly you can
skip the header row and pass a comma separated field name list.

```bash
pgfutter csv --skip-header --fields "name,state,year" traffic_violations.csv
```

If you don't have a header row in a document you should specify the field names as well.

```bash
pgfutter csv --fields "name,state,year" traffic_violations.csv
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

This works the same for invalid JSON objects.

### Custom Table

`pgfutter` will take the sanitized filename as the table name. If you want to specify a custom table name or import into your predefined table schema you can specify the table explicitly.

```bash
pgfutter --table violations csv traffic_violations.csv
```

## Alternatives

For more sophisticated needs you should take a look at [pgloader](http://pgloader.io).

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
docker run --rm -v "$(pwd)":/usr/src/pgfutter -w /usr/src/pgfutter tcnksm/gox:1.9

```
