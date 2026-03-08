package main

import (
	"fmt"
	"unsafe"

	"github.com/timandy/routine"
)

func main() {
	onStack := checkStackAddr()
	fmt.Printf("Slice backing store on stack: %v\n", onStack)
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

	// Alternatives in slice allocation. Both keep allocation on stack,
	// all things being equal.
	// x := make([]byte, 64)
	var x []byte
	x = append(x, 'a') // allocation happens here
	x = append(x, 'b')

	/* These all cause heap escapes:
	// fmt.Printf("backing store at %p\n", unsafe.SliceData(x))
	// fmt.Printf("backing store at %x\n", unsafe.SliceData(x))
	// fmt.Printf("backing store at %p\n", unsafe.Pointer(unsafe.SliceData(x)))
	// fmt.Printf("backing store at %p\n", &(x[0]))
	// fmt.Printf("Slice %p\n", x)
	// fmt.Printf("Slice %p\n", &x)
	*/

	// Alternatives in getting backing store address: both allow
	// stack allocations, all other things being equal.
	backingStore := uintptr(unsafe.Pointer(&(x[0])))
	// backingStore := uintptr(unsafe.Pointer(unsafe.SliceData(x)))

	// The lesson here is that in Go v1.26, escape analysis stops
	// when the unitptr() type conversion happens.

	fmt.Printf("stack top     %016x\n", myg.stack.hi)
	/* this fools escape analysis the hard way
	b1 := byte(0xff & (backingStore >> 56))
	b2 := byte(0xff & (backingStore >> 48))
	b3 := byte(0xff & (backingStore >> 40))
	b4 := byte(0xff & (backingStore >> 32))
	b5 := byte(0xff & (backingStore >> 24))
	b6 := byte(0xff & (backingStore >> 16))
	b7 := byte(0xff & (backingStore >> 8))
	b8 := byte(0xff & (backingStore >> 0))
	fmt.Printf("backing store %02x%02x%02x%02x", b1, b2, b3, b4)
	fmt.Printf("%02x%02x%02x%02x\n", b5, b6, b7, b8)
	*/
	fmt.Printf("backing store %016x\n", uintptr(unsafe.Pointer(unsafe.SliceData(x))))
	fmt.Printf("stack bottom  %016x\n", myg.stack.lo)

	if myg.stack.hi > backingStore && backingStore > myg.stack.lo {
		return true
	}

	return false
}
