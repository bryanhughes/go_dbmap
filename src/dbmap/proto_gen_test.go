package dbmap

import (
	"testing"
)



func TestRemoveCompositeColumns(t *testing.T) {
	fcols := []ForeignColumns{
		{
			ForeignColumn:   "column_a",
			LocalColumn:     "column_a",
			OrdinalPosition: 1,
		},
		{
			ForeignColumn:   "column_b",
			LocalColumn:     "column_b1",
			OrdinalPosition: 2,
		},
	}

	cols := []Column{
		{
			TableName:       "",
			TableSchema:     "",
			ColumnName:      "column_1",
			OrdinalPosition: 1,
			DataType:        "",
			UdtName:         "",
			ColumnDefault:   "",
			IsNullable:      false,
			IsSequence:      false,
			IsPrimaryKey:    false,
		},
		{
			TableName:       "",
			TableSchema:     "",
			ColumnName:      "column_a",
			OrdinalPosition: 1,
			DataType:        "",
			UdtName:         "",
			ColumnDefault:   "",
			IsNullable:      false,
			IsSequence:      false,
			IsPrimaryKey:    false,
		},
		{
			TableName:       "",
			TableSchema:     "",
			ColumnName:      "column_b1",
			OrdinalPosition: 1,
			DataType:        "",
			UdtName:         "",
			ColumnDefault:   "",
			IsNullable:      false,
			IsSequence:      false,
			IsPrimaryKey:    false,
		},
		{
			TableName:       "",
			TableSchema:     "",
			ColumnName:      "column_c",
			OrdinalPosition: 1,
			DataType:        "",
			UdtName:         "",
			ColumnDefault:   "",
			IsNullable:      false,
			IsSequence:      false,
			IsPrimaryKey:    false,
		},
	}

	// Keep the first column in a composite but remove rest
	removeCompositeColumns(fcols, &cols)
	if len(cols) != 3 {
		t.Fatal("Expected only 3 columns")
	}
}
