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
	printMyStackAddr()

	doStuff()
	doStuff2()

}

func doStuff() {
	fmt.Printf("enter function\n")

	var x []byte

	x = append(x, 'a')
	x = append(x, 'b')

	fmt.Printf("local slice %p, variable at %p\n", x, &x)
	fmt.Printf("backing store at %p\n", backingStore(x))
	fmt.Printf("first element '%c'\n", x[0])
}

func doStuff2() {
	fmt.Printf("enter function 2\n")

	var a [2]byte

	x := a[:]

	x[0] = 'a'
	x[1] = 'b'

	fmt.Printf("local array first element at %p, variable at %p\n", &(a[0]), &a)
	fmt.Printf("local slice      %p, variable at %p\n", x, &x)
	fmt.Printf("backing store at %p\n", backingStore(x))
	fmt.Printf("first element '%c'\n", x[0])
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
			fmt.Print("heap:  ")
			fields := bytes.Split(line, []byte{' '})
			os.Stdout.Write(fields[0])
			os.Stdout.Write([]byte{'\n'})
			continue
		}
		if bytes.Contains(line, []byte("[stack]")) {
			fmt.Print("stack: ")
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

type g struct {
	stack stack
}

func printMyStackAddr() {
	uptr := routine.Getg()
	fmt.Printf("%p\n", uptr)

	myg := (*g)(uptr)
	fmt.Printf("my stack:\nhi 0x%x\nlo 0x%x\n", myg.stack.hi, myg.stack.lo)
}
