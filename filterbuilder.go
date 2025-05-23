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
)

type (
	FilterConstraints interface {
		~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 |
			~uint32 | ~uint64 | ~uintptr | ~float32 | ~float64 | ~string |
			~[]int | ~[]int8 | ~[]int16 | ~[]int32 | ~[]int64 | ~[]uint | ~[]uint8 | ~[]uint16 |
			~[]uint32 | ~[]uint64 | ~[]uintptr | ~[]float32 | ~[]float64 | ~[]string |
			time.Time | ssd.Decimal | bool | []time.Time | []ssd.Decimal | []bool
	}
)

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
func RawPair(column string, value any) Pair {
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
func RawMultiPair(column string, value ...any) MultiFieldPair {
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
		v    any
		vs   []any
	)

	if fb.Placeholder == "" {
		if fb.InSequence {
			fb.Placeholder = "@p"
		} else {
			fb.Placeholder = "?"
		}
	}

	sql = make([]string, 0)
	args = make([]any, 0)

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

	for _, ors := range fb.Or {
		if len(ors) > 0 && len(ors) < 2 {
			return sql, args, ErrPairTypeMustHaveMoreThanTwo
		}
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

	// Get Or filters
	for _, ors := range fb.Or {
		tmps := []string{}
		for _, sv := range ors {
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
			tmps = append(tmps, tmp)
		}
		if len(tmps) > 0 {
			tmp = "(" + strings.Join(tmps, " OR ") + ")"
			sql = append(sql, tmp)
		}
	}

	// Get Non-Equality filters
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
		cma, prms string
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
func (fb *Filter) ValueFor(col string) (any, error) {
	for _, v := range fb.Eq {
		if strings.EqualFold(v.Column, col) {
			return fb.Value(v.Value)
		}
	}
	for _, vfs := range fb.Or {
		for _, v := range vfs {
			if strings.EqualFold(v.Column, col) {
				return fb.Value(v.Value)
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
func ValueFor[T FilterConstraints](fb *Filter, col string) (T, error) {
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
	} else if t.Kind() == reflect.Slice {
		tType := reflect.TypeOf((*T)(nil)).Elem()        // Gets []string if T is []string
		toS := tType.Elem()                              // Gets string (element type)
		vy := reflect.MakeSlice(tType, t.Len(), t.Len()) // Makes []string as reflect.Value
		for i := range t.Len() {
			item := t.Index(i).Interface()
			val := reflect.ValueOf(item)
			if !val.Type().AssignableTo(toS) {
				panic(fmt.Sprintf("element %d is not assignable to %s", i, toS))
			}
			vy.Index(i).Set(val)
		}
		vx = vy.Interface()
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

	var vx any
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
		for _, v := range vfs {
			if sb.Len() > 0 {
				sb.WriteString("-")
			}
			sb.WriteString(sanitizeColumnForHash(v.Column))
			sb.WriteString("=")
			val, _ := fb.Value(v.Value)
			sb.WriteString("\"" + sanitizeValueForHash(anyToString(val)) + "\"")
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
		vals, _ := fb.getMultiPairValue(v.Value)
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
		vals, _ := fb.getMultiPairValue(v.Value)
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
		vals, _ := fb.getMultiPairValue(v.Value)
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

func (fb *Filter) getMultiPairValue(p []Value) ([]any, error) {
	var (
		err  error
		args []any
		v    any
	)
	args = make([]any, 0)
	for _, mv := range p {
		v, err = fb.Value(mv)
		if err != nil {
			return args, err
		}
		args = append(args, v)
	}
	return args, nil
}
