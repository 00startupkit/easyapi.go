package core

import (
	"fmt"
	"testing"
)


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


func TestDataProviderTestableProvider (t *testing.T) {
	SetupDataProviderTests(
		t,
		func (t *testing.T, ctx *interface{}) error { return nil },
		func (t *testing.T, ctx *interface{}) error { return nil },
		func(
			t *testing.T,
			schema []*TestSchemaDefinition,
			payload []map[string]interface{},
			opaq *interface{})*DataProvider {
				return CreateTestableUserProvider(payload);
	});
}
