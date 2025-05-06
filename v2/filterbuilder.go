package filterbuilder

import (
	"crypto/sha256"
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
	ErrNoFilterSet                 error = errors.New("no filters set")
	ErrColumnNotFound              error = errors.New("column not found")
	ErrDataNotSet                  error = errors.New("data was not set")
	ErrInvalidFieldName            error = errors.New("invalid field name")
	ErrDataIsNotStruct             error = errors.New("data is not struct")
	ErrDataAssertionMismatch       error = errors.New("data assertion mismatch")
	ErrTypeReflectionInvalid       error = errors.New("type reflection invalid")
	ErrPairTypeMustHaveMoreThanTwo error = errors.New("pair type must have more than two")
	ErrPairTypeMustBeTwo           error = errors.New("pair type must be two")
	ErrSourceIsNil                 error = errors.New("source is nil")
)

type (
	FieldTypeConstraint interface {
		constraints.Ordered | time.Time | ssd.Decimal | bool
	}

	// // Filter option data
	// fod struct {
	// 	ph    string
	// 	inSeq bool
	// 	off   int
	// 	nof   bool
	// }

	FilterOption func(*Filter)
)

// New creates a new Filter object
func New(opts ...FilterOption) *Filter {
	fltr := Filter{
		Placeholder: "?",
	}
	for _, opt := range opts {
		opt(&fltr)
	}
	return &fltr
}

// Placeholder changes the default placeholder (?)
func Placeholder(ph string) FilterOption {
	ph = strings.TrimSpace(ph)
	if ph == "" {
		ph = "?"
	}
	return func(f *Filter) {
		f.Placeholder = ph
	}
}

// InSequence sets the parameter to a numbered placeholder
func InSequence(value bool) FilterOption {
	return func(f *Filter) {
		f.InSequence = value
	}
}

// Offset changes the default parameter offset
func Offset(off int) FilterOption {
	return func(f *Filter) {
		f.Offset = off
	}
}

// AllowNoFilters allows filter to build
func AllowNoFilters(value bool) FilterOption {
	return func(f *Filter) {
		f.AllowNoFilters = value
	}
}

// NewPairs simplify initialization of Filterer
func NewPairs[T Filterer](pairs ...T) []T {
	return pairs
}

// BuildFunc is a builder compatible with QueryBuilder's FilterFunc
func (fb *Filter) BuildFunc(poff int, pchar string, pseq bool) ([]string, []any) {
	fb.Offset = poff
	fb.Placeholder = pchar
	fb.InSequence = pseq
	qry, args, _ := fb.Build()
	return qry, args
}

// Build the filter query
func (fb *Filter) Build() ([]string, []any, error) {

	var (
		sql  []string
		args []any
		err  error
		rv   any
		str  string
	)

	if fb.Placeholder == "" {
		if fb.InSequence {
			fb.Placeholder = "@p"
		} else {
			fb.Placeholder = "?"
		}
	}

	sql = make([]string, 0, 10)
	args = make([]any, 0, 10)

	if len(fb.In) == 0 &&
		len(fb.NotIn) == 0 &&
		len(fb.Ne) == 0 &&
		len(fb.Eq) == 0 &&
		len(fb.Or) == 0 &&
		len(fb.Lk) == 0 &&
		len(fb.Between) == 0 &&
		!fb.AllowNoFilters {
		return sql, args, ErrNoFilterSet
	}

	// Check if Ors pair is two or more
	for _, ors := range fb.Or {
		if len(ors.Pair) < 2 {
			return sql, args, ErrPairTypeMustHaveMoreThanTwo
		}
	}

	// Get Equality filters
	for _, sv := range fb.Eq {
		str, rv, fb.Offset, err = sv.Build(fb.Data, fb.Placeholder, fb.InSequence, fb.Offset)
		if err != nil {
			return sql, args, err
		}
		if rv != nil {
			args = append(args, rv)
		}
		sql = append(sql, str)
	}

	// Get Or filters
	// An Or is an array of Filterer
	// A group of Or is joined by an AND clause
	for _, ors := range fb.Or {
		str := ""
		orsarr := make([]string, 0)
		for _, orx := range ors.Pair {
			str, rv, fb.Offset, err = orx.Build(fb.Data, fb.Placeholder, fb.InSequence, fb.Offset)
			if err != nil {
				return sql, args, err
			}
			if isSliceType(rv) {
				args = append(args, rv.([]any)...)
			} else {
				args = append(args, rv)
			}
			orsarr = append(orsarr, str)
		}
		sql = append(sql, "("+strings.Join(orsarr, " OR ")+")")
	}

	// Get Non-Equality filters
	for _, sv := range fb.Ne {
		str, rv, fb.Offset, err = sv.Build(fb.Data, fb.Placeholder, fb.InSequence, fb.Offset)
		if err != nil {
			return sql, args, err
		}
		if rv != nil {
			args = append(args, rv)
		}
		sql = append(sql, str)
	}

	// Get Like filters
	for _, sv := range fb.Lk {
		str, rv, fb.Offset, err = sv.Build(fb.Data, fb.Placeholder, fb.InSequence, fb.Offset)
		if err != nil {
			return sql, args, err
		}
		if rv != nil {
			args = append(args, rv)
		}
		sql = append(sql, str)
	}

	// Get In filters
	for _, sv := range fb.In {
		str, rv, fb.Offset, err = sv.Build(fb.Data, fb.Placeholder, fb.InSequence, fb.Offset)
		if err != nil {
			return sql, args, err
		}
		rvs := rv.([]any)
		if len(rvs) > 0 {
			args = append(args, rv.([]any)...)
		}
		sql = append(sql, str)
	}

	// Get Not In filters
	for _, sv := range fb.NotIn {
		str, rv, fb.Offset, err = sv.Build(fb.Data, fb.Placeholder, fb.InSequence, fb.Offset)
		if err != nil {
			return sql, args, err
		}
		rvs := rv.([]any)
		if len(rvs) > 0 {
			args = append(args, rv.([]any)...)
		}
		sql = append(sql, str)
	}

	// Get Between filters
	for _, sv := range fb.Between {
		str, rv, fb.Offset, err = sv.Build(fb.Data, fb.Placeholder, fb.InSequence, fb.Offset)
		if err != nil {
			return sql, args, err
		}
		rvs := rv.([]any)
		if len(rvs) > 0 {
			args = append(args, rv.([]any)...)
		}
		sql = append(sql, str)
	}
	return sql, args, nil
}

// ValueFor gets the value of the filter instance by column lookup
func (fb *Filter) ValueFor(col string) (any, error) {
	for _, v := range fb.Eq {
		if strings.EqualFold(v.Column, col) {
			return fb.Value(v.Value)
		}
	}
	for _, vfs := range fb.Or {
		for _, v := range vfs.Pair {
			vOp := reflect.ValueOf(v)
			colv := vOp.FieldByName("Column").String()
			if !strings.EqualFold(col, colv) {
				continue
			}
			switch v.(type) {
			case Eq, Ne, Lk:
				val := vOp.FieldByName("Value").Interface().(Value)
				return fb.Value(val)
			case Ni, In, Bw:
				val := vOp.FieldByName("Value").Interface().([]Value)
				return fb.Values(val)
			}
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
			return fb.Values(v.Value)
		}
	}
	for _, v := range fb.NotIn {
		if strings.EqualFold(v.Column, col) {
			return fb.Values(v.Value)
		}
	}
	for _, v := range fb.Between {
		if strings.EqualFold(v.Column, col) {
			return fb.Values(v.Value)
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
	var vx any
	t := reflect.ValueOf(ifc)
	if t.Kind() == reflect.Ptr {
		if fx := t.Elem(); !fx.IsValid() {
			return *new(T), ErrTypeReflectionInvalid
		}
		vx = t.Elem().Interface()
	} else {
		vx = t.Interface()
		switch vx.(type) {
		case Null:
			return *new(T), ErrSourceIsNil
		}
	}
	val, ok := vx.(T)
	if !ok {
		return *new(T), ErrDataAssertionMismatch
	}
	return val, err
}

// Weld joins an existing SQL string and its arguments with the results from the Build function
func (fb *Filter) Weld(sql string, args []any, paramoffset int) (string, []any, error) {
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
func (fb *Filter) Value(p Value) (any, error) {
	return getFilterValue(fb.Data, p)
}

// Values gets the actual values of the struct field or the raw value that has been set
func (fb *Filter) Values(p []Value) ([]any, error) {
	return getFilterValues(fb.Data, p)
}

// Value gets the actual value of the struct field or the raw value that has been set
func getFilterValue(data any, p Value) (any, error) {
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
	if data == nil {
		return nil, ErrDataNotSet
	}

	// Get value thru reflect
	t := reflect.ValueOf(data)
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

	var vx any
	if f.Kind() == reflect.Ptr {
		if fx := f.Elem(); !fx.IsValid() {
			return nil, nil
		}
		vx = f.Elem().Interface()
	} else {
		vx = f.Interface()
		switch vx.(type) {
		case Null:
			return nil, ErrSourceIsNil
		}
	}
	return vx, nil
}

func getFilterValues(data any, p []Value) ([]any, error) {
	var (
		err  error
		args []any
		v    any
	)
	args = make([]any, 0)
	for _, mv := range p {
		v, err = getFilterValue(data, mv)
		if err != nil {
			return args, err
		}
		args = append(args, v)
	}
	return args, nil
}

// Valid checks if any filters were defined
func (fb *Filter) Valid() bool {
	return len(fb.Eq) > 0 ||
		len(fb.Or) > 0 ||
		len(fb.Ne) > 0 ||
		len(fb.Lk) > 0 ||
		len(fb.In) > 0 ||
		len(fb.NotIn) > 0 ||
		len(fb.Between) > 0
}

// MakeKey creates a unique key out of the filters created
func (fb *Filter) MakeKey() string {
	sb := strings.Builder{}
	for _, v := range fb.Eq {
		if sb.Len() > 0 {
			sb.WriteString("-")
		}
		sb.WriteString(sanitizeColumnForHash(v.Column))
		sb.WriteString("=")
		val, _ := fb.Value(v.Value)
		sb.WriteString("\"" + sanitizeValueForHash(anyToString(val)) + "\"")
	}
	for _, vfs := range fb.Or {
		for _, v := range vfs.Pair {
			if sb.Len() > 0 {
				sb.WriteString("-")
			}
			vOp := reflect.ValueOf(v)
			col := vOp.FieldByName("Column").String()
			switch v.(type) {
			case Eq, Ne, Lk:
				val := vOp.FieldByName("Value").Interface().(Value)
				sb.WriteString(sanitizeColumnForHash(col))
				sb.WriteString("=")
				valf, _ := fb.Value(val)
				sb.WriteString("\"" + sanitizeValueForHash(anyToString(valf)) + "\"")
			case Ni, In, Bw:
				val := vOp.FieldByName("Value").Interface().([]Value)
				sb.WriteString(sanitizeColumnForHash(col))
				sb.WriteString("=")
				valf, _ := fb.Values(val)
				for _, vf := range valf {
					sb.WriteString("\"" + sanitizeValueForHash(anyToString(vf)) + "\"")
				}
			}
		}
	}
	for _, v := range fb.Ne {
		if sb.Len() > 0 {
			sb.WriteString("-")
		}
		sb.WriteString(sanitizeColumnForHash(v.Column))
		sb.WriteString("=!")
		val, _ := fb.Value(v.Value)
		sb.WriteString("\"" + sanitizeValueForHash(anyToString(val)) + "\"")
	}
	for _, v := range fb.Lk {
		if sb.Len() > 0 {
			sb.WriteString("-")
		}
		sb.WriteString(sanitizeColumnForHash(v.Column))
		sb.WriteString("=%\"")
		val, _ := fb.Value(v.Value)
		sb.WriteString(sanitizeValueForHash(anyToString(val)))
		sb.WriteString("\"")
	}
	for _, v := range fb.In {
		if sb.Len() > 0 {
			sb.WriteString("-")
		}
		sb.WriteString(sanitizeColumnForHash(v.Column))
		sb.WriteString("=|\"")
		vals, _ := getFilterValues(fb.Data, v.Value)
		for i, val := range vals {
			sb.WriteString(sanitizeValueForHash(anyToString(val)))
			if i < len(vals)-1 {
				sb.WriteString(",")
			}
		}
		sb.WriteString("\"")
	}
	for _, v := range fb.NotIn {
		if sb.Len() > 0 {
			sb.WriteString("-")
		}
		sb.WriteString(sanitizeColumnForHash(v.Column))
		sb.WriteString("=!|\"")
		vals, _ := getFilterValues(fb.Data, v.Value)
		for i, val := range vals {
			sb.WriteString(sanitizeValueForHash(anyToString(val)))
			if i < len(vals)-1 {
				sb.WriteString(",")
			}
		}
		sb.WriteString("\"")
	}
	for _, v := range fb.Between {
		if sb.Len() > 0 {
			sb.WriteString("-")
		}
		sb.WriteString(sanitizeColumnForHash(v.Column))
		sb.WriteString("=+\"")
		vals, _ := getFilterValues(fb.Data, v.Value)
		for i, val := range vals {
			sb.WriteString(sanitizeValueForHash(anyToString(val)))
			if i < len(vals)-1 {
				sb.WriteString(",")
			}
		}
		sb.WriteString("\"")
	}
	return sb.String()
}

// Hash creates a hash of the filters created
func (fb *Filter) Hash() string {
	hasher := sha256.New()
	key := fb.MakeKey()
	hasher.Write([]byte(key))
	hashBytes := hasher.Sum(nil)
	return fmt.Sprintf("%x", hashBytes)
}

func sanitizeColumnForHash(col string) string {
	if col == "" {
		return ""
	}
	col = strings.TrimSpace(col)
	col = strings.ReplaceAll(col, " ", "")
	col = strings.ReplaceAll(col, "[", "")
	col = strings.ReplaceAll(col, "]", "")
	col = strings.ReplaceAll(col, "_", "-")
	col = strings.ToLower(col)
	return col
}

func sanitizeValueForHash(col string) string {
	if col == "" {
		return ""
	}
	col = strings.TrimSpace(col)
	col = strings.ReplaceAll(col, " ", "")
	return col
}

func anyToString(value any) string {
	var b string
	if value == nil {
		return ""
	}
	switch t := value.(type) {
	case string:
		b = t
	case int:
		b = strconv.FormatInt(int64(t), 10)
	case int8:
		b = strconv.FormatInt(int64(t), 10)
	case int16:
		b = strconv.FormatInt(int64(t), 10)
	case int32:
		b = strconv.FormatInt(int64(t), 10)
	case int64:
		b = strconv.FormatInt(t, 10)
	case uint:
		b = strconv.FormatUint(uint64(t), 10)
	case uint8:
		b = strconv.FormatUint(uint64(t), 10)
	case uint16:
		b = strconv.FormatUint(uint64(t), 10)
	case uint32:
		b = strconv.FormatUint(uint64(t), 10)
	case uint64:
		b = strconv.FormatUint(uint64(t), 10)
	case float32:
		b = fmt.Sprintf("%f", t)
	case float64:
		b = fmt.Sprintf("%f", t)
	case bool:
		if t {
			return "true"
		} else {
			return "false"
		}
	case time.Time:
		b = "'" + t.Format(time.RFC3339) + "'"
	case *string:
		if t == nil {
			return ""
		}
		b = *t
	case *int:
		if t == nil {
			return "0"
		}
		b = strconv.FormatInt(int64(*t), 10)
	case *int8:
		if t == nil {
			return "0"
		}
		b = strconv.FormatInt(int64(*t), 10)
	case *int16:
		if t == nil {
			return "0"
		}
		b = strconv.FormatInt(int64(*t), 10)
	case *int32:
		if t == nil {
			return "0"
		}
		b = strconv.FormatInt(int64(*t), 10)
	case *int64:
		if t == nil {
			return "0"
		}
		b = strconv.FormatInt(*t, 10)
	case *uint:
		if t == nil {
			return "0"
		}
		b = strconv.FormatUint(uint64(*t), 10)
	case *uint8:
		if t == nil {
			return "0"
		}
		b = strconv.FormatUint(uint64(*t), 10)
	case *uint16:
		if t == nil {
			return "0"
		}
		b = strconv.FormatUint(uint64(*t), 10)
	case *uint32:
		if t == nil {
			return "0"
		}
		b = strconv.FormatUint(uint64(*t), 10)
	case *uint64:
		if t == nil {
			return "0"
		}
		b = strconv.FormatUint(uint64(*t), 10)
	case *float32:
		if t == nil {
			return "0"
		}
		b = fmt.Sprintf("%f", *t)
	case *float64:
		if t == nil {
			return "0"
		}
		b = fmt.Sprintf("%f", *t)
	case *bool:
		if t == nil || !*t {
			return "false"
		}
		return "true"
	case *time.Time:
		if t == nil {
			return "'" + time.Time{}.Format(time.RFC3339) + "'"
		}
		tm := *t
		b = "'" + tm.Format(time.RFC3339) + "'"
	}

	return b
}

func isSliceType(v any) bool {
	t := reflect.TypeOf(v)
	switch t.Kind() {
	case reflect.Array:
		return true
	case reflect.Slice:
		return true
	}
	return false
}
