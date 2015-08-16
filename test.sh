#!/bin/bash
source tools/assert.sh

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

recreate_db
download_csv_samples

pgfutter csv "$SAMPLES_DIR/customer_complaints.csv"
assert $(wc -l "$SAMPLES_DIR/customer_complaints.csv") $(query_counts customer_complaints)

pgfutter csv "$SAMPLES_DIR/traffic_violations.csv"
assert $(wc -l "$SAMPLES_DIR/traffic_violations.csv") $(query_counts traffic_violations)
