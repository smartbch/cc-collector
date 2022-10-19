package main

import (
	"flag"

	"github.com/smartbch/cccollector/collector"
)

var sbchRpcUrl = "localhost:8545"

func main() {
	flag.StringVar(&sbchRpcUrl, "sbchRpcUrl", "localhost:8545", "smartBCH RPC URL")
	flag.Parse()

	collector.Run(sbchRpcUrl)
}
