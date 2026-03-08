package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

/*
 * Show that a go process has formal function
 * arguments and local variables on the heap
 * at least if you print out addresses, which
 * causes escapes to heap.
 */

func main() {
	if n, err := strconv.Atoi(os.Args[1]); err != nil {
		log.Fatal(err)
	} else {
		fmt.Printf("main   variable at %p\n", &n)
		delineateStack(n)
	}
}

func delineateStack(n int) {
	fmt.Printf("formal argument at %p\n", &n)
	l := n
	fmt.Printf("local  variable at %p\n", &l)
	if n > 0 {
		delineateStack(n - 1)
	}
}
