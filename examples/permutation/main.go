package main

import (
	"fmt"
	"github.com/zhangxianweihebei/gostl/algorithm/sort"
	"github.com/zhangxianweihebei/gostl/ds/slice"
	"github.com/zhangxianweihebei/gostl/utils/comparator"
)

func main() {
	a := slice.IntSlice(make([]int, 0))

	for i := 1; i <= 3; i++ {
		a = append(a, i)
	}
	fmt.Println("NextPermutation")
	for {
		fmt.Printf("%v\n", a)
		if !sort.NextPermutation(a.Begin(), a.End()) {
			break
		}
	}
	fmt.Println("PrePermutation")
	for {
		fmt.Printf("%v\n", a)
		if !sort.NextPermutation(a.Begin(), a.End(), comparator.Reverse(comparator.BuiltinTypeComparator)) {
			break
		}
	}
}
