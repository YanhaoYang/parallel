package parallel

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"testing"
	"time"
)

func muteLogger() *log.Logger {
	return log.New(ioutil.Discard, "", log.LUTC|log.Ldate|log.Ltime|log.Lshortfile)
}

func TestRunnerCallsProduceAndConsume(t *testing.T) {
	allProducts := make(map[string]struct{})
	allResults := make(map[string]struct{})
	mutex := &sync.Mutex{}

	r := New()
	r.Logger = muteLogger()
	r.Produce = func(products chan<- interface{}, done <-chan struct{}) {
		gid := getGID()
		for i := 0; i < 3; i++ {
			msg := fmt.Sprintf("%02d - %02d", gid, i)
			select {
			case products <- msg:
				mutex.Lock()
				allProducts[msg] = struct{}{}
				mutex.Unlock()
			case <-done:
				return
			}
		}
	}

	r.Consume = func(product interface{}, done <-chan struct{}) {
		mutex.Lock()
		allResults[product.(string)] = struct{}{}
		mutex.Unlock()
	}
	r.NumProducers = 3
	r.Run()

	if len(allProducts) != 9 {
		t.Error(
			"For", "len(allProducts)",
			"expected", 9,
			"got", len(allProducts),
		)
	}
	if len(allResults) != 9 {
		t.Error(
			"For", "len(allResults)",
			"expected", 9,
			"got", len(allResults),
		)
	}
}

func TestRunnerCanBeInterrupted(t *testing.T) {
	var cancellations []uint64
	mutex := &sync.Mutex{}
	r := New()
	r.Logger = muteLogger()
	r.Produce = func(products chan<- interface{}, done <-chan struct{}) {
		gid := getGID()
		for i := 0; i < 90; i++ {
			msg := fmt.Sprintf("%02d - %02d", gid, i)
			select {
			case products <- msg:
			case <-done:
				mutex.Lock()
				cancellations = append(cancellations, gid)
				mutex.Unlock()
				return
			}
		}
	}

	r.Consume = func(product interface{}, done <-chan struct{}) {
		gid := getGID()
		select {
		case <-done:
			mutex.Lock()
			cancellations = append(cancellations, gid)
			mutex.Unlock()
			return
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
	r.NumProducers = 3
	r.ConsumeAll = false
	TrapInterrupt(r)

	go func() {
		time.Sleep(250 * time.Millisecond)
		p, _ := os.FindProcess(os.Getpid())
		_ = p.Signal(os.Interrupt)
	}()

	r.Run()

	if len(cancellations) != 3 {
		t.Error(
			"For", "len(cancellations)",
			"expected", 3,
			"got", len(cancellations),
		)
	}
}

func TestRunnerCanBeStopped(t *testing.T) {
	var cancellations []uint64
	mutex := &sync.Mutex{}
	r := New()
	r.Logger = muteLogger()
	r.Produce = func(products chan<- interface{}, done <-chan struct{}) {
		gid := getGID()
		for i := 0; i < 90; i++ {
			msg := fmt.Sprintf("%02d - %02d", gid, i)
			select {
			case products <- msg:
			case <-done:
				mutex.Lock()
				cancellations = append(cancellations, gid)
				mutex.Unlock()
				return
			}
		}
	}

	r.Consume = func(product interface{}, done <-chan struct{}) {
		gid := getGID()
		select {
		case <-done:
			mutex.Lock()
			cancellations = append(cancellations, gid)
			mutex.Unlock()
			return
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
	r.NumProducers = 3
	r.NumConsumers = 3
	r.ConsumeAll = false
	TrapInterrupt(r)

	go func() {
		time.Sleep(250 * time.Millisecond)
		r.Stop()
	}()

	r.Run()

	if len(cancellations) != 3 {
		t.Error(
			"For", "len(cancellations)",
			"expected", 3,
			"got", len(cancellations),
		)
	}
}
