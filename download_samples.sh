#!/bin/bash
CWD=$(pwd)
SAMPLES_DIR="$CWD/samples"

function download_csv_samples() {
    mkdir -p $SAMPLES_DIR
    cd $SAMPLES_DIR
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
    cd $CWD
}

download_csv_samples
