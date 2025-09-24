package main

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"
)

func stdinHandler() error {
	// read from stdin until EOF
	scanner := bufio.NewScanner(os.Stdin)
	var b strings.Builder
	var t0 time.Time

	for scanner.Scan() {
		t0 = time.Now()
		b.WriteString(scanner.Text() + "\n")
		b.WriteString(Footer(t0))

		_, err := os.Stdout.WriteString(b.String())
		if err != nil {
			panic(err)
		}
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return nil
}

func cliHandler() string {
	t0 := time.Now()
	var b strings.Builder
	b.WriteString("hello!\n")
	for i, arg := range os.Args {
		b.WriteString("    $" + strconv.Itoa(i) + ": " + arg + "\n")
	}
	b.WriteString(Footer(t0))
	return b.String()
}
