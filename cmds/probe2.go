package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
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
	n, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	myg = (*g)(routine.Getg())
	startRecursion(n)
}

func startRecursion(ply int) {
	currentHi := myg.stack.hi
	fmt.Printf("stack top         %016x\n", currentHi)
	fmt.Printf("stack size        %016x\n\n", currentHi-myg.stack.lo)
	probeStack(ply-1, currentHi)
}

// probeStack recurses until formal argument has zero value
func probeStack(ply int, oldHi uintptr) {
	currentHi := myg.stack.hi
	if currentHi != oldHi {
		fmt.Printf("ply % 4d top was  %016x\n", ply, oldHi)
		fmt.Printf("stack top         %016x has changed\n", currentHi)
		fmt.Printf("stack size        %016x\n", currentHi-myg.stack.lo)
		fmt.Printf("formal argument   %016x\n", uintptr(unsafe.Pointer(&ply)))
		fmt.Printf("local variable    %016x\n", uintptr(unsafe.Pointer(&currentHi)))
		fmt.Printf("stack bottom      %016x\n\n", myg.stack.lo)
	}
	if ply > 0 {
		probeStack(ply-1, currentHi)
	}
}
