package drivers

import (
	"database/sql"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/ddosify/go-faker/faker"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
)

var (
	_DB_USER = "root"
	_DB_PASS = "password"
)


func MysqlConnectionString (dbname string) string {
	return fmt.Sprintf("%s:%s@/%s", _DB_USER, _DB_PASS, dbname)
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
	db, err := sql.Open("mysql", MysqlConnectionString(""))
	if err != nil { return err }
	defer db.Close()

	create_database_sql := fmt.Sprintf(`CREATE DATABASE %s`, database_name)
	_, err = db.Exec(create_database_sql)
	return err
}

// Remove the database with the given `database_name`
func cleanup_database (database_name string) error {
	db, err := sql.Open("mysql", MysqlConnectionString(""))
	if err != nil { return err }
	defer db.Close()

	create_database_sql := fmt.Sprintf(`DROP DATABASE %s`, database_name)
	_, err = db.Exec(create_database_sql)
	return err
}


func execute_query (dbname, query string) error {
	db, err := sql.Open("mysql", MysqlConnectionString(dbname))
	if err != nil { return err }
	defer db.Close()

	_, err = db.Exec(query)
	return err
}

func GenerateTestUserPayload (ct int) []map[string]interface{} {

	faker := faker.NewFaker()

	var payload []map[string]interface{}
	for i := 0; i < ct; i++ {
		var entry map[string]interface{}
		entry = make(map[string]interface{})
		for {
			entry["name"] = faker.RandomPersonFullName()
			if !strings.Contains(entry["name"].(string), "\"") { break }
		}
		for {
			entry["location"] = faker.RandomAddressCity()
			if !strings.Contains(entry["location"].(string), "\"") { break }
		}
		payload = append(payload, entry)
	}
	return payload
}


type ColType int
const (
	ColType_UNDEF ColType = 0
	ColType_INT ColType = 1
	ColType_STRING ColType = 2
)

type Column struct {
	Type ColType
	Name string
}

func bulk_insert_query_creator (table string, columns []Column, payload []map[string]interface{}) (string, error) {
	if len(columns) == 0 {
		return "", fmt.Errorf("must be at least 1 entry in the columns")
	}

	column_names := []string{}
	for _, c:= range columns {
		column_names = append(column_names, c.Name)
	}

	values := []string{}
	for _, entry := range payload {
		curr_values := []string{}
		for _, column := range columns {
			value, exists := entry[column.Name]
			if !exists {
				return "", fmt.Errorf(fmt.Sprintf("entry in payload does not have value for field: \"%s\"", column.Name))
			}

			string_value, ok := value.(string)
			if !ok {
				return "", fmt.Errorf(fmt.Sprintf("expected entry for column \"%s\" to be a string", column.Name))
			}

			switch column.Type {
				case ColType_INT: // Nothing needs to be done.
					// string_value = string_value
					break
				case ColType_STRING:
					string_value = fmt.Sprintf("\"%s\"", string_value)
					break
				default:
					return "", fmt.Errorf("Undefined column type")
			}

			curr_values = append(curr_values, string_value)
		}
		values = append(values, fmt.Sprintf("(%s)", strings.Join(curr_values, ",")))
	}

	return fmt.Sprintf(`INSERT INTO %s (%s)
		VALUES %s
	`, table, strings.Join(column_names, ","), strings.Join(values, ",\n\t")), nil
	
}

//////////////////////////////////////////////////////////////////////////

func TestMysqlSetup (t *testing.T) {
	t.Run("database setup and teardown", func (t *testing.T) {
		var dbname string = hash("example_database")
		assert.NoError(t, setup_database(dbname))
		defer func () { assert.NoError(t, cleanup_database(dbname)) }()
	});
	t.Run("table creation", func (t *testing.T) {
		var dbname string = hash("example_database")
		assert.NoError(t, setup_database(dbname))
		defer func () { assert.NoError(t, cleanup_database(dbname)) }()

		table_sql := `CREATE TABLE example_table (
			id int,
			latname varchar(255),
			firstname varchar(255)
		)`
		assert.NoError(t, execute_query(dbname, table_sql))
	});
	t.Run("bulk insertion", func (t *testing.T) {
		var dbname string = hash("example_database")
		assert.NoError(t, setup_database(dbname))
		defer func () { assert.NoError(t, cleanup_database(dbname)) }()

		table_sql := `CREATE TABLE Users (
			name varchar(255),
			location varchar(255)
		)`
		assert.NoError(t, execute_query(dbname, table_sql))

		columns := []Column{
			{Name: "name", Type: ColType_STRING },
			{Name: "location", Type: ColType_STRING },
		}
		payload := GenerateTestUserPayload(3)

		insert_query, error := bulk_insert_query_creator("Users", columns, payload)
		assert.NoError(t, error, fmt.Sprintf("Insert query creation failed: %s", insert_query))
		assert.NoError(t, execute_query(dbname, insert_query), fmt.Sprintf("Insert query failed: %s", insert_query))
	});
	t.Run("bulk insertion large", func (t *testing.T) {
		var dbname string = hash("example_database")
		assert.NoError(t, setup_database(dbname))
		defer func () { assert.NoError(t, cleanup_database(dbname)) }()

		table_sql := `CREATE TABLE Users (
			name varchar(255),
			location varchar(255)
		)`
		assert.NoError(t, execute_query(dbname, table_sql))

		columns := []Column{
			{Name: "name", Type: ColType_STRING },
			{Name: "location", Type: ColType_STRING },
		}
		payload := GenerateTestUserPayload(100)

		insert_query, error := bulk_insert_query_creator("Users", columns, payload)
		assert.NoError(t, error, fmt.Sprintf("Insert query creation failed: %s", insert_query))
		assert.NoError(t, execute_query(dbname, insert_query), fmt.Sprintf("Insert query failed: %s", insert_query))
	});
}
