package bytesutil

import (
	"reflect"
	"unsafe"
)

// Resize resizes b to n bytes and returns b (which may be newly allocated).
// []byte数组,扩容到指定大小,1Byte = 8bit
func Resize(b []byte, n int) []byte {
	if nn := n - cap(b); nn > 0 {
		// 使用make函数扩容nn个空间，填充进去
		b = append(b[:cap(b)], make([]byte, nn)...)
	}
	return b[:n]
}

// ToUnsafeString converts b to string without memory allocations.
//
// The returned string is valid only until b is reachable and unmodified.
func ToUnsafeString(b []byte) string {
	// 只需要指针转一下就行，因为指针的第一个地址是SliceHeader中的data的位置
	return *(*string)(unsafe.Pointer(&b))
}

// ToUnsafeBytes converts s to a byte slice without memory allocations.
//
// The returned byte slice is valid only until s is reachable and unmodified.
// 字符串转成成为[]byte，没有内存的分配,返回的字节数组仅仅只在's'没被使用还是修改时候使用
func ToUnsafeBytes(s string) []byte {
	// unsafe.Pointer(&s),把字符串的指针地址换成可以转换类型的,也就是转换成一种弱类型的指针;reflect.StringHeader是运行时字符串的表达形式,主要是一个指向数据的指针和它的长度
	// 把字符串转成reflect.StringHeader,那么就可以零内存分配，偷到数据指针，转到Slice了
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	var slh reflect.SliceHeader
	slh.Data = sh.Data // 指向数据的指针
	slh.Len = sh.Len   // 数据的长度
	slh.Cap = sh.Len
	return *(*[]byte)(unsafe.Pointer(&slh)) // 把数据换成byte指针，然后再继续换成数据
}
