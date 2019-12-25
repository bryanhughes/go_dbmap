package test_schema

import "database/sql"

func (m *Foo) Create(db *sql.DB) (err error) {
	// nullable := toNullableUser(m)
	// rows, err := _db.Query(insertStr, nullable.firstName, nullable.lastName, nullable.email, nullable.userToken, nullable.enabled, nullable.akaId, nullable.lon, nullable.lat)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer rows.Close()
	//
	// var returning = nullableUser{}
	// for rows.Next() {
	// 	if err := rows.Scan(&returning.userId, &returning.firstName, &returning.lastName, &returning.email, &returning.userToken, &returning.enabled, &returning.akaId, &returning.lat, &returning.lon); err != nil {
	// 		log.Fatal(err)
	// 	}
	// }
	//
	// if !rows.NextResultSet() {
	// 	log.Fatalf("expected more result sets: %v", rows.Err())
	// }
	//
	// if err := rows.Err(); err != nil {
	// 	log.Fatal(err)
	// }
	//
	// return fromNullableUser(returning)

	return nil
}
