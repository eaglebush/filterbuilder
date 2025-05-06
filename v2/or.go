package filterbuilder

// Or is the OR expression in SQL
type Or struct {
	Pair []Filterer
}
