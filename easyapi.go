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
	FindOne func(constraints []Constraint) (*map[string]interface{}, error)
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

type RequestDefinition struct {
	name string
	method RequestType
	action func(route_params *UrlParams, provider *DataProvider) (interface{}, error)
}

type RouteResult struct {
	// The route that should be registered.
	Route string
	// The type that the route should be registered as.
	Type RequestType

	_definition RequestDefinition
	_schema *Schema
}

func (r *RouteResult) Action (route_params string) (interface{}, error) {
	parsed_params, err := parse_route_params(route_params)
	if err != nil { return nil, err }
	return r._definition.action(parsed_params, r._schema.Provider)
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

func (u *UrlParams) GetInt(key string) (int, error) {
	str_value, err := u.Get(key)
	if err != nil { return -1, err }
	int_value, err := strconv.Atoi(str_value)
	if err != nil { return -1, err }
	return int_value, nil
}

type Comparison string
const (
	Comparison_UNDEF Comparison = ""
	Comparison_EQ Comparison = "eq"
	Comparison_NE Comparison = "ne"
	Comparison_LT Comparison = "lt"
	Comparison_LE Comparison = "le"
	Comparison_GT Comparison = "gt"
	Comparison_GE Comparison = "ge"
)

type Constraint struct {
	Property string
	Value string
	Comparison Comparison
}

func filter_string_array(lst []string, filter_fn func(string) bool) []string {
	p := []string{}
	for _, el := range lst {
		if filter_fn(el) {
			p = append(p, el)
		}
	}
	return p
}

// Parse a comparison part, e.g. "eq "Sam"" should return
// (Comparison_EQ, "Sam", nil)
func parse_comparison_part (part string) (Comparison, string, error) {
	part = strings.TrimSpace(part)
	last := len(part) - 1
	if len(part) >= 2 {
		if (string(part[0]) == "\"" && string(part[last]) == "\"") || (string(part[0]) == "'" && string(part[last]) == "'") {
			part = part[1:last]
		}
	}

	parts := strings.Split(part, " ")
	parts = filter_string_array(parts, func (el string) bool { return len(el) > 0 })

	if len(parts) != 2 {
		return Comparison_UNDEF, "", fmt.Errorf(fmt.Sprintf("expected 2 parts in comparison string, received %d. comparison string=\"%#v\"", len(parts), parts))
	}

	switch parts[0] {
	case "-eq":
		return Comparison_EQ, parts[1], nil
	case "-ne":
		return Comparison_NE, parts[1], nil
	case "-le":
		return Comparison_LE, parts[1], nil
	case "-lt":
		return Comparison_LT, parts[1], nil
	case "-gt":
		return Comparison_GT, parts[1], nil
	case "-ge":
		return Comparison_GE, parts[1], nil
	}
	return Comparison_UNDEF, "", fmt.Errorf("unknown comparison operator: \"%s\"", parts[0])
}

func parse_constraints(route_params *UrlParams) ([]Constraint, error) {

	var constraints []Constraint

	for key, values := range route_params.params {
		if len(values) != 1 { continue }
		var value string = values[0]
		
		comparison, right_value, err := parse_comparison_part(value)
		if err != nil { return nil, err }
		
		c := Constraint {}
		c.Property = key
		c.Value = right_value
		c.Comparison = comparison

		constraints = append(constraints, c)
	}

	return constraints, nil

}

var _requestDefinitions = []RequestDefinition {
	{
		name: "all",
		method: RequestType_GET,
		action: func (route_params *UrlParams, provider *DataProvider) (interface{}, error) {

			offset, err := route_params.GetInt("offset")
			if err !=  nil { offset = 0 }
			ct, err := route_params.GetInt("count")
			if err != nil { ct = math.MaxInt32 }

			if offset < 0 || ct < 0 {
				return nil, fmt.Errorf(fmt.Sprintf("negative offset or count not allowed, offset = %d, count = %d", offset, ct))
			}

			payload, err := provider.All(offset, ct)
			if err != nil {
				return nil, err
			}

			return &payload, nil
		},
	},{
		name: "findone",
		method: RequestType_GET,
		action: func (route_params *UrlParams, provider *DataProvider) (interface{}, error) {
			constraints, err := parse_constraints(route_params)
			if err != nil { return nil, err }
			return provider.FindOne(constraints)
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
			route_result._definition = definition
			route_result._schema = schema
			results.Routes = append(results.Routes, &route_result)
		}
	}
	return results, nil
}
