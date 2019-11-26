package dbmap

type TableConstraint struct {
	constraintSchema  string
	constraintName    string
	tableSchema       string
	tableName         string
	constraintType    string
	isDeferrable      bool
	initiallyDeferred bool
}

type Column struct {
	tableName             string
	tableSchema           string
	columnName            string
	ordinalPosition       int
	columnDefault         string
	isNullable            bool
	charMaxLength         int
	charOctetLength       int
	numericPrecision      int
	numericPrecisionRadix int
	numericScale          int
	datetimePrecision     int
	domainSchema          string
	domainName            string
	udtSchema             string
	udtName               string
	isUpdatable           bool
}

type Table struct {
	tableName    string
	tableSchema  string
	tableType    string
	isInsertable bool
}

type Schema struct {
	schemaName string
	tables     []Table
}
