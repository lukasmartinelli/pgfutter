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

You can also fully configure `pgfutter`.

```
pgfutter csv <csv-file> [--table people] [--skip-header-row] [--fields="id,name,birthday"] [--delimiter=" "] [--quote '"']
```

## Import JSON

When importing JSON you will get a single row with the new `JSONB` datatype.
If you specify the `--expand` flag, all flat properties are automatically
expanded into columns.

```
pgfutter json events.json
```

This can probably be compared with [pgloader](http://pgloader.io).
>>>>>>> Improved readme
