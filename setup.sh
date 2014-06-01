#!/bin/bash

mkdir -p /gisdata/temp /gisdata/mapblue

go build load_census_data.go
go build serve_census_data.go
cp install_gis.sh load_census_data serve_census_data /gisdata/mapblue
sudo chown postgres:users /gisdata -R

