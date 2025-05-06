package filterbuilder

// Ne is an inequality filter in SQL
type Ne struct {
	Column string `json:"column,omitempty"` // Database table column
	Value  Value  `json:"value,omitempty"`  // Struct field to get value or the value itself
}

// NeRawPair simplifies raw Ne pair.
// Pairs reads the value argument raw.
func NeRawPair(column string, value any) Ne {
	return Ne{
		Column: column,
		Value: Value{
			Src: value,
			Raw: true,
		},
	}
}

// NeDataPair simplifies data Ne pair.
// Pairs reads the value from Filter data field
func NeDataPair(column string, value any) Ne {
	return Ne{
		Column: column,
		Value: Value{
			Src: value,
		},
	}
}

func (f Ne) Build(data any, ph string, inSeq bool, offset int) (string, any, int, error) {
	return buildPair(f, data, "<>", ph, inSeq, offset)
}

func (f Ne) GetPair() any {
	return f
}
