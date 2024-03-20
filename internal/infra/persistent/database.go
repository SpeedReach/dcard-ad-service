package persistent

import "database/sql"

type database struct {
	inner *sql.DB
}

func NewSQLDatabase(inner *sql.DB) Storage {
	return database{inner: inner}
}
