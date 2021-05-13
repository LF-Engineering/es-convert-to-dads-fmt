package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"time"
)

var (
	gESURL string
)

func fatalOnError(err error) {
	if err != nil {
		tm := time.Now()
		msg := fmt.Sprintf("Error(time=%+v):\nError: '%s'\nStacktrace:\n%s\n", tm, err.Error(), string(debug.Stack()))
		fmt.Printf("%s", msg)
		fmt.Fprintf(os.Stderr, "%s", msg)
		panic("stacktrace")
	}
}

func fatalf(f string, a ...interface{}) {
	fatalOnError(fmt.Errorf(f, a...))
}

func main() {
	if len(os.Args) < 3 {
		fatalf("ES_URL=... %s: ds-type from-index to-index\n", os.Args[0])
		return
	}
	gESURL = os.Getenv("ES_URL")
	if gESURL == "" {
		fatalf("%s: you need to set ES_URL environment variable\n", os.Args[0])
		return
	}
}
