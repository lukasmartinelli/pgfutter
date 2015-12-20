#!/bin/bash
CWD=$(pwd)
SAMPLES_DIR="$CWD/samples"

function download_json_samples() {
    mkdir -p $SAMPLES_DIR
    cd $SAMPLES_DIR
    wget -nc http://data.githubarchive.org/2015-01-01-15.json.gz && gunzip -f 2015-01-01-15.json.gz
    cd $CWD
}

function download_csv_samples() {
    mkdir -p $SAMPLES_DIR
    cd $SAMPLES_DIR
    wget -nc -O parking_garage_availability.csv https://data.montgomerycountymd.gov/api/views/qahs-fevu/rows.csv
    wget -nc -O local_severe_wheather_warning_systems.csv https://data.mo.gov/api/views/n59h-ggai/rows.csv
    wget -nc -O montgomery_crime.csv https://data.montgomerycountymd.gov/api/views/icn6-v9z3/rows.csv
    wget -nc -O employee_salaries.csv https://data.montgomerycountymd.gov/api/views/54rh-89p8/rows.csv
    wget -nc -O residential_permits.csv https://data.montgomerycountymd.gov/api/views/m88u-pqki/rows.csv
    wget -nc -O customer_complaints.csv https://data.consumerfinance.gov/api/views/x94z-ydhh/rows.csv
    wget -nc -O traffic_violations.csv https://data.montgomerycountymd.gov/api/views/4mse-ku6q/rows.csv
    wget -nc -O distribution_of_wealth_switzerland.csv http://bar-opendata-ch.s3.amazonaws.com/Kanton-ZH/Statistik/Distribution_of_wealth.csv
    wget -nc http://bar-opendata-ch.s3.amazonaws.com/Kanton-ZH/Statistik/Wealth_groups.csv
    wget -nc http://bar-opendata-ch.s3.amazonaws.com/Kanton-ZH/Statistik/Vermoegensklassen.csv
    wget -nc http://bar-opendata-ch.s3.amazonaws.com/Kanton-ZH/Statistik/Steuertarife.csv
    wget -nc http://bar-opendata-ch.s3.amazonaws.com/Kanton-ZH/Statistik/Tax_rates.csv
    wget -nc -O whitehouse_visits_2014.zip https://www.whitehouse.gov/sites/default/files/disclosures/whitehouse_waves-2014_12.csv_.zip && unzip -o  whitehouse_visits_2014.zip && rm -f whitehouse_visits_2014.csv && mv whitehouse_waves-2014_12.csv.csv whitehouse_visits_2014.csv
    wget -nc http://bar-opendata-ch.s3.amazonaws.com/ch.bag/Spitalstatistikdateien/qip/2012/qip12_tabdaten.csv
    wget -nc http://bar-opendata-ch.s3.amazonaws.com/ch.bar.bar-02/Metadatenbank-Vernehmlassungen-OGD-V1-3.csv
    wget -nc https://www.data.gov/app/uploads/2015/08/opendatasites.csv
    cd $CWD
}

download_csv_samples
download_json_samples
