CREATE USER dbmap_test WITH SUPERUSER PASSWORD 'dbmap_test' CREATEDB;
GRANT ALL PRIVILEGES ON DATABASE dbmap_test TO dbmap_test;
CREATE EXTENSION if not exists "uuid-ossp";
CREATE EXTENSION if not exists "postgis";