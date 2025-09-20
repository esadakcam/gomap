package main

import (
	"fmt"
	"os"

	"github.com/esadakcam/gomap/internal/gomap"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: gomap <ip-range>")
		os.Exit(1)
	}

	ipRange := os.Args[1]
	app, err := gomap.Scan(ipRange)
	if err != nil {
		panic(err)
	}
	fmt.Println(app)
}
