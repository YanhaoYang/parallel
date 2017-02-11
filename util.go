package parallel

import (
	"bytes"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
)

// GetEnvInt reads an environment variable and returns it as int
// If the environment variable is empty, returns the defaultValue
func GetEnvInt(name string, defaultValue int) int {
	env := os.Getenv(name)
	if env == "" {
		return defaultValue
	}
	v, _ := strconv.Atoi(env)
	return v
}

type stoppable interface {
	Stop()
}

// TrapInterrupt traps os.Interrupt signal and closes done channel
func TrapInterrupt(s stoppable) {
	go func(s stoppable) {
		signalCh := make(chan os.Signal, 1)
		defer close(signalCh)

		signal.Notify(signalCh, os.Interrupt)

		for {
			sig := <-signalCh
			switch sig {
			case syscall.SIGINT:
				s.Stop()
			default:
			}
		}
	}(s)
}

func getGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}
