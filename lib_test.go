package parsing

import (
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
