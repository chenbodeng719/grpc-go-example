package main

import (
	"sync"
	"testing"
)

//func test(x int) {
//	defer println("a")
//	defer println("b")
//	defer func() {
//		println(100 / x) // div0 异常未被捕获，逐步往外传递，最终终⽌进程。
//	}()
//	defer println("c")
//}
var lock sync.Mutex

func test() {
	lock.Lock()
	lock.Unlock()
}
func testdefer() {
	lock.Lock()
	defer lock.Unlock()
}
func BenchmarkTest(b *testing.B) {
	for i := 0; i < b.N; i++ {
		test()
	}
}
func BenchmarkTestDefer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		testdefer()
	}
}
