package main

var ch = make(chan int, 1)

var s string

func f() {
	s = "hello, world"
	// <-ch
	close(ch)
}

func main() {
	go f()
	ch <- 1
	print(s)
}
