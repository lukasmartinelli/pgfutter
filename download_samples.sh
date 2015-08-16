#!/bin/bash
CWD=$(pwd)
SAMPLES_DIR="$CWD/samples"

function download_csv_samples() {
    mkdir -p $SAMPLES_DIR
    cd $SAMPLES_DIR
    wget -nc -O customer_complaints.csv https://data.consumerfinance.gov/api/views/x94z-ydhh/rows.csv
    wget -nc -O traffic_violations.csv https://data.montgomerycountymd.gov/api/views/4mse-ku6q/rows.csv
    cd $CWD
}

download_csv_samples
