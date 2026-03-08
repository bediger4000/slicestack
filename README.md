# Demonstrate that slice backing store can be stack allocated

[The Go Blog](https://go.dev/blog/)
has a post titled
[Allocating on the Stack](https://go.dev/blog/allocation-optimizations).
The post makes the claim that 
programs compiled with Go v1.26 compiler,
and using the v1.26 runtime
can allocate backing store of slices on the stack
under some circumstances.
Performance gains ensue,
presumably because CPU hardware has instructions to push and pop stack frames
for function calls,
treating the stack as an informal
[arena allocator](https://en.wikipedia.org/wiki/Region-based_memory_management).

I figured out an admirable hack to demonstrate this to myself.
I did this using Linux and Go v1.26.
You might be able to do the same, or similar, stunt under a lesser operating system.

The trick has two parts.

1. Get the high and low addresses of a goroutine's stack
from the Go runtime.
2. Find the address of a slice's backing store.

Once these 3 pieces of data are available,
compare address of backing store to the stack addresses.

Here's how to set up to get the current goroutine's stack addresses.

### 1. Obtain code for this repository

```
$ cd $GOPATH/src
$ git clone https://github.com/bediger4000/slicestack.git
```

### 2. Obtain and fix up `package routine`

```
$ cd $GOPATH/src
$ git clone https://github.com/timandy/routine.git
```

Add the following function to file `routine/g.go`:
```
func Getg() unsafe.Pointer {
    return unsafe.Pointer(getg())
}
```

The author(s) of `package routine` did not export a way
to retrieve the current goroutine's `struct g` from the
Go runtime.
This extra function adds an exported version.

### 3. Set `go.mod` to use modified `package routine`

Edit this repo's `go.mod` file so its contents resemble this:

```
module slicestack

go 1.25.0

require github.com/timandy/routine v1.1.6

replace github.com/timandy/routine v1.1.6 => $GOPATH/src/routine
```

You must replace `$GOPATH` in the example `go.mod` contents above
with your actual `$GOPATH` fully qualified directory.
I don't believe `go build` does shell expansion on `go.mod` contents.

This piece of prestidigitation lets you specify that the newly added `func Getg()`
will actually be compiled from the *modified* `package routine`,
not some plain vanilla, rule bound `package routine` downloaded from the internet.

## Compile and run program

```
$ cd $GOPATH/src/slicestack
$ go build -gcflags="-m"  $PWD 2> escape.analysis
```

That should compile without complaint.
Running it should look something like this:

```
$ ./slicestack
system heap:  ec882000000-ec882800000
system stack: 7ffc2822e000-7ffc2824f000

slice backing store addressing check
backing store at 0xec8823560c0
first element at 0xec8823560c0

stack top     00000ec8823af000
backing store 00000ec8823aee60
stack bottom  00000ec8823ae000
Slice backing store on stack: true
```

### What does this code do?

1. Prints the traditional Unix heap and stack memory allocations,
via reading `/proc/$PID/maps` pseudo-file.
2. Demonstrates that the address of slice's backing store
can be obtained with `&(slice[0])`.
3. Checks that a slice's backing store gets allocated
on the goroutine's stack.

This is the function that allocates a slice, finds stack addresses,
and checks that the address of the backing store is between
the stack's high and low addresses.

```
 1	func checkStackAddr() bool {
 2	    myg := (*g)(routine.Getg())
 3	    
 4	    x := make([]byte, 64)
 5	    backingStore := uintptr(unsafe.Pointer(&(x[0])))
 6	
 7	    fmt.Printf("stack top     %016x\n", myg.stack.hi)
 8	    fmt.Printf("backing store %016x\n", backingStore)
 9	    fmt.Printf("stack bottom  %016x\n", myg.stack.lo)
10	    
11	    if backingStore > myg.stack.hi || backingStore < myg.stack.lo {
12	        return false
13	    }
14	
15	    return true
16	}
```

- Line 2 calls the exported function added to `package routine`.
- Line 4 allocates a slice with 64 bytes of backing storage.
- Line 5 obtains the address of the backing store,
*without triggering the slice escaping to the heap*.
- Line 11 checks that the address of the backing store
lies between the goroutine stack's high and low addresses.

## How this code does it

### Compiler shenanigans

The procedure above plays some compiler tricks
to get access to a struct that
has the current goroutine's stack addresses.

Code in `package routine`
gains access to a Go runtime function `func getgc() *g`
via linker sleight of hand.
This is real weird stuff, and very exacting work,
to be honest.

Unfortunately, `package routine` doesn't export the function
resulting from the linker sleight of hand.
My compilation procedure adds an exported function, `func routine.Getg() unsafe.Pointer`,
to the local clone of the source code.
After compiling with the local source code,
my code can receive the return value of Go runtime `func getg()`,
and from that, get addresses of the top and bottom of the current
goroutine's call stack.

### Type system head fake 1.

My code defines two structures:

```
type stack struct {
    lo uintptr
    hi uintptr
}

type g struct {
    stack       stack
    stackguard0 uintptr
    stackguard1 uintptr
}   
```
I lifted `type stack` verbatim from `/usr/lib/go/src/runtime/runtime2.go`.
My `type g struct` is the first 3 elements of a Go runtime unexported `type g`.
Instances of `type g` inside the Go runtime represent all necessary info about a goroutine,
including its call stack.
I'm interested in only the `stack` element.

Official [Go documentation](https://github.com/golang/go/blob/master/src/runtime/HACKING.md)
talks about `type g` and other relevant runtime things.

Since `routine.Getg()` returns a value of type `unsafe.Pointer`,
doing a type conversion to my `type g struct` of that `unsafe.Pointer` value
gets my code access to high and low addresses of the goroutine's call stack.

### Type system head fake 2.

You can find the address of a slice's backing store from `func unsafe.SliceData()`,
now part of the standard Go library.
Calling that function causes the compiler to allocate the slice
and its backing store from the heap.

You can also define a new struct, and do some unsafe type conversions:

```
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
```

Calling `func backingStore` also causes the compiler to allocate
slice and backing store on the heap,
so you'd have to include this code in-line.

It's simpler to take advantage of the organization of slice structs
in the Go runtime.
It turns out that doing something like the following actually gives you the address
of a slice's backing store:

```
sl := make([]int, 24)
backingStoreAddress := unsafe.Pointer(&(sl[0]))
```

I don't like this trick,
although the same "take address of first element of an array"
works in C programs as well.

### Avoid escape analysis causing heap allocation

[func checkStackAddr](#what-does-this-code-do)
returns a boolean value partially to avoid
having the Go compiler's [escape analysis]()
decide to do heap allocation.

I avoided doing a couple of inobvious things in `func checkStackAddr`
to avoid escapes to heap in `func checkStackAddr`.

- Don't call `fmt.Printf()` to show the slice or backing store address
with the wrong type argument.
- Don't return backing store address or slice itself for output.

I did discover that type converting slice backing store addresses
from, say `*byte` to `uintptr` via `unsafe.Pointer` seems to fool
escape analyis.

The compiler must be aware of something special
about function `unsafe.SliceData`.
Calling any other function with a slice argument seems to cause
the compiler to set heap allocation for that slice.
Note that the allocation of a slice takes place before the function call or return
statement that causes an escape to heap.

My [example compilation](#4-compile-and-run-program) command line
saves escape analysis output to a file to demonstrate.
Add a `fmt.Printf("%p\n", backingStore)` after line 5
of [func checkStackAddr](#what-does-this-code-do).
Re-compile as shown. 
A look in resulting file `escape.analysis` will show that
slice `x` escapes to the heap.
Running the program will cause `func checkStackAddr` to return false as well.
