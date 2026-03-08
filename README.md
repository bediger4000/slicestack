# Demonstrate that slice backing store can be stack allocated

[The Go Blog](https://go.dev/blog/)
has a post titled
[Allocating on the Stack](https://go.dev/blog/allocation-optimizations).

The Go Blog post makes the claim that 
programs compiled with Go v1.26 compiler,
and using the v1.26 runtime
allocate backing store of slices on the stack.
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

The Go runtime carefully hides stack addresses from a running goroutine.
The newish standard library function `unsafe.SliceData` can give you
a slice's backing store address, but for reasons is not useful for
finding out if stack allocation of slice backing store occurs.


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

### 4. Compile and run program

```
$ cd $GOPATH/src/slicestack
$ go build -gcflags="-m"  $PWD 2> escape.analysis
```

That should compile without complaint.
Running it should look something like this:

```
$ ./slicestack
system heap:  186c7e000000-186c7e400000
system stack: 7ffc35139000-7ffc3515a000

slice backing store addressing check
backing store at 0x186c7e1980c0
first element at 0x186c7e1980c0

Slice backing store on stack: true
```

## What does this code do?

1. It prints the traditional Unix heap and stack memory allocations,
via reading `/proc/$PID/maps` pseudo-file.
2. It proves that the address of slice's backing store
can be obtained with `&(slice[0])`.
3. It checks that a slice's backing store gets allocated
on the goroutine's stack.

```
 1	func checkStackAddr() bool {
 2	    myg := (*g)(routine.Getg())
 3	
 4	    x := make([]byte, 64)
 5	    backingStore := uintptr(unsafe.Pointer(&(x[0])))
 6	
 7	    if myg.stack.hi > backingStore && backingStore > myg.stack.lo {
 8	        return true
 9	    }
10	
11	    return false
12	}
```

- Line 2 calls the exported function added to `package routine`.
- Line 4 allocates a slice with 64 bytes of backing storage.
- Line 5 obtains the address of the backing store,
*without triggering the slice escaping to the heap*.
- Line 7 checks that the address of the backing store
lies between the goroutine stack's high and low addresses.

## How this code does it

### Compiler shenanigans

The procedure above does some compiler shenanigans
to get around Go's strict type system, and to get access to the Go runtime
at run time.

My program uses code in `package routine`
that "exports" a Go runtime function `func getgc() *g`
by doing some linker sleight of hand.

Unfortunately, `package routine` doesn't export its own function.
My compilation procedure adds an exported function, `routine.Getg() unsafe.Pointer`,
to the local clone of the source code.
After compiling with the local source code,
my code can receive the return value of Go runtime `func getg()`.

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
Since `routine.Getg()` returns a value of type `unsafe.Pointer`,
doing a type conversion to my `type g struct` of that `unsafe.Pointer` value
gets my code access to high and low addresses of the goroutine's call stack.

### Type system head fake 2.

You can find the address of a slice's backing store from `func unsafe.SliceData()`,
now part of the standard Go library,
or by defining a new struct, and some unsafe type conversions:

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

Or you can take advantage of the organization of slice structs
in the Go runtime.

It turns out that doing something like the following actually gives you the address
of a slice's backing store:

```
sl := make([]int, 24)
backingStoreAddress := unsafe.Pointer(&(sl[0]))
```

I don't like that this trick actually works,
although the same "take address of first element of an array"
works in C programs as well.

### Don't try to fool escape analysis

[func checkStackAddr](#what-does-this-code-do)
returns a boolean value partially to avoid
having the Go compiler's [escape analysis]()
decide to do heap allocation.

I wrote a number of inobvious things in `func checkStackAddr`
to avoid escapes to heap in `func checkStackAddr`.

- Allocate slice using `make()`
- Avoid calling `append()` to set a slice value.
- Get address of backing store via `uintptr(unsafe.Pointer(&(x[0])))`
instead of calling `unsafe.SliceData`.
- Doesn't calling `fmt.Printf()` to show the slice or backing store address.
- Don't return backing store address or slice itself for output.

Any of these things causes the Go compiler's escape analysis
to allocate slice and backing store from the heap.
My [example compilation](#4-compile-and-run-program)
saves escape analysis output to a file to demonstrate.
Add a `fmt.Printf("%p\n", backingStore)` after line 5
of [func checkStackAddr](#what-does-this-code-do).
Re-compile as shown. 
A look in resulting file `escape.analysis` will show that
slice `x` escapes to the heap.
Running the program will cause `func checkStackAddr` to return false as well.
