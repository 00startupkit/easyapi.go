package drivers

import (
	"database/sql"
	"fmt"
	"math/rand"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
)

var (
	_DB_USER = "root"
	_DB_PASS = "password"
)

func MysqlConnectionString () string {
	return fmt.Sprintf("%s:%s@/", _DB_USER, _DB_PASS)
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seed := rand.NewSource(time.Now().UnixNano())
	random := rand.New(seed)

	result := make([]byte, length)
	for i := range result {
			result[i] = charset[random.Intn(len(charset))]
	}
	return string(result)
}

func hash (dbname string) string {
	return fmt.Sprintf("%s_%s", dbname, generateRandomString(6))
}

// Create a database with the given `database_name`
func setup_database (database_name string) error {
	db, err := sql.Open("mysql", MysqlConnectionString())
	if err != nil { return err }
	defer db.Close()

	create_database_sql := fmt.Sprintf(`CREATE DATABASE %s`, database_name)
	_, err = db.Exec(create_database_sql)
	return err
}

// Remove the database with the given `database_name`
func cleanup_database (database_name string) error {
	db, err := sql.Open("mysql", MysqlConnectionString())
	if err != nil { return err }
	defer db.Close()

	create_database_sql := fmt.Sprintf(`DROP DATABASE %s`, database_name)
	_, err = db.Exec(create_database_sql)
	return err
}

func TestMysqlSetup (t *testing.T) {
	t.Run("database setup and teardown", func (t *testing.T) {
		var dbname string = hash("example_database")
		assert.NoError(t, setup_database(dbname))
		defer func () { assert.NoError(t, cleanup_database(dbname)) }()
	});
}
