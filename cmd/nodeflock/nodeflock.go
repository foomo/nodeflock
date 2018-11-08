package main

import (
	"flag"
	"fmt"
	"github.com/foomo/nodeflock"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"time"
)

var (
	version string
)

const (
	defaultNodeflockAddress          = ":8888"
	defaultNodeflockSize             = 10
	defaultNodeflockMaxExecutionTime = 10

	defaultPrometheusAddress = ":80"
)

func main() {
	flagAddr := flag.String("addr", defaultNodeflockAddress, "address to bind to")
	flagSize := flag.Int("flock-size", defaultNodeflockSize, "number of node processes to run")
	flagMaxExecutionTime := flag.Int("max-execution-time", defaultNodeflockMaxExecutionTime, "max execution time in seconds")
	flag.Parse()

	fmt.Println("Running version", version)

	go startMetricsListener()
	go startNodeflockListener(*flagAddr, *flagSize, *flagMaxExecutionTime)

	select {}
}

func startMetricsListener() {
	prometheus := http.NewServeMux()
	prometheus.HandleFunc("/metrics", promhttp.Handler().ServeHTTP)
	fmt.Println("Starting metrics listener on ", defaultPrometheusAddress)
	log.Print(http.ListenAndServe(defaultPrometheusAddress, prometheus))
}

func startNodeflockListener(address string, size int, maxExecutionTime int) {
	fmt.Println("Starting socket server on", address, "with a flock size of", size, "and a max execution time of", maxExecutionTime, "s")
	log.Fatal(nodeflock.RunServer(address, size, flag.Arg(0), time.Duration(maxExecutionTime)*time.Second))
}
