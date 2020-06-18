package main

import (
	"fmt"
	"unsafe"
)

func main() {
	var str string = "hello"
	addr := &str
	fmt.Printf("old addr:%v\n", addr)

	up := unsafe.Pointer(addr) // 把指针转换了
	fmt.Printf("new addr:%d\n", uintptr(up))
}
