package main

import (
	"fmt"
	"github.com/zhangxianweihebei/gostl/algorithm/sort"
	"github.com/zhangxianweihebei/gostl/ds/slice"
	"github.com/zhangxianweihebei/gostl/utils/comparator"
)

func main() {
	a := slice.IntSlice(make([]int, 0))
	a = append(a, 2)
	a = append(a, 1)
	a = append(a, 3)
	fmt.Printf("%v\n", a)

	// sort in ascending
	sort.Sort(a.Begin(), a.End())
	fmt.Printf("%v\n", a)

	// sort in descending
	sort.Sort(a.Begin(), a.End(), comparator.Reverse(comparator.BuiltinTypeComparator))
	fmt.Printf("%v\n", a)
}
