package core

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ddosify/go-faker/faker"
	"github.com/stretchr/testify/assert"
)

func UserSchemaDefinition () []*TestSchemaDefinition {
	return []*TestSchemaDefinition{
		{ FieldName: "name", FieldType: TestSchemaFieldType_STRING },
		{ FieldName: "location", FieldType: TestSchemaFieldType_STRING },
	}
}
func GenerateTestUserPayload (ct int) ([]map[string]interface{}, []*TestSchemaDefinition) {
	faker := faker.NewFaker()

	var payload []map[string]interface{}
	for i := 0; i < ct; i++ {
		var entry map[string]interface{}
		entry = make(map[string]interface{})

		var name string
		var location string
		for {
			name = faker.RandomPersonFullName()
			if !strings.Contains(name, "\"") { break }
		}
		for {
			location = faker.RandomAddressCity()
			if !strings.Contains(location, "\"") { break }
		}
		entry["name"] = name
		entry["location"] = location
		payload = append(payload, entry)
	}

	return payload, UserSchemaDefinition()
}


func GetRoute (res *Result, route_path string) *RouteResult {
	if res == nil { return nil }
	for _, route := range res.Routes {
		if route.Route == route_path { return route }
	}
	return nil
}

type TestSchemaFieldType int
const (
	TestSchemaFieldType_UNDEF TestSchemaFieldType = 0
	TestSchemaFieldType_STRING TestSchemaFieldType = 1
	TestSchemaFieldType_INT TestSchemaFieldType = 2
)

type TestSchemaDefinition struct {
	FieldName string
	FieldType TestSchemaFieldType
}

func SetupDataProviderTests (
	t *testing.T,
	setup_fn func(t *testing.T, opaq *interface{}) error,
	teardown_fn func(t *testing.T, opaq *interface{}) error,
	data_provider_creator func(
		t *testing.T,
		schema []*TestSchemaDefinition,
		payload []map[string]interface{},
		opaq *interface{},
	)*DataProvider,
) {

	t.Run("fetch all", func (t *testing.T) {
		var ctx interface{}
		assert.NoError(t, setup_fn(t, &ctx))
		defer func () {
			assert.NoError(t, teardown_fn(t, &ctx))
		}()

		payload, schema := GenerateTestUserPayload(10)
		test_user_provider := data_provider_creator(t, schema, payload, &ctx)
		assert.NotNil(t, test_user_provider, "Failed to create data provider")
		if test_user_provider == nil {  return }

		config := &Config{
			Schemas: []*Schema{
				{ 
					Name: "Users",
					Provider: test_user_provider,
				},
			},
		}

		res, err := EasyApiImpl(config);
		assert.NoError(t, err, "Default config failed.");
		assert.GreaterOrEqual(t, len(res.Routes), 1)

		all_route := GetRoute(res, "/api/users/all")
		assert.NotNil(t, all_route)
		assert.Equal(t, all_route.Type, RequestType_GET)

	});

	t.Run("setup api root", func (t *testing.T) {
		var ctx interface{}
		assert.NoError(t, setup_fn(t, &ctx))
		defer func () {
			assert.NoError(t, teardown_fn(t, &ctx))
		}()

		payload, schema := GenerateTestUserPayload(10)
		test_user_provider := data_provider_creator(t, schema, payload, &ctx)
		assert.NotNil(t, test_user_provider, "Failed to create data provider")
		if test_user_provider == nil {  return }

		config := &Config{
			Root: "/home",
			Schemas: []*Schema{
				{ 
					Name: "Users",
					Provider: test_user_provider,
				},
			},
		}
		res, err := EasyApiImpl(config);
		assert.NoError(t, err, "Default config failed.");
		assert.NotEmpty(t, res.Routes)
		all_route := GetRoute(res, "/home/users/all")
		assert.NotNil(t, all_route)
	});

	t.Run("invalid api root", func (t *testing.T) {
		var ctx interface{}
		assert.NoError(t, setup_fn(t, &ctx))
		defer func () {
			assert.NoError(t, teardown_fn(t, &ctx))
		}()

		payload, schema := GenerateTestUserPayload(10)
		test_user_provider := data_provider_creator(t, schema, payload, &ctx)
		assert.NotNil(t, test_user_provider, "Failed to create data provider")
		if test_user_provider == nil {  return }
		config := &Config{
			Root: "home",
			Schemas: []*Schema{
				{ 
					Name: "Users",
					Provider: test_user_provider,
				},
			},
		}
		_, err := EasyApiImpl(config);
		assert.Error(t, err, "Default config failed.");
		assert.ErrorContains(t, err, "root must begin with a forward slash")
	});

	t.Run("fetch all", func (t *testing.T) {
		var ctx interface{}
		assert.NoError(t, setup_fn(t, &ctx))
		defer func () {
			assert.NoError(t, teardown_fn(t, &ctx))
		}()

		var kDataSize int = 10
		payload, schema := GenerateTestUserPayload(kDataSize)
		test_user_provider := data_provider_creator(t, schema, payload, &ctx)
		assert.NotNil(t, test_user_provider, "Failed to create data provider")
		if test_user_provider == nil {  return }

		config := &Config{
			Schemas: []*Schema{
				{ 
					Name: "Users",
					Provider: test_user_provider,
				},
			},
		}

		res, err := EasyApiImpl(config);
		assert.NoError(t, err, "Default config failed.")
		assert.GreaterOrEqual(t, len(res.Routes), 1)

		all_route := GetRoute(res, "/api/users/all")
		assert.NotNil(t, all_route)

		res_opaque, err := all_route.Action("")
		assert.NoError(t, err)

		data, ok := res_opaque.(*[]map[string]interface{})
		assert.True(t, ok)
		assert.NotNil(t, data)
		assert.Equal(t, len(*data), kDataSize)
	});

	t.Run("fetch all bounded", func (t *testing.T) {
		var ctx interface{}
		assert.NoError(t, setup_fn(t, &ctx))
		defer func () {
			assert.NoError(t, teardown_fn(t, &ctx))
		}()

		var kDataSize int = 10
		payload, schema := GenerateTestUserPayload(kDataSize)
		test_user_provider := data_provider_creator(t, schema, payload, &ctx)
		assert.NotNil(t, test_user_provider, "Failed to create data provider")
		if test_user_provider == nil {  return }

		config := &Config{
			Schemas: []*Schema{
				{ 
					Name: "Users",
					Provider: test_user_provider,
				},
			},
		}

		res, err := EasyApiImpl(config);
		assert.NoError(t, err, "Default config failed.")
		assert.NotEmpty(t, res.Routes)
		
		all_route := GetRoute(res, "/api/users/all")
		assert.NotNil(t, all_route)
		var kNbEntriesToRetrieve = 4
		res_opaque, err := all_route.Action(fmt.Sprintf("offset=2&&count=%d", kNbEntriesToRetrieve))
		assert.NoError(t, err)

		data, ok := res_opaque.(*[]map[string]interface{})
		assert.True(t, ok)
		assert.NotNil(t, data)
		assert.Equal(t, kNbEntriesToRetrieve, len(*data))
	});

	t.Run("fetch all bounded 2", func (t *testing.T) {
		var ctx interface{}
		assert.NoError(t, setup_fn(t, &ctx))
		defer func () {
			assert.NoError(t, teardown_fn(t, &ctx))
		}()

		var kDataSize int = 10
		payload, schema := GenerateTestUserPayload(kDataSize)
		test_user_provider := data_provider_creator(t, schema, payload, &ctx)
		assert.NotNil(t, test_user_provider, "Failed to create data provider")
		if test_user_provider == nil {  return }

		config := &Config{
			Schemas: []*Schema{
				{ 
					Name: "Users",
					Provider: test_user_provider,
				},
			},
		}

		res, err := EasyApiImpl(config);
		assert.NoError(t, err, "Default config failed.")
		assert.NotEmpty(t, res.Routes)

		all_route := GetRoute(res, "/api/users/all")
		assert.NotNil(t, all_route)

		var kNbEntriesToRetrieve = 3
		res_opaque, err := all_route.Action(fmt.Sprintf("offset=8&&count=%d", kNbEntriesToRetrieve))
		assert.NoError(t, err)

		data, ok := res_opaque.(*[]map[string]interface{})
		assert.True(t, ok)
		assert.NotNil(t, data)
		assert.Equal(t, kNbEntriesToRetrieve - 1, len(*data))
	});

	t.Run("fetch all bounded negative fails", func (t *testing.T) {
		var ctx interface{}
		assert.NoError(t, setup_fn(t, &ctx))
		defer func () {
			assert.NoError(t, teardown_fn(t, &ctx))
		}()

		var kDataSize int = 10
		payload, schema := GenerateTestUserPayload(kDataSize)
		test_user_provider := data_provider_creator(t, schema, payload, &ctx)
		assert.NotNil(t, test_user_provider, "Failed to create data provider")
		if test_user_provider == nil {  return }

		config := &Config{
			Schemas: []*Schema{
				{ 
					Name: "Users",
					Provider: test_user_provider,
				},
			},
		}

		res, err := EasyApiImpl(config);
		assert.NoError(t, err, "Default config failed.")
		assert.NotEmpty(t, res.Routes)

		all_route := GetRoute(res, "/api/users/all")
		assert.NotNil(t, all_route)

		res_opaque, err := all_route.Action("offset=-1&&count=-1" )
		assert.Error(t, err)
		assert.Nil(t, res_opaque)
	});

	t.Run("find one result", func (t *testing.T) {
		var ctx interface{}
		assert.NoError(t, setup_fn(t, &ctx))
		defer func () {
			assert.NoError(t, teardown_fn(t, &ctx))
		}()

		payload := []map[string]interface{}{
			{
				"name": "John",
				"location": "Arizona",
			}, {
				"name": "Jimmy",
				"location": "California",
			},{
				"name": "Alex",
				"location": "Texas",
			}, 
		}
		schema := UserSchemaDefinition()

		test_user_provider := data_provider_creator(t, schema, payload, &ctx)
		assert.NotNil(t, test_user_provider, "Failed to create data provider")
		if test_user_provider == nil {  return }

		config := &Config{
			Schemas: []*Schema{
				{ 
					Name: "Users",
					Provider: test_user_provider,
				},
			},
		}

		res, err := EasyApiImpl(config);
		assert.NoError(t, err, "Default config failed.")
		assert.NotEmpty(t, res.Routes)

		all_route := GetRoute(res, "/api/users/findone")
		assert.NotNil(t, all_route)

		res_opaque, err := all_route.Action("name=\"-eq John\"&&location=\"-eq Arizona\"")
		assert.NoError(t, err)

		data, ok := res_opaque.(*map[string]interface{})
		assert.True(t, ok)
		assert.NotNil(t, data)
		assert.Equal(t, (*data)["name"], "John")
		assert.Equal(t, (*data)["location"], "Arizona")
	});
}
