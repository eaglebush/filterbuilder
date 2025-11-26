package filterbuilder

type Gt struct {
	Column string
	Value  Value
}

func GtRawPair(column string, value any) Gt {
	return Gt{
		Column: column,
		Value:  Value{Src: value, Raw: true},
	}
}

func GtDataPair(column string, field string) Gt {
	return Gt{
		Column: column,
		Value:  Value{Src: field},
	}
}

func (f Gt) GetPair() any {
	return f
}

func (f Gt) Build(data any, ph string, inSeq bool, offset int) (string, any, int, error) {
	return buildPair(f, data, ">", ph, inSeq, offset)
}
