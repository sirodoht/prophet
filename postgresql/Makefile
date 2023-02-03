.PHONY: init
init:
	PGDATA=postgres-data/ pg_ctl init
	PGDATA=postgres-data/ pg_ctl start
	createuser prophet
	psql -U sirodoht -d postgres -c "ALTER USER prophet CREATEDB;"
	psql -U prophet -d postgres -c "CREATE DATABASE prophet;"
	psql -U prophet -d prophet -f schema.sql

.PHONY: start
start:
	PGDATA=postgres-data/ pg_ctl start

.PHONY: stop
stop:
	PGDATA=postgres-data/ pg_ctl stop
