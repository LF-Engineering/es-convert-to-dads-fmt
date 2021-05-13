package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"time"
)

var (
	gESURL   string
	gDSTypes = map[string]struct{}{
		"git":    {},
		"github": {},
	}
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

func convertGitHub(idxFrom, idxTo string) (err error) {
	return nil
}

func convertGit(idxFrom, idxTo string) (err error) {
	return fmt.Errorf("git is not implemented yet")
}

func convert(dsType, idxFrom, idxTo string) (err error) {
	switch dsType {
	case "git":
		err = convertGit(idxFrom, idxTo)
	case "github":
		err = convertGitHub(idxFrom, idxTo)
	default:
		err = fmt.Errorf("%s support not implemented", dsType)
	}
	return
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
	dsType := os.Args[1]
	_, ok := gDSTypes[dsType]
	if !ok {
		fatalf("%s: %s is not a know ds-type, allowed are: %+v\n", os.Args[0], dsType, gDSTypes)
		return
	}
	fatalOnError(convert(dsType, os.Args[2], os.Args[3]))
}
