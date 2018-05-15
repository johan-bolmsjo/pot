package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/johan-bolmsjo/pot"
)

func main() {
	buf, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fatalf("Faild to read from stdin, %s\n")
	}

	if buf, err = pot.PrettyPrint(buf); err != nil {
		fatalf("Failed to pretty print POT, %s\n", err)
	}

	fmt.Println(string(buf))
}

func fatalf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(1)
}
