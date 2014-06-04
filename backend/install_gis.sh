#!/bin/bash

CENSUSURL="http://www2.census.gov/census_2010/04-Summary_File_1"
INDIANAURL="$CENSUSURL/Indiana/in2010.sf1.zip"

## Create the database
psql -c "DROP DATABASE census;"
psql -c "CREATE DATABASE census WITH OWNER census;"

## Install PostGIS extensions
psql -d census -c "CREATE EXTENSION postgis;"
psql -d census -c "CREATE EXTENSION fuzzystrmatch;"
psql -d census -c "CREATE EXTENSION postgis_tiger_geocoder;"
psql -d census -c "CREATE EXTENSION postgis_topology;"

## Generate the loader script
psql -A -t -d census -c \
    "SELECT loader_generate_script(ARRAY['IN'], 'sh');" > \
    load_indiana_tiger_data.sh

## Download census data files
wget $INDIANAURL && unzip in2010.sf1.zip && rm in2010.sf1.zip

## Patch load_indiana_tiger_data.sh here
# ...

## Load TIGER data
# dtach -n load_indiana_tiger_data.sock ./load_indiana_tiger_data.sh

## Load Census data
# dtach -n load_census_data.sock ./load_census_data

