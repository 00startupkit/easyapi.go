package main

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

func CreateTestableUserProvider (payload []map[string]interface{}) *DataProvider {
	return &DataProvider{
		All: func (offset int, count int) ([]map[string]interface{}, error) {
			end := min(offset+count, len(payload))
			return payload[offset:end], nil
		},
	}
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
	assert.Equal(t, 1, len(res.Routes))
	assert.Equal(t, res.Routes[0].Route, "/api/users/all")
	assert.Equal(t, res.Routes[0].Type, RequestType_GET)
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
	assert.Equal(t, 1, len(res.Routes))
	assert.Equal(t, res.Routes[0].Route, "/home/users/all")
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
	assert.Equal(t, 1, len(res.Routes))
	
	all_route := res.Routes[0]
	assert.Equal(t, all_route.Route, "/api/users/all")
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
	assert.Equal(t, 1, len(res.Routes))
	
	all_route := res.Routes[0]
	assert.Equal(t, all_route.Route, "/api/users/all")
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
	assert.Equal(t, 1, len(res.Routes))
	
	all_route := res.Routes[0]
	assert.Equal(t, all_route.Route, "/api/users/all")
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
	assert.Equal(t, 1, len(res.Routes))
	
	all_route := res.Routes[0]
	assert.Equal(t, all_route.Route, "/api/users/all")
	res_opaque, err := all_route.Action("offset=-1&&count=-1" )
	assert.Error(t, err)
	assert.Nil(t, res_opaque)
}