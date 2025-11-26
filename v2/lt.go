package filterbuilder

type Lt struct {
	Column string
	Value  Value
}

func LtRawPair(column string, value any) Lt {
	return Lt{
		Column: column,
		Value:  Value{Src: value, Raw: true},
	}
}

func LtDataPair(column string, field string) Lt {
	return Lt{
		Column: column,
		Value:  Value{Src: field},
	}
}

func (f Lt) GetPair() any {
	return f
}

func (f Lt) Build(data any, ph string, inSeq bool, offset int) (string, any, int, error) {
	return buildPair(f, data, "<", ph, inSeq, offset)
}
