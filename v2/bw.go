package filterbuilder

// Bw is BETWEEN in SQL filter
type Bw struct {
	Column string  `json:"column,omitempty"` // Database table column
	Value  []Value `json:"value,omitempty"`  // Struct field to get value
}

// BwRawPair simplifies raw Bw pair.
// Pairs reads the value argument raw.
func BwRawPair(column string, value ...any) Bw {
	values := make([]Value, 0, len(value))
	for _, v := range value {
		values = append(values, Value{
			Src: v,
			Raw: true,
		})
	}
	return Bw{
		Column: column,
		Value:  values,
	}
}

// BwDataPair simplifies Bw data multi-pair.
// Pairs reads the Data field values via fieldName argument.
func BwDataPair(column string, fieldName ...string) Bw {
	v := make([]Value, 0, len(fieldName))
	for _, a := range fieldName {
		v = append(v, Value{
			Src: a,
		})
	}
	return Bw{
		Column: column,
		Value:  v,
	}
}

func (f Bw) Build(data any, ph string, inSeq bool, offset int) (string, any, int, error) {
	return buildRangePair(f, data, ph, inSeq, offset)
}

func (f Bw) GetPair() any {
	return f
}
