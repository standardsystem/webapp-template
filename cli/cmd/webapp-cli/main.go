package main

import (
	"fmt"
	"io"
	"os"
)

// exitFn is swapped in tests.
var exitFn = os.Exit

func main() {
	code := run(os.Stdout, os.Stderr, os.Args[1:])
	exitFn(code)
}

func run(stdout, stderr io.Writer, args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage: webapp-cli <version|health>")
		return 2
	}
	switch args[0] {
	case "version":
		fmt.Fprintln(stdout, Version)
		return 0
	case "health":
		url := "http://localhost:8080/health"
		if len(args) > 1 {
			url = args[1]
		}
		if err := checkHealth(url); err != nil {
			fmt.Fprintf(stderr, "health check failed: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, "ok")
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command %q\n", args[0])
		fmt.Fprintln(stderr, "usage: webapp-cli <version|health [url]>")
		return 2
	}
}
