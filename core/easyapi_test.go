package core

import (
	"fmt"
	"testing"

	"github.com/ddosify/go-faker/faker"
	"github.com/stretchr/testify/assert"
)

func GenerateTestUserPayload (ct int) []map[string]interface{} {

	faker := faker.NewFaker()

	var payload []map[string]interface{}
	for i := 0; i < ct; i++ {
		var entry map[string]interface{}
		entry = make(map[string]interface{})
		entry["name"] = faker.RandomPersonFullName()
		entry["loation"] = faker.RandomAddressCity()
		payload = append(payload, entry)
	}
	return payload
}

func min(a int, b int) int {
	if a < b { return a }
	return b
}

func matches_constraint (entry map[string]interface{}, constraint Constraint) bool {
	val, ok := entry[constraint.Property]
	if !ok { return false }
	switch constraint.Comparison {
	case Comparison_EQ:
		return val == constraint.Value
	case Comparison_NE:
		return val != constraint.Value
	}
	return false
}

func matches_constraints (entry map[string]interface{}, constraints []Constraint) bool {
	for _, constraint := range constraints {
		if !matches_constraint(entry, constraint) {
			return false
		}
	}
	return true
}

func CreateTestableUserProvider (payload []map[string]interface{}) *DataProvider {
	return &DataProvider{
		All: func (offset int, count int) ([]map[string]interface{}, error) {
			end := min(offset+count, len(payload))
			return payload[offset:end], nil
		},
		FindOne: func(constraints []Constraint) (*map[string]interface{}, error) {

			for _, entry := range payload {
				if matches_constraints(entry, constraints) {
					return &entry, nil
				}
			}
			return nil, fmt.Errorf("no matching entry found")
		},
	}
}

func GetRoute (res *Result, route_path string) *RouteResult {
	if res == nil { return nil }
	for _, route := range res.Routes {
		if route.Route == route_path { return route }
	}
	return nil
}

func TestNormalConfig (t *testing.T) {

	payload := GenerateTestUserPayload(10)
	test_user_provider := CreateTestableUserProvider(payload)

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
}

func TestApiRootConfig (t *testing.T) {
	payload := GenerateTestUserPayload(10)
	test_user_provider := CreateTestableUserProvider(payload)

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
}

func TestApiInvalidRootConfig (t *testing.T) {
	payload := GenerateTestUserPayload(10)
	test_user_provider := CreateTestableUserProvider(payload)
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
}

func TestFetchAll (t *testing.T) {

	var kDataSize int = 10
	payload := GenerateTestUserPayload(kDataSize)
	test_user_provider := CreateTestableUserProvider(payload)

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
}

func TestFetchAllBounded (t *testing.T) {

	var kDataSize int = 10
	payload := GenerateTestUserPayload(kDataSize)
	test_user_provider := CreateTestableUserProvider(payload)

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
}

func TestFetchAllBounded2 (t *testing.T) {

	var kDataSize int = 10
	payload := GenerateTestUserPayload(kDataSize)
	test_user_provider := CreateTestableUserProvider(payload)

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
}

func TestFetchAllBoundedNegative (t *testing.T) {

	var kDataSize int = 10
	payload := GenerateTestUserPayload(kDataSize)
	test_user_provider := CreateTestableUserProvider(payload)

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
}

func TestFindOne (t *testing.T) {

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

	test_user_provider := CreateTestableUserProvider(payload)

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
}
