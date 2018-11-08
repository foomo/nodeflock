package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/foomo/nodeflock"
)

var Version string

func main() {
	flagAddr := flag.String("addr", ":8888", "address to bind to")
	flagSize := flag.Int("flock-size", 10, "number of node processes to run")
	flagMaxExecutionTime := flag.Int("max-execution-time", 10, "max execution time in seconds")
	flag.Parse()
	fmt.Println("Running version: " + Version)
	fmt.Println("Starting socket server on", *flagAddr, "with a flock size of", *flagSize, "and a max execution time of", *flagMaxExecutionTime, "s")
	fmt.Println("Exiting with error", nodeflock.RunServer(*flagAddr, *flagSize, flag.Arg(0), time.Duration(*flagMaxExecutionTime)*time.Second))
}
