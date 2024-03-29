package model

import (
	"database/sql"
	"log"
	"time"
)

func SetNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{
			String: "",
			Valid:  false,
		}
	} else {
		return sql.NullString{
			String: *s,
			Valid:  true,
		}
	}
}

func SetString(s sql.NullString) *string {
	if s.Valid {
		return &s.String
	} else {
		return nil
	}
}

func SetNullInt32(i *int32) sql.NullInt32 {
	if i == nil {
		return sql.NullInt32{
			Int32: 0,
			Valid: false,
		}
	} else {
		return sql.NullInt32{
			Int32: *i,
			Valid: true,
		}
	}
}

func SetInt32(i sql.NullInt32) *int32 {
	if i.Valid {
		return &i.Int32
	} else {
		return nil
	}
}

func SetNullInt64(i *int64) sql.NullInt64 {
	if i == nil {
		return sql.NullInt64{
			Int64: 0,
			Valid: false,
		}
	} else {
		return sql.NullInt64{
			Int64: *i,
			Valid: true,
		}
	}
}

func SetInt64(i sql.NullInt64) *int64 {
	if i.Valid {
		return &i.Int64
	} else {
		return nil
	}
}

func SetNullFloat64(f *float64) sql.NullFloat64 {
	if f == nil {
		return sql.NullFloat64{
			Float64: 0,
			Valid:   false,
		}
	} else {
		return sql.NullFloat64{
			Float64: *f,
			Valid:   true,
		}
	}
}

func SetFloat64(f sql.NullFloat64) *float64 {
	if f.Valid {
		return &f.Float64
	} else {
		return nil
	}
}

func SetNullBool(b *bool) sql.NullBool {
	if b == nil {
		return sql.NullBool{
			Bool:  false,
			Valid: false,
		}
	} else {
		return sql.NullBool{
			Bool:  *b,
			Valid: true,
		}
	}
}

func SetBool(b sql.NullBool) *bool {
	if b.Valid {
		return &b.Bool
	} else {
		return nil
	}
}

func SetNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{
			Time:  time.Time{},
			Valid: false,
		}
	} else {
		return sql.NullTime{
			Time:  *t,
			Valid: true,
		}
	}
}

func SetTime(t sql.NullTime) *time.Time {
	if t.Valid {
		return &t.Time
	} else {
		return nil
	}
}

func ReadResults(rows *sql.Rows, err error) (results []map[string]interface{}, errOut error) {
	if err != nil {
		log.Print(err)
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Print(err)
		}
	}(rows)

	results = make([]map[string]interface{}, 0)
	cols, _ := rows.Columns()
	for rows.Next() {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
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
