package main

import (
	"errors"
	"fmt"
	"math"
)

var ErrIndexOutOfRange = errors.New("下标超出范围")

// DeleteAt 删除指定位置的元素
// 如果下标不是合法的下标，返回 ErrIndexOutOfRange
func DeleteAt(s []int, idx int) ([]int, error) {
	if idx < 0 || idx >= len(s) {
		return nil, ErrIndexOutOfRange
	}
	newArrInt := make([]int, len(s)-1)
	i := 0

	for ; i < idx; i++ {
		newArrInt[i] = s[i]
	}
	for i = idx; i < len(s)-1; i++ {
		newArrInt[i] = s[i+1]
	}

	return newArrInt, nil
}

// 优化版本
func DeleteAt2(s []int, idx int) ([]int, error) {
	if idx < 0 || idx >= len(s) {
		return nil, ErrIndexOutOfRange
	}
	len := len(s)
	for i := idx; i < len-1; i++ {
		s[i] = s[i+1]
	}
	return s[:len-1], nil
}

// 泛型
func DeleteAt3[T any](s []T, idx int) ([]T, error) {
	if idx < 0 || idx >= len(s) {
		return nil, ErrIndexOutOfRange
	}
	len := len(s)
	for i := idx; i < len-1; i++ {
		s[i] = s[i+1]
	}
	return s[:len-1], nil

}

// 加上缩容机制，就是当实际长度较之前的容量低于1/2时，重新分配容量为之1/2的内存空间，并返回新切片
func DeleteAt4[T any](s []T, idx int) ([]T, error) {
	length := len(s)
	if idx < 0 || idx >= length {
		return nil, ErrIndexOutOfRange
	}

	delfunc := func(src []T, dest []T, idx int) ([]T, error) {
		j := 0
		for i, val := range src {
			if i != idx {
				dest[j] = val
				j++
			}
		}
		dest = dest[:j]
		return dest, nil
	}
	orgCap := cap(s)
	//fmt.Printf("len=%d,cap=%d,comp:%.2f\n", length-1, orgCap, float64(length-1)/float64(orgCap))
	if orgCap < 64 {
		return delfunc(s, s, idx)
	} else if float64(length-1)/float64(orgCap) < 0.5 {
		fmt.Println("缩容了")
		newcap := int(math.Ceil(float64(orgCap) / 2))
		newS := make([]T, newcap, newcap)
		return delfunc(s, newS, idx)
	} else {
		return delfunc(s, s, idx)
	}

}

func main() {
	arrint := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	//删除第5个位置
	arrint, err := DeleteAt(arrint, 5)
	if err != nil {
		println("delete data error !")
		return
	}
	for _, a := range arrint {
		fmt.Println(a)
	}

	//删除第1个位置
	arrint, err = DeleteAt2(arrint, 0)
	if err != nil {
		println("delete data error !")
		return
	}
	fmt.Printf("DeleteAt2:\n")
	for _, a := range arrint {
		fmt.Println(a)
	}

	//删除最后一个位置
	arrint, err = DeleteAt3[int](arrint, 7)
	if err != nil {
		println("delete data error !")
		return
	}
	for _, a := range arrint {
		fmt.Println(a)
	}

	arrtest := make([]int, 0, 100)
	for i := 0; i < 100; i++ {
		arrtest = append(arrtest, i)
	}
	//var err error
	for i := 0; i < 60; i++ {
		arrtest, err = DeleteAt4[int](arrtest, 1)
		if err != nil {
			fmt.Println("删除失败")
			return
		}
		fmt.Printf("删除后切片长度为:%d\n", len(arrtest))
		fmt.Printf("删除后切片容量为:%d\n", cap(arrtest))
	}

}
