package filterbuilder

// Filter - the filter struct
type Filter struct {
	Data           any              `json:"data,omitempty"`
	Eq             []Pair           `json:"eq,omitempty"`               // Equality pairs
	Ne             []Pair           `json:"ne,omitempty"`               // Not equality pairs
	Lk             []Pair           `json:"lk,omitempty"`               // Like pairs
	Or             [][]Pair         `json:"or,omitempty"`               // Or pairs
	In             []MultiFieldPair `json:"in,omitempty"`               // In column pair.
	NotIn          []MultiFieldPair `json:"not_in,omitempty"`           // Not In column pair
	Between        []MultiFieldPair `json:"between,omitempty"`          // Between column pair
	Placeholder    string           `json:"placeholder,omitempty"`      // Parameter place holder
	InSequence     bool             `json:"in_sequence,omitempty"`      // Parameter place holders would be numbered in sequence
	Offset         int              `json:"offset,omitempty"`           // Sets the start of parameter number
	AllowNoFilters bool             `json:"allow_no_filters,omitempty"` // Allow no filter upon building
}
