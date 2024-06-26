#!/bin/bash
# This project uses github.com/volatiletech/sqlboiler to generate the database models and query methods.
# It is a "sql first" "orm" which means you need to have an updated local database that is running with the parameters
# in sqlboiler.toml (using db.sh is the same)

cd "$(dirname "$0")" || exit

# todo add check that db is up

# check that sqlboiler is installed
if ! type "sqlboiler" > /dev/null; then
  # note the latest sqlboiler 4.7.0 is a little broken
  go install github.com/volatiletech/sqlboiler/v4@v4.6.0
  go install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql@v4.6.0
fi

# remove old generated stuff
# shellcheck disable=2046
rm $(grep -l "Code generated by SQLBoiler" sqlbdb/*.go)

sqlboiler -c sqlboiler.toml \
          --add-soft-deletes \
          --no-rows-affected \
          --no-tests \
          -o sqlbdb \
          -p sqlbdb \
          psql
