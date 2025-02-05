package filterbuilder

import (
	"fmt"
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {

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
}

func TestOr(t *testing.T) {

	fb := NewFilter(
		[]Pair{
			{Column: "first_name", Value: Value{Src: "Zaldy", Raw: true}},
			{Column: "last_name", Value: Value{Src: "Baguinon", Raw: true}},
		}, true, "$")

	fb.Ne = append(fb.Ne, Pair{Column: "first_name", Value: Value{Src: "James", Raw: true}})
	fb.Ne = append(fb.Ne, Pair{Column: "last_name", Value: Value{Src: "Lumibao", Raw: true}})

	or1 := []Pair{
		{Column: "[nick_name]", Value: Value{Src: "James", Raw: true}},
		{Column: "[maiden_name]", Value: Value{Src: "Garcia", Raw: true}},
	}

	or2 := []Pair{
		{Column: "''", Value: Value{Src: "222", Raw: true}},
		{Column: "[user_name]", Value: Value{Src: "minium", Raw: true}},
	}
	fb.Or = append(fb.Or, or1)
	fb.Or = append(fb.Or, or2)

	sql, args, err := fb.Build()
	if err != nil {
		t.Logf("Error: %s", err)
		t.Fail()
	}

	t.Log(sql)
	t.Log(args)
}

func TestSetPairs(t *testing.T) {
	f := Filter{
		Placeholder: "?",
	}
	SetPair(&f.Eq, "first_name", Value{
		Src: "Bob",
		Raw: true,
	})
	SetPair(&f.Eq, "last_name", Value{
		Src: "Odenkirk",
		Raw: true,
	})
	SetPair(&f.Eq, "last_name", Value{
		Src: "Odenkirk1",
		Raw: true,
	})
	t.Logf("%v", f)
}

func TestNilString(t *testing.T) {

	var nilStr *string
	name := "Zaldy"

	fb := NewFilter(
		[]Pair{
			{Column: "first_name", Value: Value{Src: name, Raw: true}},
			{Column: "last_name", Value: Value{Src: "Baguinon", Raw: true}},
			{Column: "nil_str", Value: Value{Src: nilStr, Raw: true}},
		}, true, "@p")

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
		Eq: []Pair{
			{Column: "first_name", Value: Value{Src: "FirstName"}},
			{Column: "last_name", Value: Value{Src: n, Raw: true}},
		},
		Ne: []Pair{
			{Column: "first_name", Value: Value{Src: "FirstName"}},
			{Column: "last_name", Value: Value{Src: "LastName"}},
		},
		Lk: []Pair{
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

func TestOnly(t *testing.T) {

	type Struct struct {
		Name   *string
		Age    *int
		Gender *string
	}

	s := Struct{
		Name: new(string),
	}

	var x interface{}

	vo := reflect.ValueOf(s)
	for i := 0; i < vo.NumField(); i++ {

		if !vo.Field(i).CanInterface() {
			continue
		}

		val := vo.Field(i).Interface()

		switch v := val.(type) {
		case *string:
			x = val.(*string)
			fmt.Println(x)
		case *int:
			if v == nil {
				fmt.Println("nil")
			} else {
				x = val.(*int)
				fmt.Println(x)
			}

		default:
			fmt.Println(v)
		}

		_ = val
		_ = x

	}

}

func TestValueFor(t *testing.T) {

	fb := NewFilter(
		[]Pair{
			{Column: "first_name", Value: Value{Src: "Zaldy", Raw: true}},
			{Column: "last_name", Value: Value{Src: "Baguinon", Raw: true}},
			{Column: "age", Value: Value{Src: 47, Raw: true}},
			{Column: "title", Value: Value{Src: nil, Raw: true}},
		}, true, "@p")

	fb.Ne = append(fb.Ne, Pair{Column: "first_name", Value: Value{Src: "James", Raw: true}})
	fb.Ne = append(fb.Ne, Pair{Column: "last_name", Value: Value{Src: "Lumibao", Raw: true}})

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

	fb := NewFilter(
		[]Pair{
			{Column: "first_name", Value: Value{Src: "Zaldy", Raw: true}},
			{Column: "last_name", Value: Value{Src: "Baguinon", Raw: true}},
			{Column: "age", Value: Value{Src: &age, Raw: true}},
		}, true, "@p")

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
}
