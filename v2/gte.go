package filterbuilder

type Gte struct {
	Column string
	Value  Value
}

func GteRawPair(column string, value any) Gte {
	return Gte{
		Column: column,
		Value:  Value{Src: value, Raw: true},
	}
}

func GteDataPair(column string, field string) Gte {
	return Gte{
		Column: column,
		Value:  Value{Src: field},
	}
}

func (f Gte) GetPair() any {
	return f
}

func (f Gte) Build(data any, ph string, inSeq bool, offset int) (string, any, int, error) {
	return buildPair(f, data, ">=", ph, inSeq, offset)
}
