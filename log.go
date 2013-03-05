package gor

// 简化的Log调用

import (
	"flag"
	"fmt"
	"path/filepath"
	"runtime"
	"time"
)

var (
	debug = flag.Bool("debug", true, "Enable Debug, verbose output")
)

func D(v ...interface{}) {
	if !*debug {
		return
	}
	_, file, line, ok := runtime.Caller(1)

	if ok {
		if file != "" {
			file = filepath.Base(file)
		}
		fmt.Printf("%s %s:%d: %s", time.Now().Format("15:04:05"), file, line, fmt.Sprintln(v...))
	} else {
		fmt.Printf("%s >> %s", time.Now().String(), fmt.Sprint(v...))
	}
}
