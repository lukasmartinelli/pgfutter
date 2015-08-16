#!/bin/bash
set -v

# customer complaints
wget -nc -O customer_complaints.csv https://data.consumerfinance.gov/api/views/x94z-ydhh/rows.csv

pgfutter csv customer_complaints.csv
wc -l customer_complaints.csv
psql -U postgres -d integration_test -t -c 'select count(*) from import.customer_complaints'

wget -nc http://www2.census.gov/econ/sbo/07/pums/pums_csv.zip
unzip pums_csv.zip
mv pums.csv business_survey.csv

wget -nc -O traffic_violations.csv https://data.montgomerycountymd.gov/api/views/4mse-ku6q/rows.csv
