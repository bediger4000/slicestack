package main

import (
	"fmt"
	"unsafe"

	"github.com/timandy/routine"
)

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

var myg *g

func main() {
	myg = (*g)(routine.Getg())
	startRecursion()
}

func startRecursion() {
	var localVar int
	localPointer := &localVar
	localUintptr := uintptr(unsafe.Pointer(localPointer))
	currentHi := myg.stack.hi
	aSlice := make([]byte, 128)
	fmt.Printf("local pointer       %016x\n", uintptr(unsafe.Pointer(&localPointer)))
	fmt.Printf("local pointer val   %016x\n", uintptr(unsafe.Pointer(localPointer)))
	fmt.Printf("local uintptr val   %016x\n", localUintptr)
	fmt.Printf("stack top           %016x\n", currentHi)
	fmt.Printf("local variable      %016x\n", uintptr(unsafe.Pointer(&localVar)))
	fmt.Printf("backing store       %016x\n", uintptr(unsafe.Pointer(&(aSlice[0]))))
	fmt.Printf("stack bottom        %016x\n", myg.stack.lo)
	fmt.Printf("stack size          %016x\n\n", myg.stack.hi-myg.stack.lo)
	enlargeStack(currentHi)
	fmt.Printf("stack top 2         %016x\n", myg.stack.hi)
	fmt.Printf("local variable 2    %016x\n", uintptr(unsafe.Pointer(&localVar)))
	fmt.Printf("backing store 2     %016x\n", uintptr(unsafe.Pointer(&(aSlice[0]))))
	fmt.Printf("stack bottom 2      %016x\n", myg.stack.lo)
	fmt.Printf("stack size 2        %016x\n", myg.stack.hi-myg.stack.lo)
	fmt.Printf("local pointer 2     %016x\n", uintptr(unsafe.Pointer(&localPointer)))
	fmt.Printf("local pointer val 2 %016x\n", uintptr(unsafe.Pointer(localPointer)))
	fmt.Printf("local uintptr val 2 %016x\n", localUintptr)
}

// enlargeStack recurses until formal argument has zero value,
// or the top-of-stack address changes
func enlargeStack(oldHi uintptr) {
	currentHi := myg.stack.hi
	if currentHi != oldHi {
		fmt.Printf("top was           %016x\n", oldHi)
		fmt.Printf("stack top         %016x has changed\n", currentHi)
		fmt.Printf("formal argument   %016x\n", uintptr(unsafe.Pointer(&oldHi)))
		fmt.Printf("local variable    %016x\n", uintptr(unsafe.Pointer(&currentHi)))
		fmt.Printf("stack bottom      %016x\n", myg.stack.lo)
		fmt.Printf("stack size        %016x\n\n", currentHi-myg.stack.lo)
		return
	}
	enlargeStack(currentHi)
}
