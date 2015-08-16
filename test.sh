#!/bin/bash
CWD=$(pwd)
SAMPLES_DIR="$CWD/samples"
DB_USER="postgres"
DB_NAME="integration_test"
DB_SCHEMA="import"

function recreate_db() {
  psql -U ${DB_USER} -c "drop database if exists ${DB_NAME};"
  psql -U ${DB_USER} -c "create database ${DB_NAME};"
}

function download_csv_samples() {
    mkdir -p $SAMPLES_DIR
    cd $SAMPLES_DIR
    wget -nc -O customer_complaints.csv https://data.consumerfinance.gov/api/views/x94z-ydhh/rows.csv
    wget -nc -O traffic_violations.csv https://data.montgomerycountymd.gov/api/views/4mse-ku6q/rows.csv
    cd $CWD
}

function query_counts() {
    local table=$1
    local counts=$(psql -U ${DB_USER} -d ${DB_NAME} -t -c "select count(*) from ${DB_SCHEMA}.${table}")
    echo "$counts"
}

function import_and_test_csv() {
    local table=$1
    local filename=$2

    pgfutter csv "$filename"
    if [ $? -ne 0 ]; then
        echo "pgfutter could not import $filename"
        exit 300
    fi

    local line_count=$(cat "$filename" | wc -l)
    local expected_db_count=$(expr $line_count - 1)
    local db_count=$(query_counts $table)

    if [ "$db_count" -ne "$expected_db_count" ]; then
        echo "Import for $table failed. Only $db_count rows of $line_count rows imported"
        exit 201
    fi
}

recreate_db
download_csv_samples

import_and_test_csv "customer_complaints" "$SAMPLES_DIR/customer_complaints.csv"
import_and_test_csv "traffic_violations" "$SAMPLES_DIR/traffic_violations.csv"
