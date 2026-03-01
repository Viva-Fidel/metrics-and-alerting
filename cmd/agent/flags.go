package main

import (
    "flag"
)

var (
	flagRunAddr   string
	reportInterval int64
	pollInterval   int64
)

func parseFlags() {

    flag.StringVar(&flagRunAddr, "a", "localhost:8080", "server address")
	flag.Int64Var(&reportInterval, "r", 10, "report sending interval")
	flag.Int64Var(&pollInterval, "p", 2, "report collecting interval")

    flag.Parse()
}