#!/usr/bin/env bash

if [[ $# -lt 3 ]]; then
    echo "usage: reset_db user dbname schema_file seed_file"
    exit 1
fi

echo "Dropping Postgres database $2..."
dropdb "$2"

echo "Recreating Postgres database $2..."
createdb -U "$1" "$2"

psql "$2" << EOF
    GRANT ALL PRIVILEGES ON DATABASE $2 TO $1;
    CREATE EXTENSION if not exists "postgis";
    CREATE EXTENSION if not exists "uuid-ossp";
EOF

psql -d "$2" -a -f "$3"

if [[ $# -eq 4 ]]; then
    echo "Loading seed data..."
    psql -d "$2" -a -f "$4"
fi
