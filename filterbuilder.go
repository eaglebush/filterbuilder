package filterbuilder

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Filter - the filter struct
type Filter struct {
	Data                 interface{}
	Eq                   []Pair           // Equality pair
	Ne                   []Pair           // Not equality pair
	Lk                   []Pair           // Like pair
	In                   []MultiFieldPair // In column pair.
	NotIn                []MultiFieldPair // Not In column pair
	Between              []MultiFieldPair // Between column pair
	ParameterPlaceholder string           // parameter place holder
	ParameterOffset      int              // the start of parameter count
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

	if fb.ParameterPlaceholder == "" {
		fb.ParameterPlaceholder = "@p"
	}

	sql = make([]string, 0)
	args = make([]interface{}, 0)

	if fb.Data == nil {
		return sql, args, fmt.Errorf("no data was set")
	}

	if len(fb.In) == 0 &&
		len(fb.NotIn) == 0 &&
		len(fb.Ne) == 0 &&
		len(fb.Eq) == 0 &&
		len(fb.Lk) == 0 &&
		len(fb.Between) == 0 {
		return sql, args, fmt.Errorf("no filters set")
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
			fb.ParameterOffset++
			tmp = sv.Column + " = " + fb.ParameterPlaceholder + strconv.Itoa(fb.ParameterOffset)
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
			fb.ParameterOffset++
			tmp = sv.Column + " <> " + fb.ParameterPlaceholder + strconv.Itoa(fb.ParameterOffset)
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

		fb.ParameterOffset++

		tmp = sv.Column + " LIKE " + fb.ParameterPlaceholder + strconv.Itoa(fb.ParameterOffset)

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

			if v == nil {
				continue
			}

			fb.ParameterOffset++
			prms += cma + fb.ParameterPlaceholder + strconv.Itoa(fb.ParameterOffset)
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

			if v == nil {
				continue
			}

			fb.ParameterOffset++
			prms += cma + fb.ParameterPlaceholder + strconv.Itoa(fb.ParameterOffset)
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

			if v == nil {
				continue
			}

			fb.ParameterOffset++
			prms += cma + fb.ParameterPlaceholder + strconv.Itoa(fb.ParameterOffset)
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

func (fb *Filter) Value(p Value) (interface{}, error) {

	if p.Raw {
		return p.Src, nil
	}

	if fb.Data == nil {
		return nil, fmt.Errorf("data not set")
	}

	// get value thru reflect
	t := reflect.ValueOf(fb.Data)

	fld, ok := p.Src.(string)
	if !ok {
		return nil, fmt.Errorf("invalid field name")
	}

	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("data not struct")
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

func (fb *Filter) getMultiPairValue(p []Value) ([]interface{}, error) {

	var (
		err  error
		args []interface{}
		v    interface{}
	)

	args = make([]interface{}, 0)

	// get value thru reflect
	t := reflect.TypeOf(fb.Data)
	if t == nil {
		return nil, fmt.Errorf("no data was set")
	}

	for _, mv := range p {
		v, err = fb.Value(mv)
		if err != nil {
			return args, err
		}

		args = append(args, v)
	}

	return args, nil
}
