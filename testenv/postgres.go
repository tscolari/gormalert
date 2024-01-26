package testenv

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitPgDB(t *testing.T) *gorm.DB {
	connString := "host=127.0.0.1 port=5432 sslmode=disable user=postgres password=postgres dbname=postgres"
	db, err := gorm.Open(postgres.Open(connString))
	require.NoError(t, err, "failed to connect do database")

	_ = db.Exec("TRUNCATE TABLE fruits")
	_ = db.Exec("TRUNCATE TABLE vegetables")

	require.NoError(t, db.Exec(migration).Error)

	return db
}
