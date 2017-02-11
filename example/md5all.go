package main

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"bitbucket.org/parallel"
)

func walkFiles(paths chan<- interface{}, done <-chan struct{}, errc chan<- error, root string) {
	errc <- filepath.Walk(root, func(path string, info os.FileInfo, err error) error { // HL
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		select {
		case paths <- path:
		case <-done:
			return errors.New("walk canceled")
		}
		return nil
	})
}

type result struct {
	path string
	sum  [md5.Size]byte
	err  error
}

func digester(path interface{}) result {
	data, err := ioutil.ReadFile(path.(string))
	return result{path.(string), md5.Sum(data), err}
}

func main() {
	var results []result
	errc := make(chan error, 1)

	r := parallel.New()
	r.Produce = func(products chan<- interface{}, done <-chan struct{}) {
		walkFiles(products, done, errc, ".")
	}

	r.Consume = func(product interface{}, done <-chan struct{}) {
		results = append(results, digester(product))
	}
	parallel.TrapInterrupt(r)

	r.Run()

	m := make(map[string][md5.Size]byte)
	for _, res := range results {
		if res.err != nil {
			continue
		}
		m[res.path] = res.sum
	}

	var paths []string
	for path := range m {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		fmt.Printf("%x  %s\n", m[path], path)
	}
}
