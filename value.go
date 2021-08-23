package filterbuilder

// Null indicates the column should evaluate for NULL
type Null bool

// Value struct
type Value struct {
	Src interface{} // Struct field to get value or the value itself
	Raw bool        // When true, this value is a value set. If false, the value is a struct field.
}

// Pair struct
type Pair struct {
	Column string // Database table column
	Value  Value  // Struct field to get value or the value itself
}

// MultiFieldPair struct
type MultiFieldPair struct {
	Column string  // Database table column
	Value  []Value // Struct field to get value
}
