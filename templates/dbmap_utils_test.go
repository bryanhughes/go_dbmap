package model

import (
	"database/sql"
	"github.com/bryanhughes/go_dbmap/src/model"
	"testing"
	"time"
)

func TestSetNullBool(t *testing.T) {
	b := true
	nullBool := model.SetNullBool(&b)
	if nullBool.Valid != true {
		t.Error("expected true got false")
	}

	if nullBool.Bool != true {
		t.Error("expected true got false")
	}

	b = false
	nullBool = model.SetNullBool(&b)
	if nullBool.Valid != true {
		t.Error("expected true got false")
	}

	if nullBool.Bool != false {
		t.Error("expected false got true")
	}

	nullBool = model.SetNullBool(nil)
	if nullBool.Valid != false {
		t.Error("expected false got true")
	}

	if nullBool.Bool != false {
		t.Error("expected false got true")
	}
}

func TestSetBool(t *testing.T) {
	nullBool := sql.NullBool{
		Bool:  true,
		Valid: true,
	}
	b := model.SetBool(nullBool)
	if *b != true {
		t.Error("expected true got false")
	}

	nullBool = sql.NullBool{
		Bool:  false,
		Valid: true,
	}
	b = model.SetBool(nullBool)
	if *b != false {
		t.Error("expected false got true")
	}

	nullBool = sql.NullBool{
		Bool:  false,
		Valid: false,
	}
	b = model.SetBool(nullBool)
	if b != nil {
		t.Error("expected nil")
	}
}

func TestSetNullString(t *testing.T) {
	s := new(string)
	*s = "Hello World!"
	nullString := model.SetNullString(s)
	if nullString.Valid != true {
		t.Error("expected true got false")
	}

	if nullString.String != "Hello World!" {
		t.Error("expected Hello World!")
	}

	s = nil
	nullString = model.SetNullString(s)
	if nullString.Valid != false {
		t.Error("expected false got true")
	}
}

func TestSetString(t *testing.T) {
	nullString := sql.NullString{
		String: "Hello World!",
		Valid:  true,
	}
	s := model.SetString(nullString)
	if *s != "Hello World!" {
		t.Error("expected Hello World!")
	}

	nullString = sql.NullString{
		String: "Does not matter",
		Valid:  false,
	}
	s = model.SetString(nullString)
	if s != nil {
		t.Error("expected nil")
	}
}

func TestSetNullInt32(t *testing.T) {
	i := new(int32)
	*i = 23
	nullInt := model.SetNullInt32(i)
	if nullInt.Valid != true {
		t.Error("expected true got false")
	}

	if nullInt.Int32 != 23 {
		t.Error("expected 23")
	}

	i = nil
	nullInt = model.SetNullInt32(i)
	if nullInt.Valid != false {
		t.Error("expected false got true")
	}
}

func TestSetInt32(t *testing.T) {
	nullInt := sql.NullInt32{
		Int32: 23,
		Valid: true,
	}
	i := model.SetInt32(nullInt)
	if *i != 23 {
		t.Error("expected 23")
	}

	nullInt = sql.NullInt32{
		Int32: 0,
		Valid: false,
	}
	i = model.SetInt32(nullInt)
	if i != nil {
		t.Error("expected nil")
	}
}

func TestSetNullInt64(t *testing.T) {
	i := new(int64)
	*i = 23
	nullInt := model.SetNullInt64(i)
	if nullInt.Valid != true {
		t.Error("expected true got false")
	}

	if nullInt.Int64 != 23 {
		t.Error("expected 23")
	}

	i = nil
	nullInt = model.SetNullInt64(i)
	if nullInt.Valid != false {
		t.Error("expected false got true")
	}
}

func TestSetInt64(t *testing.T) {
	nullInt := sql.NullInt64{
		Int64: 23,
		Valid: true,
	}
	i := model.SetInt64(nullInt)
	if *i != 23 {
		t.Error("expected 23")
	}

	nullInt = sql.NullInt64{
		Int64: 0,
		Valid: false,
	}
	i = model.SetInt64(nullInt)
	if i != nil {
		t.Error("expected nil")
	}
}

func TestSetNullFloat64(t *testing.T) {
	i := new(float64)
	*i = 23.45
	nullInt := model.SetNullFloat64(i)
	if nullInt.Valid != true {
		t.Error("expected true got false")
	}

	if nullInt.Float64 != 23.45 {
		t.Error("expected 23.45")
	}

	i = nil
	nullInt = model.SetNullFloat64(i)
	if nullInt.Valid != false {
		t.Error("expected false got true")
	}
}

func TestSetFloat64(t *testing.T) {
	nullInt := sql.NullFloat64{
		Float64: 23.45,
		Valid:   true,
	}
	i := model.SetFloat64(nullInt)
	if *i != 23.45 {
		t.Error("expected 23")
	}

	nullInt = sql.NullFloat64{
		Float64: 0,
		Valid:   false,
	}
	i = model.SetFloat64(nullInt)
	if i != nil {
		t.Error("expected nil")
	}
}

func TestSetNullTime(t *testing.T) {
	ti := &time.Time{}
	nullTime := model.SetNullTime(ti)
	if nullTime.Valid != true {
		t.Error("expected true got false")
	}

	if nullTime.Time != *ti {
		t.Error("unexpected value")
	}

	ti = nil
	nullTime = model.SetNullTime(ti)
	if nullTime.Valid != false {
		t.Error("expected false got true")
	}
}

func TestTime(t *testing.T) {
	now := time.Now()
	nullTime := sql.NullTime{
		Time:  now,
		Valid: true,
	}
	ti := model.SetTime(nullTime)
	if *ti != now {
		t.Errorf("expected %s", ti)
	}

	nullTime = sql.NullTime{
		Time:  time.Time{},
		Valid: false,
	}
	ti = model.SetTime(nullTime)
	if ti != nil {
		t.Error("expected nil")
	}
}
