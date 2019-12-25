package test_schema

import (
	"database/sql"
	"errors"
	"log"
	"model"
)

// Standard CRUD
const selectWithLimitStr = "SELECT user_id, first_name, last_name, email, user_token, enabled, aka_id, ST_Y(geog::geometry) AS lat, ST_X(geog::geometry) AS lon FROM test_schema.user LIMIT $1 OFFSET $2"
const selectStr = "SELECT user_id, first_name, last_name, email, user_token, enabled, aka_id, ST_Y(geog::geometry) AS lat, ST_X(geog::geometry) AS lon FROM test_schema.user WHERE user_id=$1"
const insertStr = "INSERT INTO test_schema.user (first_name, last_name, email, user_token, enabled, aka_id, geog) VALUES ($1, $2, $3, $4, $5, $6, ST_POINT($7, $8)::geography) RETURNING user_id, first_name, last_name, email, user_token, enabled, aka_id, ST_Y(geog::geometry) AS lat, ST_X(geog::geometry) AS lon"
const updateStr = "UPDATE test_schema.user SET first_name=$2, last_name=$3, email=$4, user_token=$5, enabled=$6, aka_id=$7, geog=ST_POINT($8, $9)::geography WHERE user_id=$1 RETURNING user_id, first_name, last_name, email, user_token, enabled, aka_id, ST_Y(geog::geometry) AS lat, ST_X(geog::geometry) AS lon"
const deleteStr = "DELETE FROM test_schema.user WHERE user_id=$1"

// Lookups/Search
const lookupEmailStr = "SELECT user_id, first_name, last_name, email, user_token, enabled, aka_id, ST_Y(geog::geometry) AS lat, ST_X(geog::geometry) AS lon FROM test_schema.user WHERE email=$1"

// Custom Mappings
var customMappings = map[string]string{
	"UpdatePasswordHash": "UPDATE test_schema.user SET pword_hash = $2 WHERE email = $1",
	"GetPasswordHash":    "SELECT pword_hash FROM test_schema.user WHERE email = $1",
	"ResetPasswordHash":  "UPDATE test_schema.user SET pword_hash = NULL WHERE email = $1",
	"DisableUser":        "UPDATE test_schema.user SET enabled = false WHERE email = $1",
	"EnableUser":         "UPDATE test_schema.user SET enabled = true WHERE email = $1",
	"DeleteUserByEmail":  "DELETE FROM test_schema.user WHERE email = $1",
	"SetToken":           "UPDATE test_schema.user SET user_token = uuid_generate_v4() WHERE user_id = $1 RETURNING user_token",
	"FindNearest":        "SELECT address_id, address1, address2, city, state, country, postcode, ST_X(geog::geometry) AS lon, ST_Y(geog::geometry) AS lat FROM address WHERE ST_DWithin( geog, Geography(ST_MakePoint($1, $2)), $3 ) AND lat != 0.0 AND lng != 0.0 ORDER BY geog <-> ST_POINT($1, $2)::geography",
}

type nullableUser struct {
	userId    sql.NullInt32   // Serial data types MUST be Nullable even though they are the primary key
	firstName sql.NullString  // Nullable
	lastName  sql.NullString  // Nullable
	email     string		  // Not Null
	userToken string          // Not Null
	enabled   bool            // Not Null
	akaId     sql.NullInt32   // Nullable
	lat       sql.NullFloat64 // Nullable
	lon       sql.NullFloat64 // Nullable
}

type NearestLocation struct {
	AddressId int32
	Address1  string
	Address2  sql.NullString // Nullable
	City      string
	State     string
	Country   string
	Postcode  string
	lon       sql.NullFloat64 // Nullable
	lat       sql.NullFloat64 // Nullable
}

type PasswordHash struct {
	PwordHash string
}

func toNullableUser(user *User) nullableUser {
	return nullableUser{
		userId:    model.SetNullInt32(user.UserId),
		firstName: model.SetNullString(user.FirstName),
		lastName:  model.SetNullString(user.LastName),
		email:     *user.Email,     // QUESTION: Should I let the database handle the failure of should my code do the check?
		userToken: *user.UserToken, // Not null with a default value
		enabled:   *user.Enabled,   // Not null with a default value
		akaId:     model.SetNullInt32(user.AkaId),
		lat:       model.SetNullFloat64(user.Lat),
		lon:       model.SetNullFloat64(user.Lon),
	}
}

func fromNullableUser(user *User, nUser nullableUser) {
	user.UserId = model.SetInt32(nUser.userId)
	user.FirstName = model.SetString(nUser.firstName)
	user.LastName = model.SetString(nUser.lastName)
	user.Email = &nUser.email
	user.UserToken = &nUser.userToken
	user.Enabled = &nUser.enabled
	user.AkaId = model.SetInt32(nUser.akaId)
	user.Lat = model.SetFloat64(nUser.lat)
	user.Lon = model.SetFloat64(nUser.lon)
}

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
	if err := rows.Scan(&returning.userId, &returning.firstName, &returning.lastName, &returning.email, &returning.userToken, &returning.enabled, &returning.akaId, &returning.lat, &returning.lon); err != nil {
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

func validateNotNulls(m *User) (err error) {
	if m.UserToken == nil {
		return errors.New("user_db: UserToken is defined as not null but has a null value")
	} else if m.Enabled == nil {
		return errors.New("user_db: Enabled is defined as not null but has a null value")
	}
	return nil
}

func (m *User) Read(db *sql.DB, userId *int32) (err error) {
	rows, err := db.Query(selectStr, userId)
	if err != nil {
		log.Print(err)
		return err
	}
	defer rows.Close()

	var returning = nullableUser{}
	if rows.Next() {
		if err := rows.Scan(&returning.userId, &returning.firstName, &returning.lastName, &returning.email, &returning.userToken, &returning.enabled, &returning.akaId, &returning.lat, &returning.lon); err != nil {
			log.Print(err)
			return err
		}

		if err := rows.Err(); err != nil {
			log.Print(err)
			return err
		}

		fromNullableUser(m, returning)
	} else {
		m.Reset()
	}

	return nil
}

func (m *User) Update(db *sql.DB) (err error) {
	if err := validateNotNulls(m); err != nil {
		log.Print(err)
		return err
	}

	nullable := toNullableUser(m)
	rows, err := db.Query(updateStr, nullable.userId, nullable.firstName, nullable.lastName, nullable.email, nullable.userToken, nullable.enabled, nullable.akaId, nullable.lon, nullable.lat)
	if err != nil {
		log.Print(err)
		return err
	}
	defer rows.Close()

	var returning = nullableUser{}
	rows.Next()
	if err := rows.Scan(&returning.userId, &returning.firstName, &returning.lastName, &returning.email, &returning.userToken, &returning.enabled, &returning.akaId, &returning.lat, &returning.lon); err != nil {
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

func (m *User) Delete(db *sql.DB) (count int64, err error) {
	rows, err := db.Exec(deleteStr, m.UserId)
	if err != nil {
		log.Print(err)
		return 0, err
	} else {
		return rows.RowsAffected()
	}
}

func ListUsers(db *sql.DB, limit int32, offset int32) (user []User, count int32, err error) {
	rows, err := db.Query(selectWithLimitStr, limit, offset)
	if err != nil {
		log.Print(err)
		return []User{}, 0, err
	}
	defer rows.Close()

	count = 1
	var results []User
	var returning = nullableUser{}
	for rows.Next() {
		user := User{}
		if err := rows.Scan(&returning.userId, &returning.firstName, &returning.lastName, &returning.email, &returning.userToken, &returning.enabled, &returning.akaId, &returning.lat, &returning.lon); err != nil {
			log.Print(err)
			return []User{}, 0, err
		}

		if err := rows.Err(); err != nil {
			log.Print(err)
			return []User{}, 0, err
		}

		fromNullableUser(&user, returning)
		results = append(results, user)
		count++
	}

	return results, count, nil
}

func (m *User) LookupEmail(db *sql.DB, email *string) (err error) {
	rows, err := db.Query(lookupEmailStr, email)
	if err != nil {
		log.Print(err)
		return err
	}
	defer rows.Close()

	var returning = nullableUser{}
	if rows.Next() {
		if err := rows.Scan(&returning.userId, &returning.firstName, &returning.lastName, &returning.email, &returning.userToken, &returning.enabled, &returning.akaId, &returning.lat, &returning.lon); err != nil {
			log.Print(err)
			return err
		}

		if err := rows.Err(); err != nil {
			log.Print(err)
			return err
		}

		fromNullableUser(m, returning)
	} else {
		m.Reset()
	}

	return nil
}

func UpdatePasswordHash(db *sql.DB, value1 *string, value2 []byte) (count int64, err error) {
	rows, err := db.Exec(customMappings["UpdatePasswordHash"], value1, value2)
	if err != nil {
		log.Print(err)
		return 0, err
	} else {
		return rows.RowsAffected()
	}
}

func GetPasswordHash(db *sql.DB, value1 *string) (results []map[string]interface{}, err error) {
	rows, err := db.Query(customMappings["GetPasswordHash"], value1)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	defer rows.Close()

	results = make([]map[string]interface{}, 0)
	cols, _ := rows.Columns()
	for rows.Next() {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i, _ := range columns {
			columnPointers[i] = &columns[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			log.Print(err)
			return nil, err
		}

		if err := rows.Err(); err != nil {
			log.Print(err)
			return nil, err
		}

		m := make(map[string]interface{})
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			m[colName] = *val
		}
		results = append(results, m)
	}

	return results, nil
}

