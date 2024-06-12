package gotest

import (
	"fmt"
	"time"
)

func Add(a int, b int) int {
	return a + b
}

func SlowFunc() int {
	fmt.Printf("Before sleep")
	time.Sleep(time.Second)
	fmt.Printf("After sleep")
	return 0
}
