package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"unsafe"

	"github.com/timandy/routine"
)

func main() {
	printHeapAddress()
	printBackingStore()
	onStack := checkStackAddr()
	fmt.Printf("Slice backing store on stack: %v\n", onStack)
}

type sliceHeader struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}

func backingStore(x []byte) unsafe.Pointer {
	u := unsafe.Pointer(&x)
	sh := (*sliceHeader)(u)
	return sh.Data
}

// printBackingStore demonstrates that &(slice[0]) is the
// address of a slice's backing store
func printBackingStore() {
	x := make([]byte, 64)
	bs := backingStore(x)
	fmt.Println("slice backing store addressing check")
	fmt.Printf("backing store at %p\n", bs)
	fmt.Printf("first element at %p\n\n", &(x[0]))
}

// printHeapAddress reads /proc/$PID/maps to find system stack and system heap.
func printHeapAddress() {
	maps, err := os.ReadFile(fmt.Sprintf("/proc/%d/maps", os.Getpid()))
	if err != nil {
		log.Fatal(err)
	}
	for _, line := range bytes.Split(maps, []byte{'\n'}) {
		if bytes.Contains(line, []byte("[anon: Go: heap]")) {
			fmt.Print("system heap:  ")
			fields := bytes.Split(line, []byte{' '})
			os.Stdout.Write(fields[0])
			os.Stdout.Write([]byte{'\n'})
			continue
		}
		if bytes.Contains(line, []byte("[stack]")) {
			fmt.Print("system stack: ")
			fields := bytes.Split(line, []byte{' '})
			os.Stdout.Write(fields[0])
			os.Stdout.Write([]byte{'\n', '\n'})
			break
		}
	}
}

// type stack from /usr/lib/go/src/runtime/runtime2.go
type stack struct {
	lo uintptr
	hi uintptr
}

// part of type g from /usr/lib/go/src/runtime/runtime2.go
type g struct {
	stack       stack
	stackguard0 uintptr
	stackguard1 uintptr
}

func checkStackAddr() bool {
	myg := (*g)(routine.Getg())

	x := make([]byte, 64)
	backingStore := uintptr(unsafe.Pointer(&(x[0])))

	if myg.stack.hi > backingStore && backingStore > myg.stack.lo {
		return true
	}

	return false
}
