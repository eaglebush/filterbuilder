package filterbuilder

import (
	"testing"
)

func TestNew(t *testing.T) {
	fb := New(InSequence(true), Placeholder("@p"))
	fb.Eq = []Eq{
		{Column: "first_name", Value: Value{Src: "Zaldy", Raw: true}},
		{Column: "last_name", Value: Value{Src: "Baguinon", Raw: true}},
	}
	fb.Ne = append(fb.Ne, Ne{Column: "first_name", Value: Value{Src: "James", Raw: true}})
	fb.Ne = append(fb.Ne, Ne{Column: "last_name", Value: Value{Src: "Lumibao", Raw: true}})

	sql, args, err := fb.Build()
	if err != nil {
		t.Fail()
	}

	t.Log(sql)
	t.Log(args)
}

func TestShortCutFunc(t *testing.T) {
	fb := New(InSequence(true))
	fb.Eq = NewPairs(
		EqRawPair("first_name", "Zaldy"),
		EqRawPair("last_name", "Baguinon"),
		EqRawPair("middle_name", "Gonzales"),
	)
	sql, args, err := fb.Build()
	if err != nil {
		t.Logf("Error: %s", err)
		t.Fail()
	}

	for _, s := range sql {
		t.Log(s)
	}

	for i, a := range args {
		t.Logf("[%d] %v", i+1, a)
	}
}

func TestNewFunc(t *testing.T) {
	fb := New(InSequence(true))
	fb.Eq = NewPairs(
		EqRawPair("first_name", "Zaldy"),
		EqRawPair("last_name", "Baguinon"),
		EqRawPair("middle_name", "Gonzales"),
	)
	sql, args, err := fb.Build()
	if err != nil {
		t.Logf("Error: %s", err)
		t.Fail()
	}

	for _, s := range sql {
		t.Log(s)
	}

	for i, a := range args {
		t.Logf("[%d] %v", i+1, a)
	}
}

func TestOr(t *testing.T) {

	fb := New(InSequence(true), Placeholder("@p"))

	fb.Eq = []Eq{
		{Column: "first_name", Value: Value{Src: "Zaldy", Raw: true}},
		{Column: "last_name", Value: Value{Src: "Baguinon", Raw: true}},
	}
	fb.Ne = append(fb.Ne, Ne{Column: "first_name", Value: Value{Src: "James", Raw: true}})
	fb.Ne = append(fb.Ne, Ne{Column: "last_name", Value: Value{Src: "Lumibao", Raw: true}})

	// Or condition 1. Must have 2 or more expressions
	or1 := Or{
		Pair: []Filterer{
			Eq{
				Column: "[nick_name]",
				Value:  Value{Src: "James", Raw: true},
			},
			Eq{
				Column: "''",
				Value:  Value{Src: "222", Raw: true},
			},
		},
	}

	// Or condition 2
	or2 := Or{
		Pair: []Filterer{
			Eq{
				Column: "[last_name]",
				Value:  Value{Src: "Lumibao", Raw: true},
			},
			Eq{
				Column: "''",
				Value:  Value{Src: "333", Raw: true},
			},
		},
	}

	// Or condition 3
	// or3 := Or{
	// 	Pair: []Filterer{
	// 		Lk{
	// 			Column: "[first_name]",
	// 			Value:  Value{Src: "James%", Raw: true},
	// 		},
	// 	},
	// }

	// Or condition 2
	or4 := Or{
		Pair: []Filterer{
			Eq{
				Column: "[age]",
				Value:  Value{Src: 32, Raw: true},
			},
			Eq{
				Column: "[count]",
				Value:  Value{Src: 34, Raw: true},
			},
		},
	}

	fb.Or = append(fb.Or, or1)
	fb.Or = append(fb.Or, or2)
	//fb.Or = append(fb.Or, or3)
	fb.Or = append(fb.Or, or4)

	sql, args, err := fb.Build()
	if err != nil {
		t.Logf("Error: %s", err)
		t.Fail()
	}
	t.Log(sql)

	for _, s := range sql {
		t.Log(s)
	}

	for i, a := range args {
		t.Logf("%d: %v", i+1, a)
	}
}

func TestNilString(t *testing.T) {

	var nilStr *string
	name := "Zaldy"

	fb := New(InSequence(true), Placeholder("@p"))

	fb.Eq = []Eq{
		{Column: "first_name", Value: Value{Src: name, Raw: true}},
		{Column: "last_name", Value: Value{Src: "Baguinon", Raw: true}},
		{Column: "nil_str", Value: Value{Src: nilStr, Raw: true}},
	}

	sql, args, err := fb.Build()
	if err != nil {
		t.Fail()
	}

	t.Log(sql)
	t.Log(args)
}

func TestFilterByFieldValue(t *testing.T) {

	type simpleData struct {
		FirstName *string
		LastName  *string
		Age       int
	}

	fn := "Zaldy"
	ln := "Baguinon"

	sd := simpleData{
		FirstName: &fn,
		LastName:  &ln,
		Age:       46,
	}

	n := Null(true)

	fb := Filter{
		Data: sd,
		Eq: []Eq{
			{Column: "first_name", Value: Value{Src: "FirstName"}},
			{Column: "last_name", Value: Value{Src: n, Raw: true}},
		},
		Ne: []Ne{
			{Column: "first_name", Value: Value{Src: "FirstName"}},
			{Column: "last_name", Value: Value{Src: "LastName"}},
		},
		Lk: []Lk{
			{Column: "first_name", Value: Value{Src: "FirstName"}},
			{Column: "last_name", Value: Value{Src: "LastName"}},
		},
	}

	// fb := Filter{
	// 	Data: sd,
	// 	In: []MultiFieldPair{
	// 		{
	// 			Column: "Age",
	// 			Value: []Value{
	// 				{Src: 21, Raw: true},
	// 				{Src: "Age"},
	// 				{Src: 23, Raw: true},
	// 			},
	// 		},
	// 	},
	// 	NotIn: []MultiFieldPair{
	// 		{
	// 			Column: "Age",
	// 			Value: []Value{
	// 				{Src: 21, Raw: true},
	// 				{Src: 22, Raw: true},
	// 				{Src: 23, Raw: true},
	// 			},
	// 		},
	// 	},
	// 	Between: []MultiFieldPair{
	// 		{
	// 			Column: "ShipDate",
	// 			Value: []Value{
	// 				{Src: "01/01/2020", Raw: true},
	// 				{Src: "02/02/2021", Raw: true},
	// 			},
	// 		},
	// 	},
	// }

	sql, args, err := fb.Build()
	if err != nil {
		t.Fail()
	}

	t.Log(sql)
	t.Log(args)

}

// func TestOnly(t *testing.T) {

// 	type Struct struct {
// 		Name   *string
// 		Age    *int
// 		Gender *string
// 	}

// 	s := Struct{
// 		Name: new(string),
// 	}

// 	var x interface{}

// 	vo := reflect.ValueOf(s)
// 	for i := 0; i < vo.NumField(); i++ {

// 		if !vo.Field(i).CanInterface() {
// 			continue
// 		}

// 		val := vo.Field(i).Interface()

// 		switch v := val.(type) {
// 		case *string:
// 			x = val.(*string)
// 			fmt.Println(x)
// 		case *int:
// 			if v == nil {
// 				fmt.Println("nil")
// 			} else {
// 				x = val.(*int)
// 				fmt.Println(x)
// 			}

// 		default:
// 			fmt.Println(v)
// 		}

// 		_ = val
// 		_ = x

// 	}

// }

func TestValueFor(t *testing.T) {

	fb := New(InSequence(true), Placeholder("@p"))

	fb.Eq = []Eq{
		{Column: "first_name", Value: Value{Src: "Zaldy", Raw: true}},
		{Column: "last_name", Value: Value{Src: "Baguinon", Raw: true}},
		{Column: "age", Value: Value{Src: 47, Raw: true}},
		{Column: "title", Value: Value{Src: nil, Raw: true}},
	}

	fb.Ne = append(fb.Ne, Ne{Column: "first_name", Value: Value{Src: "James", Raw: true}})
	fb.Ne = append(fb.Ne, Ne{Column: "last_name", Value: Value{Src: "Lumibao", Raw: true}})

	res1, err := ValueFor[string](fb, "last_name")
	if err != nil {
		t.Log(err.Error())
	}
	t.Log(res1)

	res2, err := ValueFor[int](fb, "age")
	if err != nil {
		t.Log(err.Error())
	}
	t.Log(res2)

	res3, err := ValueFor[string](fb, "title")
	if err != nil {
		t.Log(err.Error())
	}
	t.Log(res3)
}

func TestValueForPtr(t *testing.T) {

	age := 47
	fb := New(InSequence(true), Placeholder("@p"))
	fb.Eq = []Eq{
		{Column: "first_name", Value: Value{Src: "Zaldy", Raw: true}},
		{Column: "last_name", Value: Value{Src: "Baguinon", Raw: true}},
		{Column: "age", Value: Value{Src: &age, Raw: true}},
	}
	fb.In = []In{
		InRawPair("status", "NEW", "STALE", "OLD"),
	}

	res1, err := ValueFor[string](fb, "last_name")
	if err != nil {
		t.Log(err.Error())
	}
	t.Log(res1)

	res2, err := ValueFor[int](fb, "age")
	if err != nil {
		t.Log(err.Error())
	}
	t.Log(res2)

	res3, err := ValueFor[[]string](fb, "status")
	if err != nil {
		t.Log(err.Error())
	}
	t.Log(res3)
}

func TestMakeKey(t *testing.T) {

	fb := New(InSequence(true), Placeholder("@p"))
	fb.Eq = []Eq{
		{Column: "first_name", Value: Value{Src: "Zaldy", Raw: true}},
		{Column: "last_name", Value: Value{Src: "Baguinon", Raw: true}},
	}

	fb.Ne = append(fb.Ne, Ne{Column: "first_name", Value: Value{Src: "James", Raw: true}})
	fb.Ne = append(fb.Ne, Ne{Column: "last_name", Value: Value{Src: "Lumibao", Raw: true}})
	fb.Lk = append(fb.Lk, Lk{Column: "middle_name", Value: Value{Src: "Garcia", Raw: true}})
	fb.In = append(fb.In,
		In{
			Column: "stooge",
			Value: []Value{
				{Src: "Larry", Raw: true},
				{Src: "Curly", Raw: true},
				{Src: "Moe", Raw: true},
			},
		})
	fb.NotIn = append(fb.NotIn,
		Ni{
			Column: "nick_name",
			Value: []Value{
				{Src: "Tito", Raw: true},
				{Src: "Vic", Raw: true},
				{Src: "Joey", Raw: true},
			},
		})
	t.Log(fb.MakeKey())
}

func TestHash(t *testing.T) {
	fb := New(InSequence(true), Placeholder("@p"))

	fb.Eq = []Eq{
		{Column: "first_name", Value: Value{Src: "Zaldy", Raw: true}},
		{Column: "last_name", Value: Value{Src: "Baguinon", Raw: true}},
	}

	fb.Ne = append(fb.Ne, Ne{Column: "first_name", Value: Value{Src: "James", Raw: true}})
	fb.Ne = append(fb.Ne, Ne{Column: "last_name", Value: Value{Src: "Lumibao", Raw: true}})
	fb.Lk = append(fb.Lk, Lk{Column: "middle_name", Value: Value{Src: "Garcia", Raw: true}})
	fb.In = append(fb.In,
		In{
			Column: "stooge",
			Value: []Value{
				{Src: "Larry", Raw: true},
				{Src: "Curly", Raw: true},
				{Src: "Moe", Raw: true},
			},
		})
	fb.NotIn = append(fb.NotIn,
		Ni{
			Column: "nick_name",
			Value: []Value{
				{Src: "Tito", Raw: true},
				{Src: "Vic", Raw: true},
				{Src: "Joey", Raw: true},
			},
		})
	t.Log(fb.Hash())
}
