package main

import (
	"flag"
	"fmt"

	"github.com/foomo/nodeflock"
)

func main() {
	flagAddr := flag.String("addr", ":8888", "address to bind to")
	flagSize := flag.Int("flock-size", 10, "number of node processes to run")
	flag.Parse()
	fmt.Println("starting socket server on", *flagAddr, "with a flock size of", *flagSize)
	fmt.Println("exiting with error", nodeflock.RunServer(*flagAddr, *flagSize, flag.Arg(0)))
}
