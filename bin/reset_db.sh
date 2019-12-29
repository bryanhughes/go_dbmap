#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
NAME="dbmap_test"

if [ -z "$1" ] || [ -z "$2" ]; then
  echo
  echo "Usage: reset_db.sh <database> <db_user>"
  echo "       database:    postgres or mariadb"
  echo "       This script will (re)create a database called $NAME using the test_schema and test_data"
  echo "NOTE: Only postgres is currently supported"
  echo
  exit
fi

if [ "$1" = "postgres" ]; then
  echo "Dropping database $NAME..."
  dropdb "$NAME"

  echo "Recreating database $NAME..."
  createdb -U "$2" "$NAME"

psql "$NAME" << EOF
    GRANT ALL PRIVILEGES ON DATABASE $NAME TO $2;
    CREATE EXTENSION if not exists "postgis";
    CREATE EXTENSION if not exists "uuid-ossp";
EOF

  psql -d "$NAME" -a -f "$DIR/../database/postgres/test_schema.sql"

  echo "Loading seed data..."
  psql -d "$NAME" -a -f "$DIR/../database/test_data.sql"
elif [ "$1" = "mariadb" ]; then
  echo "--- COMING SOON ---"
else
  echo "Unsupported database $1"
fi

