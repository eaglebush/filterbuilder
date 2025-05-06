package filterbuilder

// Eq is an equality filter in SQL
type Eq struct {
	Column string `json:"column,omitempty"` // Database table column
	Value  Value  `json:"value,omitempty"`  // Struct field to get value or the value itself
}

// EqRawPair simplifies raw Eq pair.
// Pairs reads the value argument raw.
func EqRawPair(column string, value any) Eq {
	return Eq{
		Column: column,
		Value: Value{
			Src: value,
			Raw: true,
		},
	}
}

// EqDataPair simplifies data Eq pair.
// Pairs reads the value from Filter data field
func EqDataPair(column string, value any) Eq {
	return Eq{
		Column: column,
		Value: Value{
			Src: value,
		},
	}
}

func (f Eq) Build(data any, ph string, inSeq bool, offset int) (string, any, int, error) {
	return buildPair(f, data, "=", ph, inSeq, offset)
}

func (f Eq) GetPair() any {
	return f
}
