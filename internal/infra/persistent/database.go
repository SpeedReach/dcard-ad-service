package persistent

import "database/sql"

type Database struct {
	inner *sql.DB
}

func NewSQLDatabase(inner *sql.DB) Database {
	return Database{inner: inner}
}
