# FilterBuilder #

The filterbuilder is a stand-alone library to create filter for the [querybuilder](https://github.com/eaglebush/querybuilder) library.
The library would return an output of an SQL with parameterized code and the variables for the parameters.

Here is a sample code taken from its test file:

```go
    fb := NewFilter(
		[]Pair{
			{Column: "first_name", Value: Value{Src: "Zaldy", Raw: true}},
			{Column: "last_name", Value: Value{Src: "Baguinon", Raw: true}},
		}, true, "@p")

	fb.Ne = append(fb.Ne, Pair{Column: "first_name", Value: Value{Src: "James", Raw: true}})
	fb.Ne = append(fb.Ne, Pair{Column: "last_name", Value: Value{Src: "Lumibao", Raw: true}})

	sql, args, err := fb.Build()
	if err != nil {
		t.Fail()
	}

	t.Log(sql)
	t.Log(args)
```

The filters are defined with a structure as follows:

```go
    type Filter struct {
        Data           any              `json:"data,omitempty"`
        Eq             []Pair           `json:"eq,omitempty"`               // Equality pair
        Ne             []Pair           `json:"ne,omitempty"`               // Not equality pair
        Lk             []Pair           `json:"lk,omitempty"`               // Like pair
        In             []MultiFieldPair `json:"in,omitempty"`               // In column pair.
        NotIn          []MultiFieldPair `json:"not_in,omitempty"`           // Not In column pair
        Between        []MultiFieldPair `json:"between,omitempty"`          // Between column pair
        Placeholder    string           `json:"placeholder,omitempty"`      // Parameter place holder
        InSequence     bool             `json:"in_sequence,omitempty"`      // Parameter place holders would be numbered in sequence
        Offset         int              `json:"offset,omitempty"`           // Sets the start of parameter number
        AllowNoFilters bool             `json:"allow_no_filters,omitempty"` // Allow no filter upon building
    }
```
It can be set programmatically or as Json object to be parsed and fed into a REST endpoint. This is the Json equivalent of the structure above:

```json
    {
        "eq": [
            {"column": "first_name", "value": "Eagle"},
            {"column": "last_name", "value": "Bush"},
        ]
    }

```
The rest of the fields aside from `eq`,`ne`, `lk`, `in`, `not_in`, `between` can also be set via Json snippet, but it can cause security issues.