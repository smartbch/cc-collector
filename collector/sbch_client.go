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
	defer cancelFn()

	return client.CcInfo(ctx)
}

func getRedeemingUtxosForOperators(client *sbchclient.Client) ([]*sbchrpc.UtxoInfo, error) {
	ctx, cancelFn := context.WithTimeout(context.Background(), getTimeout)
	defer cancelFn()

	utxos, err := client.RedeemingUtxosForOperators(ctx)
	if err != nil {
		return nil, err
	}
	return utxos.Infos, nil
}

func getToBeConvertedUtxosForOperators(client *sbchclient.Client) ([]*sbchrpc.UtxoInfo, error) {
	ctx, cancelFn := context.WithTimeout(context.Background(), getTimeout)
	defer cancelFn()

	utxos, err := client.ToBeConvertedUtxosForOperators(ctx)
	if err != nil {
		return nil, err
	}
	return utxos.Infos, nil
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
