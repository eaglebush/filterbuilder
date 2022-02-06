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
