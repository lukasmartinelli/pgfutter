#!/bin/bash
CWD=$(pwd)
SAMPLES_DIR="$CWD/samples"
DB_USER="postgres"
DB_NAME="integration_test"
DB_SCHEMA="import" # Use public schema instead of import because of permissions

function recreate_db() {
  psql -U ${DB_USER} -c "drop database if exists ${DB_NAME};"
  psql -U ${DB_USER} -c "create database ${DB_NAME};"
}

function query_counts() {
    local table=$1
    local counts=$(psql -U ${DB_USER} -d ${DB_NAME} -t -c "select count(*) from ${DB_SCHEMA}.${table}")
    echo "$counts"
}

function test_readme_csv_sample() {
    # test whether readme docs still work
    echo "test"
}

function import_csv_with_special_delimiter_and_trailing() {
    local table="qip12_tabdaten"
    local filename="$SAMPLES_DIR/qip12_tabdaten.csv"
    pgfutter --schema $DB_SCHEMA --db $DB_NAME --user $DB_USER csv "$filename" --delimiter=";" --trailing-comma
    if [ $? -ne 0 ]; then
        echo "pgfutter could not import $filename"
        exit 300
    else
        local db_count=$(query_counts $table)
        echo "Imported $(expr $db_count) records into $table"
    fi
}

function import_and_test_json() {
    local table=$1
    local filename=$2
    pgfutter --schema $DB_SCHEMA --db $DB_NAME --user $DB_USER json "$filename"
    if [ $? -ne 0 ]; then
        echo "pgfutter could not import $filename"
        exit 300
    else
        local db_count=$(query_counts $table)
        echo "Imported $(expr $db_count) records into $table"
    fi
}

function import_and_test_csv() {
    local table=$1
    local filename=$2

    pgfutter --schema $DB_SCHEMA --db $DB_NAME --user $DB_USER csv "$filename"
    if [ $? -ne 0 ]; then
        echo "pgfutter could not import $filename"
        exit 300
    else
        local db_count=$(query_counts $table)
        echo "Imported $(expr $db_count) records into $table"
    fi
}

recreate_db

import_csv_with_special_delimiter_and_trailing
import_and_test_json "_2015_01_01_15" "$SAMPLES_DIR/2015-01-01-15.json"
import_and_test_csv "parking_garage_availability" "$SAMPLES_DIR/parking_garage_availability.csv"
import_and_test_csv "local_severe_wheather_warning_systems" "$SAMPLES_DIR/local_severe_wheather_warning_systems.csv"
import_and_test_csv "montgomery_crime" "$SAMPLES_DIR/montgomery_crime.csv"
import_and_test_csv "employee_salaries" "$SAMPLES_DIR/employee_salaries.csv"
import_and_test_csv "residential_permits" "$SAMPLES_DIR/residential_permits.csv"
import_and_test_csv "steuertarife" "$SAMPLES_DIR/Steuertarife.csv"
import_and_test_csv "vermoegensklassen" "$SAMPLES_DIR/Vermoegensklassen.csv"
import_and_test_csv "distribution_of_wealth_switzerland" "$SAMPLES_DIR/distribution_of_wealth_switzerland.csv"
import_and_test_csv "customer_complaints" "$SAMPLES_DIR/customer_complaints.csv"
import_and_test_csv "whitehouse_visits_2014" "$SAMPLES_DIR/whitehouse_visits_2014.csv"
import_and_test_csv "traffic_violations" "$SAMPLES_DIR/traffic_violations.csv"
import_and_test_csv "metadatenbank_vernehmlassungen_ogd_v1_3" "$SAMPLES_DIR/Metadatenbank-Vernehmlassungen-OGD-V1-3.csv"
import_and_test_csv "wealth_groups" "$SAMPLES_DIR/Wealth_groups.csv"
import_and_test_json "filepaths_1" "$SAMPLES_DIR/filepaths-1.json"
