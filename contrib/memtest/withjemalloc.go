//+build jemalloc

package main

import (
	"os"

	"github.com/outcaste-io/ristretto/z"
)

func Calloc(size int) []byte { return z.Calloc(size, "memtest") }
func Free(bs []byte)         { z.Free(bs) }
func NumAllocBytes() int64   { return z.NumAllocBytes() }

func check() {
	if buf := z.CallocNoRef(1, "memtest"); len(buf) == 0 {
		panic("Not using manual memory management. Compile with jemalloc.")
		os.Exit(1)
	} else {
		z.Free(buf)
	}

	z.StatsPrint()
}
