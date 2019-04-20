package bdc

import "fmt"

type filter struct {
	field string
	op    string
	value interface{}
}

type sortOption struct {
	field string
	asc   int
}

// Parameters includes filter and/or sort parameters to include in an API query
type Parameters struct {
	filters []filter
	sorts   []sortOption
}

// NewParameters returns a pointer to a fresh Parameters object
func NewParameters() *Parameters {
	return new(Parameters)
}

// AddFilter to a Parameters object
// Allowable operators: = < > != <= >= in nin
// Not all fields can be filtered. Reference documentation at https://developer.bill.com/hc/en-us/articles/210136993-List
func (p *Parameters) AddFilter(field string, operator string, value interface{}) {
	p.filters = append(p.filters, filter{field, operator, value})
}

// AddSort to a Parameters object
// descending (asc = 0);
// ascending (asc = 1)
func (p *Parameters) AddSort(field string, asc int) {
	p.sorts = append(p.sorts, sortOption{field, asc})
}

func encodeParameters(params []*Parameters) (filters, sort []map[string]interface{}) {
	switch p := len(params); {
	case p > 1:
		fmt.Println("Can supply at most one parameter object")
		return nil, nil
	case p == 1:
		var filters, sorts []map[string]interface{}
		params := params[0]
		for _, filter := range params.filters {
			filters = append(filters, map[string]interface{}{"field": filter.field, "op": filter.op, "value": filter.value})
		}
		for _, sort := range params.sorts {
			sorts = append(sorts, map[string]interface{}{"field": sort.field, "asc": sort.asc})
		}
		return filters, sorts
	default:
		return nil, nil
	}
}
