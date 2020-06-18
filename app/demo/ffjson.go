package main

import (
	"fmt"
	"github.com/valyala/fastjson"
)

func main() {
	s := []byte(`{"foo": [123, "bar"]}`)

	fv, err := fastjson.ParseBytes(s)
	if err != nil {
		fmt.Println("Parse error")
	}

	v := fv.Get("foo")
	array, err := v.Array()
	if err != nil {
		fmt.Println("Not a array")
		return
	}

	for _, item := range array {
		fmt.Println(item.String())
	}
}
