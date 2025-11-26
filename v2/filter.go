package filterbuilder

import (
	"reflect"
	"strconv"
	"strings"
)

type Filterer interface {
	Build(data any, ph string, inSeq bool, offset int) (string, any, int, error)
	GetPair() any
}

// Filter - the filter struct
type Filter struct {
	Data           any     `json:"data,omitempty"`
	Eq             []Eq    `json:"eq,omitempty"`               // Equality pairs
	Lt             []Lt    `json:"lt,omitempty"`               // Less than pairs
	Lte            []Lte   `json:"lte,omitempty"`              // Less than equal pairs
	Gt             []Gt    `json:"gt,omitempty"`               // Greater pairs
	Gte            []Gte   `json:"gte,omitempty"`              // Greater than equal pair
	Group          []Group `json:"group,omitempty"`            // Group, a utility of grouping main comparison filters
	Ne             []Ne    `json:"ne,omitempty"`               // Not equality pairs
	Lk             []Lk    `json:"lk,omitempty"`               // Like pairs
	Or             []Or    `json:"or,omitempty"`               // Or pairs. These should be any of the definite filter
	In             []In    `json:"in,omitempty"`               // In column pair.
	NotIn          []Ni    `json:"not_in,omitempty"`           // Not In column pair
	Between        []Bw    `json:"between,omitempty"`          // Between column pair
	Placeholder    string  `json:"placeholder,omitempty"`      // Parameter place holder
	InSequence     bool    `json:"in_sequence,omitempty"`      // Parameter place holders would be numbered in sequence
	Offset         int     `json:"offset,omitempty"`           // Sets the start of parameter number
	AllowNoFilters bool    `json:"allow_no_filters,omitempty"` // Allow no filter upon building
}

func buildPair(f Filterer, srcData any, operator, ph string, inSeq bool, offset int) (string, any, int, error) {
	var (
		qry string
		v   any
		err error
	)

	p := f.GetPair()
	vOp := reflect.ValueOf(p)
	col := vOp.FieldByName("Column").String()
	val := vOp.FieldByName("Value").Interface().(Value)

	v, err = getFilterValue(srcData, val)
	if err != nil {
		return qry, v, offset, err
	}
	if v == nil {
		return qry, v, offset, err
	}
	switch v.(type) {
	case Null:
		qry = col + " IS NULL"
	default:
		offset++
		ph = strings.TrimSpace(ph)
		qry = col + " " + operator + " " + ph
		if inSeq && ph != "?" {
			qry += strconv.Itoa(offset)
		}
	}
	return qry, v, offset, nil
}

func buildMembershipPair(f Filterer, srcData any, operator, ph string, inSeq bool, offset int) (string, []any, int, error) {
	var (
		qry, cma string
		v        any
		args     []any
		err      error
	)

	p := f.GetPair()
	vOp := reflect.ValueOf(p)
	col := vOp.FieldByName("Column").String()
	val := vOp.FieldByName("Value").Interface().([]Value)

	args = make([]any, 0, 10)
	qry = col + " " + operator + " ("
	for _, pr := range val {
		v, err = getFilterValue(srcData, pr)
		if err != nil {
			return qry, args, offset, err
		}
		if v == nil {
			return qry, args, offset, err
		}
		switch v.(type) {
		case Null:
			return qry, args, offset, err
		}
		offset++
		ph = strings.TrimSpace(ph)
		qry += cma + ph
		if inSeq && ph != "?" {
			qry += strconv.Itoa(offset)
		}
		args = append(args, v)
		cma = ","
	}
	qry += ")"

	return qry, args, offset, nil
}

func buildRangePair(f Filterer, srcData any, ph string, inSeq bool, offset int) (string, []any, int, error) {
	var (
		qry, cma string
		v        any
		args     []any
		err      error
	)

	p := f.GetPair()
	vOp := reflect.ValueOf(p)
	col := vOp.FieldByName("Column").String()
	val := vOp.FieldByName("Value").Interface().([]Value)

	args = make([]any, 0, 10)
	if len(val) != 2 {
		return qry, args, offset, ErrPairTypeMustBeTwo
	}

	qry = col + " BETWEEN "
	for _, pr := range val {
		v, err = getFilterValue(srcData, pr)
		if err != nil {
			return qry, args, offset, err
		}
		if v == nil {
			return qry, args, offset, err
		}
		switch v.(type) {
		case Null:
			return qry, args, offset, err
		}
		offset++
		ph = strings.TrimSpace(ph)
		qry += cma + " " + ph
		if inSeq && ph != "?" {
			qry += strconv.Itoa(offset)
		}
		args = append(args, v)
		cma = " AND "
	}

	return qry, args, offset, nil
}
