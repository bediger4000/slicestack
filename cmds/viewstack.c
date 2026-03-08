#include <stdio.h>
#include <stdlib.h>

/*
 * Show that a C process has formal function
 * arguments and local variables on the system
 * stack.
 */

void delineateStack(int n);

int main(int argc, char **argv) {
	int n = atoi(argv[1]);
	delineateStack(n);
}

void delineateStack(int n) {
	int l = n;
	printf("formal argument at %p\n", &n);
	printf("local  variable at %p\n", &l);

	if (n == 0)
		return;

	delineateStack(n - 1);
}
