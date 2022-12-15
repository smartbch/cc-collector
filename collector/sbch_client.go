package collector

import (
	"context"
	"time"

	sbchclient "github.com/smartbch/smartbch/rpc/client"
	sbchrpc "github.com/smartbch/smartbch/rpc/types"
)

const (
	getTimeout = time.Second * 15
)

func getCcInfo(client *sbchclient.Client) (*sbchrpc.CcInfo, error) {
	ctx, cancelFn := context.WithTimeout(context.Background(), getTimeout)
	ccInfo, err := client.CcInfo(ctx)
	defer cancelFn()

	return ccInfo, err
}

func getRedeemingUtxosForOperators(client *sbchclient.Client) ([]*sbchrpc.UtxoInfo, error) {
	ctx, cancelFn := context.WithTimeout(context.Background(), getTimeout)
	utxos, err := client.RedeemingUtxosForOperators(ctx)
	defer cancelFn()

	return utxos.Infos, err
}

func getToBeConvertedUtxosForOperators(client *sbchclient.Client) ([]*sbchrpc.UtxoInfo, error) {
	ctx, cancelFn := context.WithTimeout(context.Background(), getTimeout)
	utxos, err := client.ToBeConvertedUtxosForOperators(ctx)
	defer cancelFn()

	return utxos.Infos, err
}

func getOperatorPubkeys(operators []*sbchrpc.OperatorInfo) [][]byte {
	pubkeys := make([][]byte, len(operators))
	for i, operator := range operators {
		pubkeys[i] = operator.Pubkey
	}
	return pubkeys
}
func getMonitorPubkeys(monitors []*sbchrpc.MonitorInfo) [][]byte {
	pubkeys := make([][]byte, len(monitors))
	for i, monitor := range monitors {
		pubkeys[i] = monitor.Pubkey
	}
	return pubkeys
}
