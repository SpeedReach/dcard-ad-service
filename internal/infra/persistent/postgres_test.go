//go:build integration

package persistent

import (
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestPostgresInsertAndGet(t *testing.T) {
	pgUri, found := os.LookupEnv("POSTGRES_URI")
	if !found {
		t.Error("POSTGRES_URI not found")
		return
	}
	sqlDb, err := sql.Open("pgx", pgUri)
	_, err = sqlDb.Exec("DELETE FROM Conditions WHERE 1=1")
	if err != nil {
		panic(err)
	}
	_, err = sqlDb.Exec("DELETE FROM Ads WHERE 1=1")
	if err != nil {
		panic(err)
	}

	assert.NoError(t, err)
	db := NewSQLDatabase(sqlDb)

	TestStorage(t, db)
}
