# digjson
Handling shapechanging json in go.

## Purpose
There quite a few great libraries for Go when handling JSON, however, handling json that might change shape isn't something those packages handle well. 

## What do you mean by `json that changes shape`
Working with APIs that emit JSON is usually easy, however, in some cases, one request might result in a dictionary, another an array, another a nil value. This is very hard to deal with in a typed language like Go. `digjson` attempts to make working with these "changing shapes" a bit easier.

## Goals
Turn a json path into a reasonable value that you can work with.

## Non-Goals
Performance beyond the standard libraries json.decode (I'll make it as fast as I can with the feature set).

## Is this actually a problem?
For years, I never saw anything like this, but in the course of the last 6 months I've run into two APIs (one public and one private) that have this behavior. There are some other packages that solve some of the issues, but I wanted something that fit my use case exactly.

## I still don't understand, can you give me an example

__Absolutely__, the tests show some examples, but lets use something simple.

To start, the JSON looks like this:

```
jsondata = { "user": { "username: "user1" } }
```

```
var u []string
was_found, err := Dig(jsondata, "user.username", &u)
```

The value "user1" is now in `u`.

`Dig` also works with structs:


```
jsondata = { "user": { "username: "user1" } }
```

```
type User struct {
    Username string `json:"username"`
}

var u User
was_found, err := Dig(jsondata, "user", &u)
```

So far, this isn't doing much for you compared to the normal JsonDecoder other than some nice json path access, lets look at three examples I've run into in the real world.


```
jsondata_with_obj = { "user": { "username: "user1" } }
jsondata_with_nil = { "user": null }
jsondata_with_array = { "user": [{ "username: "user1"}, {"username": "user2"} ] }

```

This is very hard to deal with, because during the decode, you don't know what the type will be, You can choose to decode to `map[string]interface{}` but this is very cumbersome with large Json Documents.

Dig assumes you want the datatype you're giving it, if it's a list, and there is only a object, it will wrap it in a list (one item), if there is a null, you'll get back a empty list. It's easier if i just show you.

```
type User struct {
    Username string `json:"username"`
}

var users []User
was_found, err := Dig(jsondata_with_obj, "user", &users)

// users == [ User{ Username: "user1" } ]

was_found, err := Dig(jsondata_with_nil, "user", &users)

// users == []

was_found, err := Dig(jsondata_with_array, "user", &users)


// users == [ User{ Username: "user1" }, User{ Username: "user2"} ]

```

## This seems super terrible
Yes

## Why would anyone ever build an API like this?
I believe the instances where I'm experiencing this are related to a platform that returns XML and someone has written a XML to JSON converter. In these cases, it's ... `understandable` that things work the way they do, but it's still very painful to consume the data.

## Can I change the separator to '.' if it doesn't work for me?
Not yet, but that is a good idea. (PR welcome)

## Does it implicitly convert "123" (string) to 123 (int) if I give it a struct field with the type int.
Yes, there is also DigStrict that will not do that and return an error. I recommend checking out the tests for a clearer description here.

## Are there tests?
Some, but I'm sure we could write more (PR Welcome).

## Is it obvious how this works?
GoLang has a reflect module that is very powerful, but it is pretty weird to use if you're not familiar with how Go works, I encourage you to read the code!

## Why are there so many questions in this README?
Good point :)
