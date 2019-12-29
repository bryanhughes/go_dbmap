go_dbmap
=====

A Go application that generates Go code that handles Search, Create, Read, Update, and Delete (SCRUD) operations against 
a relational database. This tool can also generate the Protobuf .proto files based on the schema. Please note that the 
tool does not yet support complex many-to-many relationships cleanly. 

Aside from simple CRUD, this tool allows you to generate functions based on any standard SQL query, lookups, and
transformations. The custom mappings and transformations are very powerful.

Please look at the `config` directory to an example configuration file based on the test schema to generate not just simple
CRUD, but search/lookup operations based on indexed fields, as well as custom mapping queries. 

#### Setting Up Database For Testing and Building Example Code

```
$ sudo -u postgres createuser --pwprompt --superuser dbmap_test
Enter password for new role: dbmap_test
Enter it again: dbmap_test
$ psql -U dbmap_test -h localhost -c "CREATE DATABASE dbmap_test WITH OWNER = dbmap_test ENCODING = 'UTF8' TEMPLATE = template0 CONNECTION LIMIT = -1;" postgres
$ psql -U dbmap_test -h localhost -W -c "CREATE EXTENSION postgis;" dbmap_test
```

#### Resetting the Example Database
This helper script will allow you to rapidly drop and recreate your database. Currently only postgres is supported.

NOTE: depending on your dev environment, you might have to use your own user name instead of `dbmap_test`

    $ bin/reset_db.sh postgres dbmap_test

To filter the noisy output, you can just grep for errors. Please note that the DbSchema tool will create a public
schema that is already present, so you can ignore that error

    $ bin/reset_db.sh postgres dbmap_test | grep error
    
#### Running the unit tests
The unit tests are inline with the code at the end of the modules. Several of them expect that `dbmap_test` database
has been created.

    $ go test

Run the tests again after you have generate the example code to then test the generated code `user_db` against the
`user` table to make sure everything works.

#### Building the Example
The project includes a fairly comprehensive test schema and scripts to build located in the `fdatabase` directory. You
will find the build scripts in `bin` to create or reset the example database as well as to generate
the code in the test schema SQL, which is located in each of the supported databases. The schema was generated
using [DbSchema](https://www.dbschema.com/).

**PLEASE NOTE:** You will see errors and warnings in the output. This is intentional as the test schema includes a
lot of corner cases, like a table without a primary key.

#### Testing the Generated Code
Please note that there are two eunit tests which will test the generated code. They are located in `go_dbmap` and are 
called `crud_test` and `change_id_test`.
    
Using go_dbmap
---
Include this as a dependency in your `rebar.config`. I would recommend that you copy the script
`generate_code.sh` to your project and modify accordingly. You will need to run this the first time, and
any other time you alter your database schema. 

### go_dbmap.yml
You will want to look at [config/go_dbmap.config](config/go_dbmap.config) as a guide
for your own YAML config. It gives a complete example with inline documentation of the current functionality of the tool.

`go_dbmap` provides you a lot of options for how to generate the code. As mentioned previously, `go_dbmap`
supports database objects that span multiple schema's. To generate code against them, simply include all the
schema's in a list. 

**Please note**, if your application dynamically generates schema's as a pattern for supporting multi-tenancy where
each schema is owned by a tenant, this tool will not work as it requires schema's and tables that have been statically
created.

```yaml
  schemas: ["public", "test_schema"]
```

Often, there might be meta tables or other tables that you to exclude (such as the supporting tables installed 
from the postgis extension). Simply list the tables. Note that all tables support the explicit "schema.table" 
naming. If the schema portion is left out, then "public" is defaulted.

```yaml
   excluded_tables: ["excluded", "spatial_ref_sys"]
```

`go_dbmap` provides a feature to generate all the lookup accessors based on defined indexes in the schema.
```yaml
  indexed_lookups: true
```

There are occasionally columns that have sensitive values, like a password hash that you do not want as part of
the default SELECT (which in turns means they will be absent from the generated INSERT and UPDATE functions/queries).
excluded_columns:

```yaml
    -
      table: "test_schema.user"
      columns: ["pword_hash", "geog"]
```

One of the more powerful features of `go_dbmap` is the ability to define SQL queries and have them mapped to 
methods. These are called custom mappings. You can use these to hide operations on sensitive columns such as a
password hash that you do not want exposed through the type struct and common CRUD operation, which is often
exposed to your API.

Custom query mappings will generate a function that will return a result map of column/value from the provided
query. For UPDATE and DELETE, it will return the operations response. If you have any questions, you can build the
example code and review the generated code.

`go_dbmap` needs some information when defining the mappings. Any bind parameter that you would normally write the
query with a place holder (like `$` for Postgres), you will need to expand what the name of the argument and its
expected type. The code generator will then use these in the mappings. Currently `go_dbmap` does not parse queries
to extract the native data type in the schema directly.

```yaml
  mapping:
    -
      table: "test_schema.user"
      queries:
        -
          name: "update_pword_hash"
          query: "UPDATE test_schema.user SET pword_hash = $pwordHash:string WHERE email = $email:string"
        -
          name: "get_pword_hash"
          query: "SELECT pword_hash FROM test_schema.user WHERE email = $email:string"
          
    <snip>
```

Finally `go_dbmap` supports applying transformations to values as they are read and written from the data 
mapping code to the table. A good use of this would be to convert the geog value from postgis to lat and lng. 
While you can do custom query mappings to get the lat,lng, this will put the logic in the CRUD. Because 
`go_dbmap` operates against database SQL operations, it only support transformations between columns that 
can be legally defined by SQL. This implementation is not terrible sophisticated, it is important to understand 
that record map for the specified table is defined by the SELECT statement, which are all the columns except 
those specifically excluded. This means that when you specify a select transform, if those columns do not 
exist in the table, they will generated in the record map and returned on each read. This means than you 
can then use those virtual columns generated by the SELECT/READ to then be used on any writes (INSERT/UPDATE). 
The following demonstrates how to support lat,lon in with a postgis geography column.

***NOTE:*** any columns that are referenced in a function must be preceded by a `$`
```yaml
  transforms:
    -
      table: "test_schema.user"
      xforms:
        # For the select transform, we need to know the datatype of the product of the transform. This is needed for
        # generating the protobufs
        select:
          -
            column: "lat"
            data_type: "decimal"
            xform: "ST_Y(geog::geometry)"
          -
            column: "lon"
            data_type: "decimal"
            xform: "ST_X(geog::geometry)"
        insert:
          -
            column: "geog"
            data_type: "geography"
            xform: "ST_POINT($lat, $lon)::geography"
        update:
          -
            column: "geog"
            data_type: "geography"
            xform: "ST_POINT($lat, $lon)::geography"
    -
      table: "public.foo"
      xforms:
        select:
          -
            column: "foobar"
            data_type: "integer"
            xform: "1"
```

In the case of converting a `lat` and `lon` to a `geography`, you must define each of the operations insert/create,
update, and select/read on how the column values will be handled to and from the database. The result is that the
columns `lat` and `lon` will be generated as `virtual` columns in the mapping. Note that when referencing them in
the function body (the second element of the tuple), you will need to prepend them with the `$` so that `go_dbmap`
knows they are the virtual columns being operated on.

For the `insert` operation, a single tuple is defined which will result in the extension function 
`ST_POINT($lat, $lon)::geography` to be applied to the bind values of the `INSERT` statement. Resulting in the 
following code:

```gotemplate
    const insertStr = "INSERT INTO test_schema.user (first_name, last_name, email, user_token, enabled, aka_id, geog) VALUES ($1, $2, $3, $4, $5, $6, ST_POINT($7, $8)::geography) RETURNING user_id, first_name, last_name, email, user_token, enabled, aka_id, ST_X(geog::geometry) AS lon, ST_Y(geog::geometry) AS lat"
```
and
```
func (m *User) Create(db *sql.DB) (err error) {
	if err := validateNotNulls(m); err != nil {
		log.Print(err)
		return err
	}

	nullable := toNullableUser(m)
	rows, err := db.Query(insertStr, nullable.firstName, nullable.lastName, nullable.email, nullable.userToken, nullable.enabled, nullable.akaId, nullable.lon, nullable.lat)
	if err != nil {
		log.Print(err)
		return err
	}
	defer rows.Close()

	var returning = nullableUser{}
	rows.Next()
	if err := rows.Scan(&returning.userId, &returning.firstName, &returning.lastName, &returning.email, &returning.userToken, &returning.enabled, &returning.akaId, &returning.lon, &returning.lat); err != nil {
		log.Print(err)
		return err
	}

	if err := rows.Err(); err != nil {
		log.Print(err)
		return err
	}

	fromNullableUser(m, returning)
	return nil
}
```

I would recommend that you build the example project and then review the generated code for `user_db.go` to get a better
understanding.

The more complex feature of `go_dbmap` is the ability to apply transformations or sql functions. If you are using PostGIS,
or need to apply other functions, you will need to use this feature.
    
Generating Protobuffers
---
go_dbmap will generate protobuffers that map to the relational schema including foreign key relationships. The tool
correctly handles relationships across schemas. The generated `Go` code then requires compiled protobuffer code, so
`user_db.go` requires `user_pb.go`.

*TODO:*
While the tool will handle many-to-many relationships through the join table, it does not yet add the relating message 
in the referencing message. For example if the join table is many customers with many addresses, the tool does not yet 
pull the many addresses into the customer message as a repeating Address message.  

The package for each proto will be the schema that the table is located in. The `.proto` files will be generated in 
the `output` directory specified in the config file with those proto to table mappings being written to a subdirectory 
that corresponds to the schema the table is in.  

The tool will correctly generate proto files with foreign key relationships that cross schemas. It will also handle
naming collisions due to tables with the same names in different schemas by appending a number that increments starting
at 1. A namespace of schema_table was considered, but rejected for the time being due to the very lon field names that
can be generated.

It is recommended that you download and install the latest protocol buffer compiler. If you are new to protocol buffers, start
[by reading the developer docs](https://developers.google.com/protocol-buffers/).

Using In Your Project
---
...

Next, create your `go_dbmap.config`. You can simply copy and modify the one in the repo. Also, copy and rename
the `example/bin/build_example.sh` and move it into your `bin` or other directory. Run it once to generate the 
code and protos, and any other time you modify your schema.