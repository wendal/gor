// Gor - Fastest Static Blog Engine
package main

import (
	"fmt"
)

const (
	HELP = `
Run specified gor tool

Usage:

	gor command [args...]

Init Blog layout

    gor init <dir>

Compile

	gor compile

Preview Compiled Website

	gor http

Print Configure

	gor config

Print Payload

	gor payload

run pprof (for dev)

	gor pprof

	`
)

func PrintUsage() {
	fmt.Println(HELP)
}
