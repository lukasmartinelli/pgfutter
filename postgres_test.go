package main

import "testing"

type testpair struct {
	columnName    string
	sanitizedName string
}

var tests = []testpair{
	{"Starting Date & Time", "starting_date__time"},
	{"[$MYCOLUMN]", "mycolumn"},
	{"({colname?!})", "colname"},
	{"m4 * 4 / 3", "m4__4___3"},
}

func TestPostgresify(t *testing.T) {
	for _, pair := range tests {
		str := postgresify(pair.columnName)
		if str != pair.sanitizedName {
			t.Error("Invalid PostgreSQL identifier expected ", pair.sanitizedName, "got ", str)
		}
	}
}
