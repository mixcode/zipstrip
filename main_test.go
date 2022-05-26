package main

import (
	//"fmt"
	"os"
	"testing"
)

func testz(filename string) (sz int64, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return
	}
	defer f.Close()
	return zipLength(f, 4*1024*1024)
}

func TestZIP(t *testing.T) {
	const fname = "testdata/z32.zip"
	const zsize = 109
	sz, err := testz(fname)
	if err != nil {
		t.Fatal(err)
	}
	if sz != zsize {
		t.Fatalf("zip size does not match; expected %d, result %d", zsize, sz)
	}
	//fmt.Printf("len: %d (0x%x)\n", sz, sz)
}

func TestZIP64(t *testing.T) {
	const fname = "testdata/z64.zip"
	const zsize = 337
	sz, err := testz(fname)
	if err != nil {
		t.Fatal(err)
	}
	if sz != zsize {
		t.Fatalf("zip size does not match; expected %d, result %d", zsize, sz)
	}
	//fmt.Printf("len: %d (0x%x)\n", sz, sz)
}
