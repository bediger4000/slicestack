package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"unsafe"

	"github.com/timandy/routine"
)

/*
 * Investigate Go runtime's type g struct by retrieving
 * one and printing out a hex dump of it.
 */

func main() {
	var st stack
	var gst gsignalStack
	var gb gobuf
	var ui uintptr

	printHeapAddress()
	/*
		bsAt := doSlice(42)
		fmt.Printf("slice allocation at %p\n", bsAt)
	*/
	bsAt2 := doSlice2(42)
	fmt.Printf("slice allocation at %p\n", bsAt2)

	fmt.Printf("PID %d\t0x%x\n", os.Getpid(), os.Getpid())
	fmt.Printf("sizeof type stack %d\n", unsafe.Sizeof(st))
	fmt.Printf("sizeof type gsignalStack %d\n", unsafe.Sizeof(gst))
	fmt.Printf("sizeof type gobuf %d\n", unsafe.Sizeof(gb))
	fmt.Printf("sizeof type uintptr %d\n", unsafe.Sizeof(ui))

	printG()
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
	var structM m
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

	ln := int(unsafe.Sizeof(structM))

	str := unsafe.String((*byte)(unsafe.Pointer(m)), ln)

	for i := 0; i < ln; i++ {
		if (i > 0) && ((i % 8) == 0) {
			fmt.Println()
		}
		fmt.Printf("%02x ", str[i])
	}
	fmt.Println()
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

func doSlice(z int) *byte {
	var x []byte
	x = append(x, 'a')
	x = append(x, 'b')

	sd := unsafe.SliceData(x)

	return sd
}
