package main

import (
	"flag"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/config"
)

var flagRunAddr string

func parseFlags(conf *config.Config) {

    flag.StringVar(&flagRunAddr, "a", "localhost:8080", "server address")

    flag.Parse()

    if conf.Server.Address != "" {
        flagRunAddr = conf.Server.Address
    }
}