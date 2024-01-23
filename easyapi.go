package main

import (
	"fmt"
	"math"
	"net/url"
	"path"
	"strconv"
	"strings"
)

type DataProvider struct {
	// Return all entries from the data store.
	// If a non-negative `offset` is provided, only entries from `offset`
	// into the data payload and `count` entries are returned.
	All func(offset int, count int) ([]map[string]interface{}, error)
}

type Schema struct {
	Name string
	// The data provider associated with this schema.
	Provider *DataProvider
}

type Config struct {
	// The schemas to initialize REST API endpoints for.
	Schemas []*Schema
	// The root of the api. If set, it will be prepended to the api route.
	// Default: "api"
	Root string
}

type RequestType int
const (
	RequestType_UNDEF RequestType = 0
	RequestType_GET RequestType = 1
	RequestType_POST RequestType = 2
)

type RouteResult struct {
	// The route that should be registered.
	Route string
	// The type that the route should be registered as.
	Type RequestType
	// The route action to be taken that will execute the proper action.
	Action func (route_param string) (interface{}, error)
}


type UrlParams struct {
	params url.Values
}
func CreateUrlParams (url_values url.Values) *UrlParams {
	return &UrlParams{
		params: url_values,
	}
}

// Get the param based on the key, if it exists.
// Otherwise, return null.
func (u *UrlParams) Get(key string) (string, error) {
	value := u.params.Get(key)
	if len(value) == 0 {
		return "", fmt.Errorf(fmt.Sprintf("no url param entry found for key \"%s\"", key))
	}
	return value, nil
}


type RequestDefinition struct {
	name string
	method RequestType
	action func(route_params *UrlParams, provider *DataProvider) (interface{}, error)
}

var _requestDefinitions = []RequestDefinition {
	{
		name: "all",
		method: RequestType_GET,
		action: func (route_params *UrlParams, provider *DataProvider) (interface{}, error) {
			var offset int = 0
			var ct int = math.MaxInt32

			offset_str, err := route_params.Get("offset")
			if err == nil && len(offset_str) > 0 {
				offset, err = strconv.Atoi(offset_str)
				if err != nil { return nil, err }
			}

			ct_str, err := route_params.Get("count")
			if err == nil && len(ct_str) > 0 {
				ct, err = strconv.Atoi(ct_str)
				if err != nil { return nil, err }
			}

			if offset < 0 || ct < 0 {
				return nil, fmt.Errorf(fmt.Sprintf("negative offset or count not allowed, offset = %d, count = %d", offset, ct))
			}

			payload, err := provider.All(offset, ct)
			if err != nil {
				return nil, err
			}

			return &payload, nil
		},
	},
}

// Parse url parameters from the given string, `params`.
// e.g. "key=value&&enable=true" will return { "key": "value", "enable": "true" }
func parse_route_params (params string) (*UrlParams, error) {
	query, err := url.ParseQuery(params)
	if err != nil { return nil, err }
	return CreateUrlParams(query), nil

}

type Result struct {
	Routes []*RouteResult
}
func EasyApiImpl (config *Config) (*Result, error) {
	results := &Result{}

	var root string
	if len(config.Root) == 0 {
		root = "/api"
	} else {
		if config.Root[0] != '/' {
			return nil, fmt.Errorf("root must begin with a forward slash (e.g \"/api\")")
		}
		root = config.Root
	}

	for _, schema := range config.Schemas {
		var schema_name string = strings.ToLower(schema.Name)
		if len(schema_name) == 0 {
			return nil, fmt.Errorf("schema name cannot be empty")
		}
		for _, definition := range _requestDefinitions {
			
			var route_result RouteResult
			route_result.Route = path.Join(root, schema_name, definition.name)
			route_result.Type = definition.method
			route_result.Action = func (route_params string) (interface{}, error) {
				parsed_params, err := parse_route_params(route_params)
				if err != nil { return nil, err }

				return definition.action(parsed_params, schema.Provider)
			}

			results.Routes = append(results.Routes, &route_result)
		}
	}
	return results, nil
}
