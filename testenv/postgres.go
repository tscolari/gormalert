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

	require.NoError(t, db.Exec(pgMigration).Error)

	return db
}

const pgMigration = `
CREATE TABLE IF NOT EXISTS fruits (
    id          SERIAL      PRIMARY KEY,
    name        VARCHAR,
    created_at  TIMESTAMP,
    updated_at  TIMESTAMP
);

CREATE TABLE IF NOT EXISTS vegetables (
    id          SERIAL     PRIMARY KEY,
    name        VARCHAR,
    created_at  TIMESTAMP,
    updated_at  TIMESTAMP
);

CREATE INDEX IF NOT EXISTS vegetables_names ON vegetables (name);

INSERT INTO
    fruits (name)
VALUES
    ('apple'),
    ('orange'),
    ('pineapple'),
    ('melon'),
    ('grape'),
    ('lemon'),
    ('blueberry');

INSERT INTO
    vegetables (name)
VALUES
    ('lettuce'),
    ('carrot'),
    ('tomato'),
    ('potato'),
    ('onion'),
    ('parsnip'),
    ('spinach');
`
