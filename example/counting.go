package main

import (
	"bytes"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/parallel"
)

func getGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

func main() {
	r := parallel.New()
	r.Produce = func(products chan<- interface{}, done <-chan struct{}) {
		gid := getGID()
		for i := 0; i < 99; i++ {
			msg := fmt.Sprintf("%02d - %02d", gid, i)
			select {
			case products <- msg:
				fmt.Printf("Produce - %02d: %s\n", gid, msg)
				time.Sleep(100 * time.Millisecond)
			case <-done:
				return
			}
		}
	}

	r.Consume = func(product interface{}, done <-chan struct{}) {
		gid := getGID()
		fmt.Printf("Consume - %02d: %s %s\n", gid, strings.Repeat(" ", 7), product.(string))
		time.Sleep(800 * time.Millisecond)
	}
	r.NumProducers = 2
	parallel.TrapInterrupt(r)
	go func() {
		time.Sleep(4 * time.Second)
		r.Stop()
	}()

	r.Run()
}
