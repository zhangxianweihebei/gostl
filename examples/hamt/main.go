package main

import (
	"fmt"
	"github.com/zhangxianweihebei/gostl/ds/hamt"
)

func main() {
	h := hamt.New(hamt.WithGoroutineSafe())
	key := []byte("aaaaa")
	val := "bbbbbbbbbbbbb"

	h.Insert(key, val)
	fmt.Printf("%v = %v\n", string(key), h.Get(key))

	h.Erase(key)
	fmt.Printf("%v = %v\n", string(key), h.Get(key))
}
