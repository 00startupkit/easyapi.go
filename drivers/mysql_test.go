package drivers

import (
	"database/sql"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/00startupkit/easyapi.go/core"
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

func schema_type_to_sql_type (fieldtype core.TestSchemaFieldType) (string, error) {
	switch fieldtype {
	case core.TestSchemaFieldType_STRING:
		return "varchar(255)", nil
	case core.TestSchemaFieldType_INT:
		return "int", nil
	}
	return "", fmt.Errorf(fmt.Sprintf("conversion from schema field type %d to sql type not defined", fieldtype))
}

func create_table_from_schema (dbname, tablename string, schema []*core.TestSchemaDefinition) error {
	if len(schema) == 0 { return fmt.Errorf("cannot create table from an empty schema") }

	field_parts := []string{}
	for _, s := range schema {
		sqltype, err := schema_type_to_sql_type(s.FieldType)
		if err != nil { return err }
		field_info := fmt.Sprintf(`%s %s`, s.FieldName, sqltype)

		field_parts = append(field_parts, field_info)
	}

	table_sql := fmt.Sprintf(`CREATE TABLE %s (
		%s
	)`, tablename, strings.Join(field_parts, ",\n"))
	return execute_query(dbname, table_sql)
}

func schema_type_to_col_type (schematype core.TestSchemaFieldType) (ColType, error) {
	switch schematype {
	case core.TestSchemaFieldType_INT:
		return ColType_INT, nil
	case core.TestSchemaFieldType_STRING:
		return ColType_STRING, nil
	}
	return ColType_UNDEF, fmt.Errorf(fmt.Sprintf("conversion from schema type to col type not defined for %#v", schematype))
}

type MysqlTestUnitContext struct {
	DatabaseName string
}

func TestMysqlDataProvider (t *testing.T) {
	core.SetupDataProviderTests(
		t,
		func (t *testing.T, ctx *interface{}) error {
			var dbname string = hash("test_database")
			*ctx = &MysqlTestUnitContext{
				DatabaseName: dbname,
			}
			// Database setup
			assert.NoError(t, setup_database(dbname))
			fmt.Printf("Database setup: %s\n", dbname)
			return nil
		},
		func (t *testing.T, ctx *interface{}) error {
			mysql_ctx, ok := (*ctx).(*MysqlTestUnitContext)
			if !ok {  return fmt.Errorf("data provider test context is not defined") }

			assert.NoError(t, cleanup_database(mysql_ctx.DatabaseName))
			fmt.Printf("Database cleaned up: %s\n", mysql_ctx.DatabaseName)
			return nil
		},
		func (
			t *testing.T,
			schema []*core.TestSchemaDefinition,
			payload []map[string]interface{},
			opaq *interface{}) *core.DataProvider {
				fmt.Printf("Fetching mysql data\n")

				mysql_ctx, ok := (*opaq).(*MysqlTestUnitContext)
				assert.True(t, ok, "mysql context not provided")

				var dbname string = mysql_ctx.DatabaseName

				// Table creation based on schema
				var tablename string = "Users"
				assert.NoError(t, create_table_from_schema(dbname, tablename, schema));

				// Insert the data payload into the sql table
				columns := []Column{}
				for _, s := range schema {
					coltype, err := schema_type_to_col_type(s.FieldType)
					assert.NoError(t, err, "Failed to convert schema type to column type")
					if err != nil { return nil }

					col := Column {}
					col.Name = s.FieldName
					col.Type = coltype

					columns = append(columns, col)
				}
				insert_query, error := bulk_insert_query_creator(tablename, columns, payload)
				assert.NoError(t, error)
				assert.NoError(t, execute_query(dbname, insert_query), fmt.Sprintf("Insert query failed: %s", insert_query))

				// Create the data driver for accessing the newly inserted data from mysql
				mysql_dataprovider, err := CreateMysqlDataProvider(_DB_USER, _DB_PASS, dbname, tablename, columns)
				assert.NoError(t, err)
				return mysql_dataprovider
	})
}