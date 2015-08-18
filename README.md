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

A lot of event logs contain JSON objects nowadays (e.g. [GitHub Archive](https://www.githubarchive.org/)).
`pgfutter` expects each line to have a valid JSON object.

**Example: friends.json**

```json
{"name": "Guido", "age": 21, "friends": ["Linus"]}
{"name": "Linus", "age": 25, "friends": []}
```

Importing such files is straightforward.

```
pgfutter json friends.json
```

Per default your JSON objects will be stored in a single `JSONB` column called `data`.

data                                                  |
------------------------------------------------------|
`{"name": "Guido", "age": 21, "friends": ["Linus"]}`  |
`{"name": "Linus", "age": 25, "friends": []}`         |

[PostgreSQL has excellent JSON support](http://www.postgresql.org/docs/9.3/static/functions-json.html) which means you can then start
normalizing your data.

```sql
CREATE TABLE public.person (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200),
    age INTEGER
)

SELECT data->'name', data->'age'
INTO public.person
FROM import.friends

CREATE TABLE public.friend (
    person INTEGER
    friend INTEGER
    FOREIGN KEY(person) REFERENCES public.person(id)
    FOREIGN KEY(friend) REFERENCES public.person(id)
)

SELECT data->'name' as name, data->'age' as age
INTO public.friends
FROM import.friends

WITH friend_relationships AS (
    SELECT data->'name' as name, json_array_elements(data->'friends')
    FROM import.friends
)
SELECT
(SELECT id FROM public.person WHERE name=fr.name) as person,
(SELECT id FROM public.person WHERE name=fr.name) as friend
FROM friend_relationships as fr
INTO public.friends
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
