package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Printf("ES_URL=... %s: ds-type from-index to-index\n", os.Args[0])
		os.Exit(1)
		return
	}
}
