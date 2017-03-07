package main

import "fmt"

func printSlice(x []int) {
	fmt.Printf("len=%d cap=%d slice=%v\n", len(x), cap(x), x)
	for key, value := range x {
		fmt.Print(key, value)
	}
}

func main() {
	var a []int
	a = append(a, 1,2,3,4, 5)
	printSlice(a)
}
