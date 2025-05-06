package filterbuilder

// Ni is NOT IN filter in SQL
type Ni struct {
	Column string  `json:"column,omitempty"` // Database table column
	Value  []Value `json:"value,omitempty"`  // Struct field to get value
}

// NiRawPair simplifies raw Ni pair.
// Pairs reads the value argument raw.
func NiRawPair(column string, value ...any) Ni {
	values := make([]Value, 0, len(value))
	for _, v := range value {
		values = append(values, Value{
			Src: v,
			Raw: true,
		})
	}
	return Ni{
		Column: column,
		Value:  values,
	}
}

// NiDataPair simplifies Ni data multi-pair.
// Pairs reads the Data field values via fieldName argument.
func NiDataPair(column string, fieldName ...string) Ni {
	v := make([]Value, 0, len(fieldName))
	for _, a := range fieldName {
		v = append(v, Value{
			Src: a,
		})
	}
	return Ni{
		Column: column,
		Value:  v,
	}
}

func (f Ni) Build(data any, ph string, inSeq bool, offset int) (string, any, int, error) {
	return buildMembershipPair(f, data, "NOT IN", ph, inSeq, offset)
}

func (f Ni) GetPair() any {
	return f
}
