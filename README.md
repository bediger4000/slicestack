# Demonstrate that slice backing store can be stack allocated

[The Go Blog](https://go.dev/blog/)
has a post titled
[Allocating on the Stack](https://go.dev/blog/allocation-optimizations).

This post makes the claim that Go v1.26, some combination of compiler and runtime,
allocates backing store of slices on the stack.
Performance gains ensue,
presumably because CPU hardware has instructions to push and pop stack frames
for function calls,
treating the stack as an informal
[arena allocator](https://en.wikipedia.org/wiki/Region-based_memory_management).

### Obtain code for this repository

```
$ cd $GOPATH/src
$ git clone $ cd $GOPATH/src
$ git clone https://github.com/timandy/routine.git
```

### Obtain and fix up `package routine`
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

### Set `go.mod` to use modified `package routine`

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

This piece of prestidigitation lets you specify that the `func Getg()`
my code invokes will actually be compiled from the *modified* `package routine`,
not some plain vanilla, rule bound `package routine` downloaded from the internet.
