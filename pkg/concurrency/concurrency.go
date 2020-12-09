package concurrency

import "fmt"

const (
	// DefaultMax default max
	DefaultMax = 256
)

// GoLimit go limit
type GoLimit struct {
	ch chan int
}

// NewGoLimit new go limit
func NewGoLimit(max int) *GoLimit {
	return &GoLimit{
		ch: make(chan int, max),
	}
}

// NewDefaultGoLimit new default go limit
func NewDefaultGoLimit() *GoLimit {
	return NewGoLimit(DefaultMax)
}

// Add add num
func (g *GoLimit) Add() {
	g.ch <- 1
}

// Done remove num
func (g *GoLimit) Done() {
	<-g.ch
}

// Close close chan
func (g *GoLimit) Close() {
	defer func() {
		if recover() != nil {
			fmt.Println("golimit closed")
		}
	}()
	close(g.ch)
}

// Async 异步,并发，不阻塞，任务未执行完函数就结束
func Async() {

}

// AsyncWithLimit 异步,并发，不阻塞，任务未执行完函数就结束
func AsyncWithLimit(limit *GoLimit) {

}

// AsyncWithDefaultLimit 异步,并发，不阻塞，任务未执行完函数就结束
func AsyncWithDefaultLimit() {

}

// Await 阻塞，同步，任务执行完函数才结束
func Await() {

}

// AwaitWithLimit 阻塞，同步，任务执行完函数才结束
func AwaitWithLimit() {

}

// AwaitWithDefaultLimit 阻塞，同步，任务执行完函数才结束
func AwaitWithDefaultLimit() {

}
