package test_schema

import (
	"database/sql"
	"dbmap"
	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	_ "github.com/google/uuid"
	_ "github.com/lib/pq"
	"log"
	"os"
	"reflect"
	"testing"
	"time"
)

var db *sql.DB = nil

var enabled = true

func newUUID() uuid.UUID {
	var u uuid.UUID
	u, _ = uuid.NewRandom()
	return u
}

func toPointer(v string) *string {
	return &v
}

func setupTestCase(t *testing.T) func(t *testing.T) {
	configFile := os.Getenv("CONFIG")
	t.Logf("setup test case, using config: %s", configFile)

	if db == nil {
		var cfg dbmap.Config
		dbmap.ReadFile(&cfg, configFile)

		t.Logf("Connecting to %s://user=%s:%s/%s\n", cfg.Database.Provider, cfg.Database.Username, cfg.Database.Host, cfg.Database.Database)

		dataSource := "host=" + cfg.Database.Host + " port=" + cfg.Database.Port +
			" user=" + cfg.Database.Username + " password=" + cfg.Database.Password +
			" dbname=" + cfg.Database.Database + " sslmode=disable"

		var err error
		db, err = sql.Open("postgres", dataSource)
		if err != nil {
			t.Fatal(err)
		}

		db.SetMaxOpenConns(5)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(time.Hour)

		if err = db.Ping(); err != nil {
			t.Fatal(err)
		}

		tables := []string{"test_schema.user", "foo"}
		err = cleanTables(db, tables)
		if err != nil {
			t.Fatal("Failed to cleanup data")
		}
	}

	return func(t *testing.T) {
		t.Log("teardown test case")
	}
}

func cleanTables(db *sql.DB, tables []string) error {
	for _, tableName := range tables {
		rows, err := db.Query("DELETE FROM " + tableName)
		if err != nil {
			log.Print(err)
			rows.Close()
			return err
		}
		rows.Close()
	}
	return nil
}

func TestCreate(t *testing.T) {
	t.Log("----------- TestCreate start -----------")

	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	badUser := User{FirstName: proto.String("Bryan"), LastName: proto.String("Hughes"), Email: proto.String("bh@gmail.com") }
	if err := badUser.Create(db); err == nil {
		t.Fatal("Failed to catch error")
	}

	var cases = []User{
		{FirstName: proto.String("Bryan"), LastName: proto.String("Hughes"), Email: proto.String("bh@gmail.com"), UserToken: toPointer(newUUID().String()), Enabled: &enabled},
		{FirstName: proto.String("Tom"), LastName: proto.String("Bagby"), Email: proto.String("tb@gmail.com"), UserToken: toPointer(newUUID().String()), Enabled: &enabled},
		{FirstName: proto.String("Alice"), LastName: proto.String("Tenfeet"), Email: proto.String("alice@tenfeet.com"), UserToken: toPointer(newUUID().String()), Enabled: &enabled},
		{FirstName: proto.String("Mary"), LastName: proto.String("Littlelamb"), Email: proto.String("mary@gmail.com"), UserToken: toPointer(newUUID().String()), Enabled: &enabled},
	}

	var err error
	var user *User
	for i := 0; i < len(cases); i++ {
		user = &cases[i]

		err = user.Create(db)
		if user.UserId == nil || err != nil {
			t.Fatalf("Failed to create user record. Got a nil UserId instead of a database sequence - %s", err)
		}

		user1 := &User{}
		err = user1.Read(db, user.UserId)
		if user.UserId == nil || err != nil {
			t.Fatalf("Failed to read user record. Got back a nil UserId instead of %d - %s", user.UserId, err)
		}

		if ! reflect.DeepEqual(user, user1) {
			t.Fatal("user an user1 are not equal")
		}

		fName := "Yeti"
		lat := 37.763964
		lon := -122.388983

		user1.FirstName = &fName
		user1.Lat = &lat
		user1.Lon = &lon

		err = user1.Update(db)
		if err != nil {
			t.Fatalf("Failed to update user record - %s", err)
		}

		err = user1.Read(db, user.UserId)
		if err != nil {
			t.Fatalf("Failed to read user record - %s", err)
		}

		if *user1.FirstName != fName {
			t.Fatal("Failed to read back first_name change")
		}

		if *user1.Lat != lat {
			t.Fatal("Failed to read back lat change")
		}

		if *user1.Lon != lon {
			t.Fatal("Failed to read back lon change")
		}

		cases[i] = *user1
	}

	// Test lookups
	user = &cases[1]
	user1 := &User{}
	err = user1.LookupEmail(db, user.Email)
	if user1.UserId == nil || err != nil {
		t.Fatalf("Failed to lookup user record. Got back a nil UserId instead of %d - %s", user1.UserId, err)
	}

	if ! reflect.DeepEqual(user, user1) {
		t.Fatal("lookup up does not match")
	}

	var list []User
	var cnt int32
	list, cnt, err = ListUsers(db,100, 0)

	if len(list) != 4 {
		t.Fatalf("Expected 4 users but got %d", len(list))
	}

	if cnt != 5 {
		t.Fatalf("Expected 5 cnt but got %d", cnt)
	}

	var count int64
	u := newUUID()
	bvalue, _ := u.MarshalBinary()
	count, err = UpdatePasswordHash(db, user.Email, bvalue)
	if count != 1 {
		t.Fatal("Expected 1 update")
	}

	var results []map[string]interface{}
	results, err = GetPasswordHash(db, user.Email)
	if results == nil {
		t.Fatal("Expected a non nil result")
	}

	v := results[0]["pword_hash"]
	if ! reflect.DeepEqual(v, bvalue) {
		t.Fatalf("Got %s instead of %s", v, u)
	}

	// Test delete
	for i := 0; i < len(cases); i++ {
		user = &cases[i]

		var count int64
		count, err = user.Delete(db)
		if count != 1 || err != nil {
			t.Fatalf("Failed to delete user record. Got back a nil UserId instead of %d - %s", user.UserId, err)
		}

		err = user.Read(db, user.UserId)
		if err != nil {
			t.Fatalf("Failed to read user record - %s", err)
		}

		if user.UserId != nil {
			t.Fatal("Should not have read")
		}
	}

	t.Log("----------- TestCreate done -----------")
}
