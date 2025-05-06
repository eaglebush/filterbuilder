package filterbuilder

// Lk is a pattern seeking filter in SQL
type Lk struct {
	Column string `json:"column,omitempty"` // Database table column
	Value  Value  `json:"value,omitempty"`  // Struct field to get value or the value itself
}

// LkRawPair simplifies raw Lk pair.
// Pairs reads the value argument raw.
func LkRawPair(column string, value any) Lk {
	return Lk{
		Column: column,
		Value: Value{
			Src: value,
			Raw: true,
		},
	}
}

// LkDataPair simplifies data Lk pair.
// Pairs reads the value from Filter data field
func LkDataPair(column string, value any) Lk {
	return Lk{
		Column: column,
		Value: Value{
			Src: value,
		},
	}
}

func (f Lk) Build(data any, ph string, inSeq bool, offset int) (string, any, int, error) {
	return buildPair(f, data, "LIKE", ph, inSeq, offset)
}

func (f Lk) GetPair() any {
	return f
}
