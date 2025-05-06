package filterbuilder

// In is the IN filter in SQL
type In struct {
	Column string  `json:"column,omitempty"` // Database table column
	Value  []Value `json:"value,omitempty"`  // Struct field to get value
}

// InRawPair simplifies raw In pair.
// Pairs reads the value argument raw.
func InRawPair(column string, value ...any) In {
	values := make([]Value, 0, len(value))
	for _, v := range value {
		values = append(values, Value{
			Src: v,
			Raw: true,
		})
	}
	return In{
		Column: column,
		Value:  values,
	}
}

// InDataPair simplifies In data multi-pair.
// Pairs reads the Data field values via fieldName argument.
func InDataPair(column string, fieldName ...string) In {
	v := make([]Value, 0, len(fieldName))
	for _, a := range fieldName {
		v = append(v, Value{
			Src: a,
		})
	}
	return In{
		Column: column,
		Value:  v,
	}
}

func (f In) Build(data any, ph string, inSeq bool, offset int) (string, any, int, error) {
	return buildMembershipPair(f, data, "IN", ph, inSeq, offset)
}
func (f In) GetPair() any {
	return f
}
