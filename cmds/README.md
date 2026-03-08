# Programs used to investigate

Some code I used to figure out how the Go runtime `type g`,
runtime stacks and other things work.

- `dumpg.go` Hex dump of Go runtime's `type g`
- `probestack.go` Use a recursive function to see about stack variables
- `viewstack.go`
- `viewstack.c` Show that C programs put function arguments and local variables on the stack
