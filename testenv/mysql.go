package testenv

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitMysqlDB(t *testing.T) *gorm.DB {
	connString := "root:root@tcp(127.0.0.1:3306)/test"
	db, err := gorm.Open(mysql.Open(connString))
	require.NoError(t, err, "failed to connect do database")

	_ = db.Exec("TRUNCATE TABLE fruits")
	_ = db.Exec("TRUNCATE TABLE vegetables")

	for _, migration := range mysqlMigration {
		require.NoError(t, db.Exec(migration).Error)
	}

	return db
}

var mysqlMigration = []string{
	`CREATE TABLE IF NOT EXISTS fruits (
		id          INT AUTO_INCREMENT,
		name        VARCHAR(255),
		created_at  TIMESTAMP,
		updated_at  TIMESTAMP,
		PRIMARY KEY (id)
	);`,

	` CREATE TABLE IF NOT EXISTS vegetables (
		id          INT	AUTO_INCREMENT,
		name        VARCHAR(255),
		created_at  TIMESTAMP,
		updated_at  TIMESTAMP,
		PRIMARY KEY (id),
		INDEX names(name)
	);`,

	`INSERT INTO
		fruits (name)
	VALUES
		('apple'),
		('orange'),
		('pineapple'),
		('melon'),
		('grape'),
		('lemon'),
		('blueberry');`,

	`INSERT INTO
		vegetables (name)
	VALUES
		('lettuce'),
		('carrot'),
		('tomato'),
		('potato'),
		('onion'),
		('parsnip'),
		('spinach');
	`,
}
