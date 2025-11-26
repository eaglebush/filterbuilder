package filterbuilder

type Lte struct {
	Column string
	Value  Value
}

func LteRawPair(column string, value any) Lte {
	return Lte{
		Column: column,
		Value:  Value{Src: value, Raw: true},
	}
}

func LteDataPair(column string, field string) Lte {
	return Lte{
		Column: column,
		Value:  Value{Src: field},
	}
}

func (f Lte) GetPair() any {
	return f
}

func (f Lte) Build(data any, ph string, inSeq bool, offset int) (string, any, int, error) {
	return buildPair(f, data, "<=", ph, inSeq, offset)
}
