package filterbuilder

import (
	"strings"
)

type Group struct {
	And []Filterer
}

func (g Group) GetPair() any {
	return g
}

func (g Group) Build(data any, ph string, inSeq bool, offset int) (string, any, int, error) {
	parts := []string{}
	args := []any{}

	for _, f := range g.And {
		str, rv, newOffset, err := f.Build(data, ph, inSeq, offset)
		if err != nil {
			return "", nil, offset, err
		}
		offset = newOffset

		parts = append(parts, str)

		if isSliceType(rv) {
			args = append(args, rv.([]any)...)
		} else if rv != nil {
			args = append(args, rv)
		}
	}

	return "(" + strings.Join(parts, " AND ") + ")", args, offset, nil
}
