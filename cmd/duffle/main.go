package main

import (
	"fmt"
	"os"
)

func unimplemented(msg string) {
	panic(fmt.Errorf("unimplemented: %s", msg))
}

func must(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "duffle: fatal: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				must(err)
			}
		}
	}()
	must(newRootCmd(os.Stdout).Execute())
}
