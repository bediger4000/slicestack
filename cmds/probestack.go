package main

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"os"
	"unsafe"

	"github.com/timandy/routine"
)

/*
 * Call a recursive function many times to figure out
 * addresses on the goroutine's stack, before seeing
 * if a slice gets allocated on the stack or out of the heap.
 *
 * Fails to account for variables that escape to heap, so
 * everything gets allocated from the heap.
 *
 * Also gets type g, type m, structs
 * from the Go runtime.
 */

func main() {

	printHeapAddress()
	probeStack(os.Getpid(), rand.Intn(5)+20)
	bsAt2 := doSlice2(42)
	fmt.Printf("slice allocation at %p\n", bsAt2)

	printG()
}

func probeStack(x int, n int) {
	fmt.Printf("formal args at %p\t%p\n", &x, &n)
	if n == 0 {
		return
	}
	a := make([]int, x)
	for i := range a {
		a[i] = n
	}
	probeStack(rand.Intn(9), n-1)
}

type sliceHeader struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}

func backingStore(x []byte) unsafe.Pointer {
	u := unsafe.Pointer(&x)
	sh := (*sliceHeader)(u)
	i := (*byte)(sh.Data)
	*i = 'Z'
	return sh.Data
}

func printHeapAddress() {
	maps, err := os.ReadFile(fmt.Sprintf("/proc/%d/maps", os.Getpid()))
	if err != nil {
		log.Fatal(err)
	}
	for _, line := range bytes.Split(maps, []byte{'\n'}) {
		if bytes.Contains(line, []byte("[anon: Go: heap]")) {
			fmt.Print("process heap:  ")
			fields := bytes.Split(line, []byte{' '})
			os.Stdout.Write(fields[0])
			os.Stdout.Write([]byte{'\n'})
			continue
		}
		if bytes.Contains(line, []byte("[stack]")) {
			fmt.Print("process stack: ")
			fields := bytes.Split(line, []byte{' '})
			os.Stdout.Write(fields[0])
			os.Stdout.Write([]byte{'\n'})
			break
		}
	}
}

type stack struct {
	lo uintptr
	hi uintptr
}

type m struct {
	g0         *g
	morebuf    gobuf
	divmod     uint32
	procid     uint64
	gsignal    *g
	goSigStack gsignalStack
	sigmask    [2]uint32
	tls        [6]uintptr
	mstartfn   func()
	curg       *g
	xtra       [20]uintptr
}

type g struct {
	stack       stack
	stackguard0 uintptr
	stackguard1 uintptr
	pAnIc       *int
	dEfeR       *int
	m           *m
}

type gsignalStack struct {
	stack       stack
	stackguard0 uintptr
	stackguard1 uintptr
	stktopsp    uintptr
}

type gobuf struct {
	sp   uintptr
	pc   uintptr
	g    uintptr
	ctxt unsafe.Pointer
	lr   uintptr
	bp   uintptr
}

func printG() {
	uptr := routine.Getg()
	someg := (*g)(uptr)
	m := someg.m
	fmt.Printf("g        %p\n", someg)
	if someg != nil {
		fmt.Printf("\tstack hi 0x%x\n", someg.stack.hi)
		fmt.Printf("\tstack lo 0x%x\n", someg.stack.lo)
	}
	fmt.Printf("m.g0     %p\n", m.g0)
	if m.g0 != nil {
		fmt.Printf("\tstack hi 0x%x\n", m.g0.stack.hi)
		fmt.Printf("\tstack lo 0x%x\n", m.g0.stack.lo)
	}
	fmt.Printf("PID %d (0x%x)\n", m.procid, m.procid)
	fmt.Printf("gsignal  %p\n", m.gsignal)
	if m.gsignal != nil {
		fmt.Printf("\tstack hi 0x%x\n", m.gsignal.stack.hi)
		fmt.Printf("\tstack lo 0x%x\n", m.gsignal.stack.lo)
	}
	fmt.Printf("m.curg   %p\n", m.curg)
	if m.curg != nil {
		fmt.Printf("\tstack hi 0x%x\n", m.curg.stack.hi)
		fmt.Printf("\tstack lo 0x%x\n", m.curg.stack.lo)
	}
}

func doSlice2(z int) unsafe.Pointer {
	fmt.Printf("formal argument at %p\n", &z)
	x := make([]byte, 2)
	x[0] = 'a'
	x[1] = 'b'
	u := unsafe.Pointer(&x)
	sh := (*sliceHeader)(u)
	return sh.Data
}
