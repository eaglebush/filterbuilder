package filterbuilder

// Null indicates the column should evaluate for NULL
type Null bool

// Value struct
type Value struct {
	Src interface{} `json:"src,omitempty"` // Struct field to get value or the value itself
	Raw bool        `json:"raw,omitempty"` // When true, the Src was set to a raw value. When false, the value is retrieved from the struct field in the Data.
}

// Pair struct
type Pair struct {
	Column string `json:"column,omitempty"` // Database table column
	Value  Value  `json:"value,omitempty"`  // Struct field to get value or the value itself
}

// MultiFieldPair struct
type MultiFieldPair struct {
	Column string  `json:"column,omitempty"` // Database table column
	Value  []Value `json:"value,omitempty"`  // Struct field to get value
}
