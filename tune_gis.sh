#!/bin/bash

## Set schema permissions properly
for schema in public tiger topology
do
    psql -d "GRANT USAGE ON SCHEMA $schema TO census;"
    psql -d census -c "GRANT SELECT ON ALL TABLES IN SCHEMA $schema TO census;"
done

## Create indices
psql -d census -c "CREATE UNIQUE INDEX idx_p10_logrecno ON p10 (logrecno);"
psql -d census -c "CREATE UNIQUE INDEX idx_p16_logrecno ON p16 (logrecno);"
psql -d census -c "CREATE UNIQUE INDEX idx_p19_logrecno ON p19 (logrecno);"
psql -d census -c "CREATE UNIQUE INDEX idx_p29_logrecno ON p29 (logrecno);"
psql -d census -c \
    "CREATE INDEX idx_geo_locations_intptlat ON geo_locations (intptlat);"
psql -d census -c \
    "CREATE INDEX idx_geo_locations_intptlon ON geo_locations (intptlon);"
psql -d census -c \
    "CREATE UNIQUE INDEX idx_tabblock_intptlat ON tabblock (intptlat);"
psql -d census -c \
    "CREATE UNIQUE INDEX idx_tabblock_intptlon ON tabblock (intptlon);"

## Vacuum so the indices are effective
VACUUM ANALYZE

