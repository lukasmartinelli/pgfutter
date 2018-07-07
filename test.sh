#!/bin/bash
readonly CWD=$(pwd)
readonly SAMPLES_DIR="$CWD/samples"
readonly DB_USER="${DB_USER:-postgres}"
readonly DB_NAME="integration_test"
readonly DB_SCHEMA="import" # Use public schema instead of import because of permissions

function recreate_db() {
  psql -U "${DB_USER}" -c "drop database if exists ${DB_NAME};"
  psql -U "${DB_USER}" -c "create database ${DB_NAME};"
}

function query_counts() {
    local table="$1"
    local counts=$(psql -U "${DB_USER}" -d "${DB_NAME}" -t -c "select count(*) from ${DB_SCHEMA}.${table}")
    echo "$counts"
}

function query_field_type() {
    local table="$1"
    local data_type=$(psql -U "${DB_USER}" -d "${DB_NAME}" -t -c "SELECT data_type FROM information_schema.columns WHERE table_schema='${DB_SCHEMA}' AND table_name='${table}'")
    echo "$data_type"
}

function import_csv_with_special_delimiter_and_trailing() {
    local table="csv_sample_qip12_tabdaten"
    local filename="$SAMPLES_DIR/csv_sample_qip12_tabdaten.csv"
    pgfutter --schema "$DB_SCHEMA" --db "$DB_NAME" --user "$DB_USER" csv "$filename" --delimiter=";"
    if [ $? -ne 0 ]; then
        echo "pgfutter could not import $filename"
        exit 300
    else
        local db_count=$(query_counts $table)
        echo "Imported $(expr $db_count) records into $table"
    fi
}

function import_csv_and_skip_header_row_with_custom_fields() {
    local table="csv_sample_qip12_tabdaten"
    local filename="$SAMPLES_DIR/csv_sample_qip12_tabdaten.csv"
    pgfutter --schema "$DB_SCHEMA" --db "$DB_NAME" --user "$DB_USER" csv --delimiter ";" "$filename"
}

function test_json() {
    local table=$1
    local filename=$2
    pgfutter --schema "$DB_SCHEMA" --db "$DB_NAME" --user "$DB_USER" json "$filename"
    if [ $? -ne 0 ]; then
        echo "pgfutter could not import $filename"
        exit 300
    else
        local db_count=$(query_counts $table)
        local data_type=$(query_field_type $table)
        echo "Imported $(expr $db_count) records into $table as $data_type"
    fi
}

function test_json_as_jsonb() {
    local table=$1
    local filename=$2
    pgfutter --schema "$DB_SCHEMA" --db "$DB_NAME" --user "$DB_USER" --jsonb json "$filename"
    if [ $? -ne 0 ]; then
        echo "pgfutter could not import $filename"
        exit 300
    else
        local db_count=$(query_counts "$table")
        local data_type=$(query_field_type "$table")
        echo "Imported $(expr $db_count) records into $table as $data_type"
    fi
}

function test_excel_csv() {
    local table=$1
    local filename=$2
    local delimiter=${3:-,}
    local general_args=${4:-}

    pgfutter $general_args --table "$table" --schema "$DB_SCHEMA" --db "$DB_NAME" --user "$DB_USER" csv "$filename" --delimiter "$delimiter" --excel
    if [ $? -ne 0 ]; then
        echo "pgfutter could not import $filename"
        exit 300
    else
        local db_count=$(query_counts $table)
        echo "Imported $(expr $db_count) records into $table from $filename"
    fi
}

function test_csv() {
    local table=$1
    local filename=$2
    local delimiter=${3:-,}
    local general_args=${4:-}

    pgfutter $general_args --table "$table" --schema "$DB_SCHEMA" --db "$DB_NAME" --user "$DB_USER" csv "$filename" --delimiter "$delimiter"
    if [ $? -ne 0 ]; then
        echo "pgfutter could not import $filename"
        exit 300
    else
        local db_count=$(query_counts $table)
        echo "Imported $(expr $db_count) records into $table from $filename"
    fi
}

recreate_db


# We change the type of the data column for this test, so we have to recreate the database
recreate_db

#TODO does not work cause quoted multiline char
# test_csv "local_severe_wheather_warning_systems" "$SAMPLES_DIR/csv_sample_local_severe_wheather_warning_systems.csv"
#TODO does not work cause quoted multiline char
# test_csv "residential_permits" "$SAMPLES_DIR/csv_sample_residential_permits.csv"
test_csv "distribution_of_wealth_switzerland" "$SAMPLES_DIR/csv_sample_distribution_of_wealth_switzerland.csv"
test_excel_csv "techcrunch_continental_usa" "$SAMPLES_DIR/csv_sample_techcrunch_continental_usa.csv"
test_csv "employee_salaries" "$SAMPLES_DIR/csv_sample_employee_salaries.csv"
test_csv "montgomery_crime" "$SAMPLES_DIR/csv_sample_montgomery_crime.csv"
test_excel_csv "sacramentocrime_jan_2006" "$SAMPLES_DIR/csv_sample_sacramentocrime_jan_2006.csv"
test_excel_csv "sacramento_realestate_transactions" "$SAMPLES_DIR/csv_sample_sacramento_realestate_transactions.csv"
test_excel_csv "sales_jan_2009" "$SAMPLES_DIR/csv_sample_sales_jan_2009.csv"
test_csv "steuertarife" "$SAMPLES_DIR/csv_sample_steuertarife.csv"
test_csv "vermoegensklassen" "$SAMPLES_DIR/csv_sample_vermoegensklassen.csv"

import_csv_and_skip_header_row_with_custom_fields
import_csv_with_special_delimiter_and_trailing

test_json_as_jsonb "json_sample_2015_01_01_15" "$SAMPLES_DIR/json_sample_2015-01-01-15.json"

recreate_db
