package digjson

import (
	"reflect"
	"testing"
)

func TestDigBasicString(t *testing.T) {
	json_with_nested := []byte(`{"a": {"b": {"c": "d"}}}`)

	var f string
	was_found, err := Dig(json_with_nested, "a.b.c", &f)

	if err != nil {
		panic(err)
	}

	if was_found != true {
		t.Errorf("Value should have been found, it wasn't.")
	}

	if f != "d" {
		t.Errorf("Value should have been `d`, it was `%s`.", f)
	}
}

func TestDigBasicInt(t *testing.T) {
	json_with_nested := []byte(`{"a": {"b": {"c": 12345}}}`)

	var f int
	was_found, err := Dig(json_with_nested, "a.b.c", &f)

	if err != nil {
		panic(err)
	}

	if was_found != true {
		t.Errorf("Value should have been found, it wasn't.")
	}

	if f != 12345 {
		t.Errorf("Value should have been `d`, it was `%d`.", f)
	}
}

func TestDigStruct(t *testing.T) {
	type User struct {
		Name   string `json:"name"`
		Active bool   `json:"active"`
		Count  int    `json:"count"`
	}

	json_with_struct := []byte(`{"a": {"b": {"name":"graham", "active": true, "count": 1000}}}`)

	var u User
	was_found, err := Dig(json_with_struct, "a.b", &u)

	if err != nil {
		panic(err)
	}

	if was_found != true {
		t.Errorf("Value should have been found, it wasn't.")
	}

	if u.Name != "graham" {
		t.Errorf("String value not copied correctly.")
	}

	if u.Active != true {
		t.Errorf("Bool value not copied correctly.")
	}

	if u.Count != 1000 {
		t.Errorf("Int value not copied correctly.")
	}
}

func TestDigSliceOfStructs(t *testing.T) {
	type User struct {
		Name   string `json:"name"`
		Active bool   `json:"active"`
		Count  int    `json:"count"`
	}

	json_with_slice := []byte(`{"a": {"b": [
{"name":"graham", "active": true, "count": 1000},
{"name":"grimes", "active": true, "count": 9000},
{"name":"bowie", "active": false, "count": -200}
]}}`)

	var u []User
	was_found, err := Dig(json_with_slice, "a.b", &u)

	if err != nil {
		panic(err)
	}

	if was_found != true {
		t.Errorf("Value should have been found, it wasn't.")
	}

	if len(u) != 3 {
		t.Errorf("Slice should be len 3 but is `%d`", len(u))
	}
}

func TestDig_ExpectSlice_GetNil(t *testing.T) {
	type User struct {
		Name   string `json:"name"`
		Active bool   `json:"active"`
		Count  int    `json:"count"`
	}

	json_with_slice := []byte(`{"a": {"b": null}}`)

	var u []User
	was_found, err := Dig(json_with_slice, "a.b", &u)

	if err != nil {
		panic(err)
	}

	if was_found != true {
		t.Errorf("Value should have been found, it wasn't.")
	}

	if len(u) != 0 {
		t.Errorf("Slice should be len 0 but is `%d`", len(u))
	}
}

func TestDig_ConvertFromStringToInt(t *testing.T) {
	json_with_string := []byte(`{"a": "1000"}`)

	var value int
	was_found, err := Dig(json_with_string, "a", &value)

	if err != nil {
		panic(err)
	}

	if was_found != true {
		t.Errorf("Value should have been found, it wasn't.")
	}

	if value != 1000 {
		t.Errorf("Value should have been 1000, it wasn't.")
	}
}

func TestDigStrict_ConvertFromStringToIntFails(t *testing.T) {
	json_with_string := []byte(`{"a": "1000"}`)

	var value int
	was_found, err := DigStrict(json_with_string, "a", &value)

	if err == nil {
		t.Errorf("This should have returned an error, it didn't")
	}

	if was_found != true {
		t.Errorf("Value should have been found, it wasn't.")
	}

}

func TestDig_MissingField(t *testing.T) {
	json_with_string := []byte(`{"a": "1000"}`)

	var value int
	was_found, err := DigStrict(json_with_string, "a.b.c.d", &value)

	if err != nil {
		t.Errorf("This shouldn't have returned an error, but it did.")
	}

	if was_found == true {
		t.Errorf("Value shouldn't have been found, it was.")
	}

}

func TestTransmuteValue_StringToInt(t *testing.T) {
	var value int
	target := reflect.ValueOf(&value)
	indr := reflect.Indirect(target)

	err := transmute_value_to_target(reflect.String, reflect.Int, "1234", &indr)

	if err != nil {
		t.Errorf("This shouldn't return an error")
	}

	if value != 1234 {
		t.Errorf("Value was not converted correctly.")
	}
}

func TestTransmuteValue_IntToString(t *testing.T) {
	var value string
	target := reflect.ValueOf(&value)
	indr := reflect.Indirect(target)

	var source int = 1234

	err := transmute_value_to_target(reflect.Int, reflect.String, source, &indr)

	if err != nil {
		t.Errorf("This shouldn't return an error %s", err)
	}

	if value != "1234" {
		t.Errorf("Value was not converted correctly. %s", value)
	}
}

func TestTransmuteValue_NotSupported(t *testing.T) {
	var value int
	target := reflect.ValueOf(&value)
	indr := reflect.Indirect(target)

	var source float64 = 1234.1234

	err := transmute_value_to_target(reflect.Float64, reflect.Int, source, &indr)

	if err == nil {
		t.Errorf("This didn't return an error, but it did")
	}

}
