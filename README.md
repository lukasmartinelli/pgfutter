# pgfutter

<img align="right" alt="elephant" src="elephant.jpg" />

Import CSV and JSON into PostgreSQL the easy way.
This small tool abstract all the hassles and swearing you normally
have to deal with when you just want to dump some data into the database.

## Install

You can download a single binary for Linux, OSX or Windows.

## Import CSV

We want to import a CSV file with a header row.
`pgfutter` will automatically detect the delimiter and whether it has
a header row or not. It will create a table and copy the data over to the
PostgreSQL database.

```
pgfutter csv <csv-file>
```

Given the CSV.

```
name age friends
Lukas 21 Alfred
Alfred 25
```

It will create the table

name   | age | friends
-------|-----|--------
Lukas  | 21  | Alfred"
Alfred | 21  |

### More options

You can also fully configure `pgfutter`.

```
pgfutter csv <csv-file> [--table people] [--skip-header-row] [--fields="id,name,birthday"] [--delimiter=" "] [--quote '"']
```

### JSON

The JSON import expects lines of individual JSON objects.

```json
{"name": "Lukas", "age": 21, "friends": ["Alfred"]}
{"name": "Alfred", "age": 25, "friends": []}
```

Per default your JSON object will be stored in a single plain-text column.

json                                                |
----------------------------------------------------|
{"name": "Lukas", "age": 21, "friends": ["Alfred"]} |
{"name": "Alfred", "age": 25, "friends": []}        |

## Alternatives

For more sophisticated needs you should use [pgloader](http://pgloader.io).

# Advances Usage Examples

## Flattening objects

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
