package main

import (
	"flag"

	"github.com/smartbch/cc-collector/collector"
)

var (
	sbchRpcUrl = "localhost:8545"
	bchRpcUrl  = "localhost:8332"
	bchRpcUser = "user"
	bchRpcPass = "pass"
)

func main() {
	flag.StringVar(&sbchRpcUrl, "sbch-rpc-url", sbchRpcUrl, "smartBCH RPC URL")
	flag.StringVar(&bchRpcUrl, "bch-rpc-url", bchRpcUrl, "BitcoinCash RPC URL")
	flag.StringVar(&bchRpcUser, "bch-rpc-user", bchRpcUser, "BitcoinCash RPC username")
	flag.StringVar(&bchRpcPass, "bch-rpc-pass", bchRpcPass, "BitcoinCash RPC password")
	flag.Parse()

	collector.Run(sbchRpcUrl, bchRpcUrl, bchRpcUser, bchRpcPass)
}
