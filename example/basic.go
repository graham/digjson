package main

import (
	"fmt"

	"github.com/graham/digjson"
)

func main() {
	fmt.Println("hello world")

	jsondata_with_nil := []byte(`{ "user": null }`)
	jsondata_with_obj := []byte(`{ "user": { "username": "user1" } }`)
	jsondata_with_array := []byte(`{"user": [{ "username": "user1"}, {"username": "user2"} ] }`)

	type User struct {
		Username string `json:"username"`
	}

	var users []User
	was_found, err := digjson.Dig(jsondata_with_obj, "user", &users)

	fmt.Println("list should have one element", users, was_found, err)

	users = []User{}
	was_found, err = digjson.Dig(jsondata_with_array, "user", &users)

	fmt.Println("list should have multiple elements", users, was_found, err)

	users = []User{}
	was_found, err = digjson.Dig(jsondata_with_nil, "user", &users)

	fmt.Println("list should be empty", users, was_found, err)
}
