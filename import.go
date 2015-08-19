package main

import (
	"database/sql"

	"github.com/lib/pq"
)

type Import struct {
	txn  *sql.Tx
	stmt *sql.Stmt
}

func NewCSVImport(db *sql.DB, schema string, tableName string, columns []string) (*Import, error) {

	table, err := createTable(db, schema, tableName, columns)
	if err != nil {
		return nil, err
	}

	_, err = table.Exec()
	if err != nil {
		return nil, err
	}

	return newImport(db, schema, tableName, columns)
}

func NewJSONImport(db *sql.DB, schema string, tableName string, column string) (*Import, error) {

	table, err := createJSONTable(db, schema, tableName, column)
	if err != nil {
		return nil, err
	}

	_, err = table.Exec()
	if err != nil {
		return nil, err
	}

	return newImport(db, schema, tableName, []string{column})
}

func newImport(db *sql.DB, schema string, tableName string, columns []string) (*Import, error) {

	txn, err := db.Begin()
	if err != nil {
		return nil, err
	}

	stmt, err := txn.Prepare(pq.CopyInSchema(schema, tableName, columns...))
	if err != nil {
		return nil, err
	}

	return &Import{txn, stmt}, nil
}

func (i *Import) AddRow(columns ...interface{}) error {
	_, err := i.stmt.Exec(columns...)
	return err
}

func (i *Import) Commit() error {

	_, err := i.stmt.Exec()
	if err != nil {
		return err
	}

	err = i.stmt.Close()
	if err != nil {
		return err
	}

	err = i.txn.Commit()
	if err != nil {
		return err
	}

	return nil
}
