#!/bin/bash

psql -c "DROP DATABASE census;"
psql -c "CREATE DATABASE census WITH OWNER census;"

psql -d census -c "CREATE EXTENSION postgis;"
psql -d census -c "CREATE EXTENSION fuzzystrmatch;"
psql -d census -c "CREATE EXTENSION postgis_tiger_geocoder;"
psql -d census -c "CREATE EXTENSION postgis_topology;"
for schema in public tiger topology
do
    psql -d "GRANT USAGE ON SCHEMA $schema TO census;"
    psql -d census -c "GRANT SELECT ON ALL TABLES IN SCHEMA $schema TO census;"
done

psql -A -t -d census -c \
    "SELECT loader_generate_script(ARRAY['IN'], 'sh');" > \
    load_indiana_tiger_data.sh

