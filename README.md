go_dbmap
=====

A Go application that generates Go code that handles Search, Create, Read, Update, and Delete (SCRUD) operations against 
a relational database. This tool can also generate the Protobuf .proto files based on the schema. Please note that the 
tool does not yet support complex many-to-many relationships cleanly. 

Aside from simple CRUD, this tool allows you to generate functions based on any standard SQL query, lookups, and
transformations. The custom mappings and transformations are very powerful.

Please look at the `example` directory to an example schema and configuration file to generate not just simple
CRUD, but search/lookup operations based on indexed fields, as well as custom mapping queries. 

#### Setting Up Database For Building The Examples

```
$ sudo -u postgres createuser --pwprompt --superuser go_dbmap
could not change directory to "/home/...": Permission denied
Enter password for new role: go_dbmap
Enter it again: go_dbmap
$ psql -U go_dbmap -h localhost -c "CREATE DATABASE go_dbmap WITH OWNER = go_dbmap ENCODING = 'UTF8' TEMPLATE = template0 CONNECTION LIMIT = -1;" postgres
$ psql -U go_dbmap -h localhost -W -c "CREATE EXTENSION postgis;" go_dbmap
```

#### Resetting the Example Database
This helper script will allow you to rapidly drop and recreate your database

    $ bin/reset_db.sh go_dbmap go_dbmap example/sql/example_schema.sql example/sql/example_data.sql
    
#### Building the Example
The project includes an example schema and scripts to build located in the `example` directory. You
will find the build scripts in `example/bin` to create or reset the example database as well as to generate
the code in the example schema, which is located in `example/sql`. The schema was generated using [DbSchema](https://www.dbschema.com/).

**PLEASE NOTE:** You will see errors and warnings in the output. This is intentional as the example schema includes a
lot of corner cases, like a table without a primary key.

#### Running the unit tests
The unit tests are inline with the code at the end of the modules. Several of them expect that go_dbmap database
from the example directory to have been built.

    rebar3 eunit
    
You can run the eunit tests again after your generate the example code to then test the generated code `user_db` against the
`user` table to make sure everything works.
    
#### Testing the Generated Code    
Please note that there are two eunit tests which will test the generated code. They are located in `go_dbmap` and are 
called `crud_test` and `change_id_test`.
    
Using go_dbmap
---
Include this as a dependency in your `rebar.config`. I would recommend that you copy the script
`generate_code.sh` to your project and modify accordingly. You will need to run this the first time, and
any other time you alter your database schema. 

### go_dbmap.config
You will want to look at [example/config/go_dbmap.config](example/config/go_dbmap.config) as a guide
for your own YAML config. It gives a complete example with inline documentation of the current functionality of the tool.

```yaml
  lookup:
    -
      table: "test_schema.address"
      columns: [["address1", "address2", "city", "state", "country", "postcode"],
                ["postcode"]]
    -
      table: "test_schema.user"
      columns: [["email"]]
```
This configuration will generate two `lookup/1` functions on the `address` module. It is advised that you have matching indexes on the
table covering the columns. The second config will generate a lookup that covers the three columns so a corresponding 
composite index will be required.

The more complex feature of `go_dbmap` is the ability to apply transformations or sql functions. If you are using PostGIS,
or need to apply other functions, you will need to use this feature.

```
{transforms, [
    {"test_schema.user", [
        {insert, [{"geog", "ST_POINT($lat, $lon)::geography"}]},
        {update, [{"geog", "ST_POINT($lat, $lon)::geography"}]},
        {select, [{"lat", "ST_Y(geog::geometry)"},
                  {"lon", "ST_X(geog::geometry)"}]}]},
    {"public.foo", [
        {select, [{"foobar", "1"}]}]}
]}.
```
Following the same pattern of a list of tables with a list of tuples. In the case of converting a `lat` and `lon` to a 
`geography`, you must define each of the operations insert/create, update, and select/read on how the column values will
be handled to and from the database. The result is that the columns `lat` and `lon` will be generated as `virtual` 
columns in the mapping. Note that when referencing them in the function body (the second element of the tuple), you will
need to prepend them with the `$` so that `go_dbmap` knows they are the virtual columns being operated on. 
For the `insert` operation, a single tuple is defined which will
result in the extension function `ST_POINT($lat, $lon)::geography` to be applied to the bind values of the `INSERT` 
statement. Resulting in the following code:

```erlang
-define(INSERT, "INSERT INTO test_schema.user (first_name, last_name, email, user_token, enabled, change_id, geog) VALUES ($1, $2, $3, $4, $5, 0, ST_POINT($6, $7)::geography) RETURNING user_id").
```
and
```
create(M = #{first_name := FirstName, last_name := LastName, email := Email, user_token := UserToken, enabled := Enabled, lat := Lat, lon := Long}) when is_map(M) ->
    Params = [FirstName, LastName, Email, UserToken, Enabled, Lat, Long],
    case pgo:query(?INSERT, Params) of
        #{command := insert, num_rows := 1, rows := [{UserId}]} ->
            {ok, M#{user_id => UserId, change_id => 0}};
        {error, Reason} ->
            {error, Reason}
    end;
create(_M) ->
    {error, invalid_map}.
```

I would recommend that you build the example project and then review the generated code for `user_db.erl` to get a better
understanding.

#### The special change_id column
The go_dbmap framework implements all the necessary code to support tracking changes to a table that has a column called 
`change_id`. When a table has this column, go_dbmap will generate code that will automatically handle updating the 
column value on a good update, as well as guard against updating the table from a stale map. If the current value of
`change_id` is `100` and you attempt to update using a map that has the change_id value of 90, the update will return
with `not_found`, otherwise update returns `{ok, Map}`.

In our example database, the `user` table has a column named `change_id`. 
```
-define(UPDATE, "UPDATE test_schema.user SET first_name=$2, last_name=$3, email=$4, user_token=$5, enabled=$6, aka_id=$7, change_id=change_id + 1, geog=ST_POINT($9, $10)::geography WHERE user_id=$1 AND change_id<=$8 RETURNING user_id, first_name, last_name, email, user_token, enabled, aka_id, change_id, ST_Y(geog::geometry) AS lat, ST_X(geog::geometry) AS lon").
```

This results in the following code generation and logic for UPDATES. Please note that the INSERT sql and code is also different.
```
update(M = #{user_id := UserId, first_name := FirstName, last_name := LastName, email := Email, user_token := UserToken, enabled := Enabled, aka_id := AkaId, change_id := ChangeId, lat := Lat, lon := Lon}) when is_map(M) ->
    Params = [UserId, FirstName, LastName, Email, UserToken, Enabled, AkaId, ChangeId, Lat, Lon],
    case pgo:query(?UPDATE, Params, #{decode_opts => [{return_rows_as_maps, true}, {column_name_as_atom, true}]}) of
        #{command := update, num_rows := 0} ->
            not_found;
        #{command := update, num_rows := 1, rows := [Row]} ->
            {ok, Row};
        {error, Reason} ->
            {error, Reason}
    end;
update(_M) ->
    {error, invalid_map}.
```

### erleans/pgo

Finally, `go_dbmap` uses [erleans/pgo](https://github.com/erleans/pgo) for its Postgres connectivity (which is currently)
the only supported database. If you want the pgo client to return UUID as binary strings, set an application environment
variable or in your sys.config:

    {pg_types, {uuid, string}},

    {pgo, [{pools, [{default, #{pool_size => 10,
                                host => "127.0.0.1",
                                database => "go_dbmap",
                                user => "go_dbmap",
                                password => "go_dbmap"}}]}]}

Size the pool according to your requirements.    
    
Generating Protobuffers
---
go_dbmap will generate protobuffers that map to the relational schema including foreign key relationships. The tool
correctly handles relationships across schemas. 

*TODO:*
While the tool will handle many-to-many relationships through the 
join table, it does not yet add the relating message in the referencing message. For example if the join table is many
customers with many addresses, the tool does not yet pull the many addresses into the customer message as a repeating
Address message.  

The package for each proto will be the schema that the table is located in. The `.proto` files will be generated in 
the `output` directory specified in the config file with those proto to table mappings being written to a subdirectory 
that corresponds to the schema the table is in.  

The tool will correctly generate proto files with foreign key relationships that cross schemas. It will also handle
naming collisions due to tables with the same names in different schemas by appending a number that increments starting
at 1. A namespace of schema_table was considered, but rejected for the time being due to the very lon field names that
can be generated.

It is recommended that you download and install the protocol buffer compiler. If you are new to protocol buffers, start
[by reading the developer docs](https://developers.google.com/protocol-buffers/).

Using In Your Project
---
You will need to include the deps in your `rebar.config`:

    {go_dbmap, {git, "https://github.com/bryanhughes/go_dbmap.git", {branch, "master"}}},

Next, create your `go_dbmap.config`. You can simply copy and modify the one in the repo. Also, copy and rename
the `example/bin/build_example.sh` and move it into your `bin` or other directory. Run it once to generate the 
code and protos, and any other time you modify your schema.