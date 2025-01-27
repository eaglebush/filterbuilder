package filterbuilder

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	ssd "github.com/shopspring/decimal"
	"golang.org/x/exp/constraints"
)

var (
	ErrNoFilterSet           error = errors.New("no filters set")
	ErrColumnNotFound        error = errors.New("column not found")
	ErrDataNotSet            error = errors.New("data was not set")
	ErrInvalidFieldName      error = errors.New("invalid field name")
	ErrDataIsNotStruct       error = errors.New("data is not struct")
	ErrDataAssertionMismatch error = errors.New("data assertion mismatch")
	ErrTypeReflectionInvalid error = errors.New("type reflection invalid")
)

type FieldTypeConstraint interface {
	constraints.Ordered | time.Time | ssd.Decimal | bool
}

// NewFilter creates a new Filter object
func NewFilter(eq []Pair, paramInSequence bool, paramPlaceHolder string) *Filter {
	if paramPlaceHolder == "" {
		if paramInSequence {
			paramPlaceHolder = "@p"
		} else {
			paramPlaceHolder = "?"
		}
	}
	return &Filter{
		Eq:          eq,
		InSequence:  paramInSequence,
		Placeholder: paramPlaceHolder,
	}
}

// RawPair simplifies raw pair.
// Pairs reads the value argument raw.
func RawPair(column string, value interface{}) Pair {
	return Pair{
		Column: column,
		Value: Value{
			Src: value,
			Raw: true,
		},
	}
}

// RawMultiPair simplifies raw multi-pair.
// Pairs reads the value argument raw.
func RawMultiPair(column string, value ...interface{}) MultiFieldPair {
	v := make([]Value, 0, len(value))
	for _, a := range value {
		v = append(v, Value{
			Raw: true,
			Src: a,
		})
	}
	return MultiFieldPair{
		Column: column,
		Value:  v,
	}
}

// DataPair simplifies data pair.
// Pairs reads the Data field values via fieldName argument.
func DataPair(column string, fieldName string) Pair {
	return Pair{
		Column: column,
		Value: Value{
			Src: fieldName,
		},
	}
}

// DataMultiPair simplifies data multi-pair.
// Pairs reads the Data field values via fieldName argument.
func DataMultiPair(column string, fieldName ...string) MultiFieldPair {
	v := make([]Value, 0, len(fieldName))
	for _, a := range fieldName {
		v = append(v, Value{
			Src: a,
		})
	}
	return MultiFieldPair{
		Column: column,
		Value:  v,
	}
}

// NewPairs simplify initialization of Pairs
func NewPairs(pairs ...Pair) []Pair {
	return pairs
}

// NewMultiPairs simplify initialization of multi-field pairs
func NewMultiPairs(pairs ...MultiFieldPair) []MultiFieldPair {
	return pairs
}

// SetPair sets a pair array with the specified column and value
//
// If the column does not exist, it will create one
func SetPair(selector *[]Pair, column string, value Value) {
	if selector == nil {
		selector = &[]Pair{}
	}
	found := false
	if len(*selector) > 0 {
		slctr := *selector
		for i, cv := range *selector {
			if strings.EqualFold(cv.Column, column) {
				cv.Value = value
				slctr[i] = cv
				*selector = slctr
				return
			}
		}
	}
	if !found {
		*selector = append(*selector,
			Pair{
				Column: column,
				Value:  value,
			},
		)
	}
}

// SetMultiPair sets a multi-pair array with the specified column and value
//
// If the column does not exist, it will create one
func SetMultiPair(selector *[]MultiFieldPair, column string, value []Value) {
	if selector == nil {
		selector = &[]MultiFieldPair{}
	}
	found := false
	if len(*selector) > 0 {
		slctr := *selector
		for i, cv := range *selector {
			if strings.EqualFold(cv.Column, column) {
				cv.Value = value
				slctr[i] = cv
				*selector = slctr
				return
			}
		}
	}
	if !found {
		*selector = append(*selector,
			MultiFieldPair{
				Column: column,
				Value:  value,
			},
		)
	}
}

// BuildFunc is a builder compatible with QueryBuilder's FilterFunc
func (fb *Filter) BuildFunc(poff int, pchar string, pseq bool) ([]string, []interface{}) {
	fb.Offset = poff
	fb.Placeholder = pchar
	fb.InSequence = pseq
	qry, args, _ := fb.Build()
	return qry, args
}

// Build the filter query
func (fb *Filter) Build() ([]string, []interface{}, error) {

	var (
		sql  []string
		args []interface{}
		err  error
		v    interface{}
		vs   []interface{}
	)

	if fb.Placeholder == "" {
		if fb.InSequence {
			fb.Placeholder = "@p"
		} else {
			fb.Placeholder = "?"
		}
	}

	sql = make([]string, 0)
	args = make([]interface{}, 0)

	if len(fb.In) == 0 &&
		len(fb.NotIn) == 0 &&
		len(fb.Ne) == 0 &&
		len(fb.Eq) == 0 &&
		len(fb.Lk) == 0 &&
		len(fb.Between) == 0 &&
		!fb.AllowNoFilters {
		return sql, args, ErrNoFilterSet
	}

	tmp := ""

	// Get Equality filters
	for _, sv := range fb.Eq {
		v, err = fb.Value(sv.Value)
		if err != nil {
			return sql, args, err
		}
		if v == nil {
			continue
		}
		switch v.(type) {
		case Null:
			tmp = sv.Column + " IS NULL"
		default:
			fb.Offset++
			tmp = sv.Column + " = " + fb.Placeholder
			if fb.InSequence {
				tmp += strconv.Itoa(fb.Offset)
			}
			args = append(args, v)
		}
		sql = append(sql, tmp)
	}

	// Get  Non-Equality filters
	for _, sv := range fb.Ne {
		v, err = fb.Value(sv.Value)
		if err != nil {
			return sql, args, err
		}
		if v == nil {
			continue
		}
		switch v.(type) {
		case Null:
			tmp = sv.Column + " IS NOT NULL"
		default:
			fb.Offset++
			tmp = sv.Column + " <> " + fb.Placeholder
			if fb.InSequence {
				tmp += strconv.Itoa(fb.Offset)
			}
			args = append(args, v)
		}
		sql = append(sql, tmp)
	}

	// Get Like filters
	for _, sv := range fb.Lk {
		v, err = fb.Value(sv.Value)
		if err != nil {
			return sql, args, err
		}
		if v == nil {
			continue
		}
		fb.Offset++
		tmp = sv.Column + " LIKE " + fb.Placeholder
		if fb.InSequence {
			tmp += strconv.Itoa(fb.Offset)
		}
		sql = append(sql, tmp)
		args = append(args, v)
	}

	var (
		cma  string
		prms string
	)

	// Get In filters
	for _, sv := range fb.In {
		vs, err = fb.getMultiPairValue(sv.Value)
		if err != nil {
			return sql, args, err
		}
		tmp = sv.Column + " IN (%s) "
		cma = ""
		prms = ""
		for _, vx := range vs {
			if vx == nil {
				continue
			}
			fb.Offset++
			prms += cma + fb.Placeholder
			if fb.InSequence {
				prms += strconv.Itoa(fb.Offset)
			}
			cma = ","
			args = append(args, vx)
		}
		tmp = fmt.Sprintf(tmp, prms)
		sql = append(sql, tmp)
	}

	// Get Not In filters
	for _, sv := range fb.NotIn {
		vs, err = fb.getMultiPairValue(sv.Value)
		if err != nil {
			return sql, args, err
		}
		tmp = sv.Column + " NOT IN (%s) "
		cma = ""
		prms = ""
		for _, vx := range vs {
			if vx == nil {
				continue
			}
			fb.Offset++
			prms += cma + fb.Placeholder
			if fb.InSequence {
				prms += strconv.Itoa(fb.Offset)
			}
			cma = ","
			args = append(args, vx)
		}

		tmp = fmt.Sprintf(tmp, prms)
		sql = append(sql, tmp)
	}

	// Get Between filters
	for _, sv := range fb.Between {
		vs, err = fb.getMultiPairValue(sv.Value)
		if err != nil {
			return sql, args, err
		}
		tmp = sv.Column + " BETWEEN %s "
		cma = ""
		prms = ""
		for i, vx := range vs {
			if vx == nil {
				continue
			}
			fb.Offset++
			prms += cma + fb.Placeholder
			if fb.InSequence {
				prms += strconv.Itoa(fb.Offset)
			}
			cma = " AND "
			args = append(args, vx)

			// break because between only accepts 2
			if i > 2 {
				break
			}
		}
		tmp = fmt.Sprintf(tmp, prms)
		sql = append(sql, tmp)
	}
	return sql, args, nil
}

// ValueFor gets the value of the filter instance by column lookup
func (fb *Filter) ValueFor(col string) (interface{}, error) {
	for _, v := range fb.Eq {
		if strings.EqualFold(v.Column, col) {
			return fb.Value(v.Value)
		}
	}
	for _, v := range fb.Ne {
		if strings.EqualFold(v.Column, col) {
			return fb.Value(v.Value)
		}
	}
	for _, v := range fb.Lk {
		if strings.EqualFold(v.Column, col) {
			return fb.Value(v.Value)
		}
	}
	for _, v := range fb.In {
		if strings.EqualFold(v.Column, col) {
			return fb.getMultiPairValue(v.Value)
		}
	}
	for _, v := range fb.NotIn {
		if strings.EqualFold(v.Column, col) {
			return fb.getMultiPairValue(v.Value)
		}
	}
	for _, v := range fb.Between {
		if strings.EqualFold(v.Column, col) {
			return fb.getMultiPairValue(v.Value)
		}
	}
	return nil, ErrColumnNotFound
}

// ValueFor is a static way to get the value of the filter by column lookup
func ValueFor[T FieldTypeConstraint](fb *Filter, col string) (T, error) {
	ifc, err := fb.ValueFor(col)
	if err != nil {
		return *new(T), err
	}
	// get value thru reflect
	var vx interface{}
	t := reflect.ValueOf(ifc)
	if t.Kind() == reflect.Ptr {
		if fx := t.Elem(); !fx.IsValid() {
			return *new(T), ErrTypeReflectionInvalid
		}
		vx = t.Elem().Interface()
	} else {
		vx = t.Interface()
	}
	val, ok := vx.(T)
	if !ok {
		return *new(T), ErrDataAssertionMismatch
	}
	return val, err
}

// Weld joins an existing SQL string and its arguments with the results from the Build function
func (fb *Filter) Weld(sql string, args []interface{}, paramoffset int) (string, []interface{}, error) {
	fb.Offset = paramoffset
	fexp, fargs, err := fb.Build()
	if err != nil {
		return sql, args, err
	}
	if len(fexp) > 0 {
		// remove trailing space and semi-colon
		sql = strings.TrimSpace(sql)
		sql = strings.TrimRight(sql, `;`)
		sql += " WHERE " + strings.Join(fexp, " AND ")
	}
	if len(fargs) > 0 {
		args = append(args, fargs...)
	}
	return sql, args, nil
}

// Value gets the actual value of the struct field or the raw value that has been set
func (fb *Filter) Value(p Value) (interface{}, error) {
	if p.Raw {
		if p.Src == nil {
			return Null(true), nil
		}
		rv := reflect.ValueOf(p.Src)
		if rv.Kind() != reflect.Ptr {
			return p.Src, nil
		}
		vv := rv.Elem()
		if !vv.IsValid() {
			return Null(true), nil
		}
		return p.Src, nil
	}
	if fb.Data == nil {
		return nil, ErrDataNotSet
	}
	// get value thru reflect
	t := reflect.ValueOf(fb.Data)
	fld, ok := p.Src.(string)
	if !ok {
		return nil, ErrInvalidFieldName
	}
	if t.Kind() != reflect.Struct {
		return nil, ErrDataIsNotStruct
	}
	f := t.FieldByNameFunc(func(s string) bool {
		return strings.EqualFold(fld, s)
	})

	var vx interface{}
	if f.Kind() == reflect.Ptr {
		if fx := f.Elem(); !fx.IsValid() {
			return nil, nil
		}
		vx = f.Elem().Interface()
	} else {
		vx = f.Interface()
	}
	return vx, nil
}

// Valid checks if any filters were defined
func (fb *Filter) Valid() bool {
	return len(fb.Eq) > 0 ||
		len(fb.Ne) > 0 ||
		len(fb.Lk) > 0 ||
		len(fb.In) > 0 ||
		len(fb.NotIn) > 0 ||
		len(fb.Between) > 0
}

func (fb *Filter) getMultiPairValue(p []Value) ([]interface{}, error) {
	var (
		err  error
		args []interface{}
		v    interface{}
	)
	args = make([]interface{}, 0)
	for _, mv := range p {
		v, err = fb.Value(mv)
		if err != nil {
			return args, err
		}
		args = append(args, v)
	}
	return args, nil
}
