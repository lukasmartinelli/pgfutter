# pgfutter

<img align="right" alt="elephant" src="elephant.png" />

Import CSV and JSON into PostgreSQL the easy way.
This small tool abstract all the hassles and swearing you normally
have to deal with when you just want to dump some data into the database.

Data transformations and concrete type mappings should all happen in the
database. `pgfutter` only helps you to get the data into the database.


## Install

You can download a single binary for Linux, OSX or Windows.

## Import CSV

We want to import a csv file with a header row.
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

## Import JSON

```
pgfutter json events.json --flatten
```

The JSON import expects lines of individual JSON records.

```json
{"name": "Lukas", "age": 21, "friends": ["Alfred"]}
{"name": "Alfred", "age": 25, "friends": []}
```

You can choose between storing the entire JSON object in the database or whether you want to flatten it.

The flattened JSON from above would look like this.
All fields that can not be flattened are simply stored as plain unparsed JSON `TEXT`.

name   | age | friends
-------|-----|--------
Lukas  | 21  | ["Alfred"]
Alfred | 21  | []

When importing JSON you will get a single row with the new `JSONB` datatype.
If you specify the `--expand` flag, all flat properties are automatically
expanded into columns.

Without the `--flatten` flag, your JSON object will be stored in a single plain-text column.

json                                                |
----------------------------------------------------|
{"name": "Lukas", "age": 21, "friends": ["Alfred"]} |
{"name": "Alfred", "age": 25, "friends": []}        |

## Alternatives

For more sophisticated needs you should use [pgloader](http://pgloader.io).
