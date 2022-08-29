package main

import (
	"fmt"
	"time"

	gethcmn "github.com/ethereum/go-ethereum/common"

	ccc "github.com/smartbch/smartbch/crosschain/covenant"
	sbchrpc "github.com/smartbch/smartbch/rpc/api"
)

const sbchRpcUrl = "localhost:8545" // TODO

const (
	minOperatorSigCount = 7
)

func main() {
	sbchClient, err := newSbchClient(sbchRpcUrl)
	if err != nil {
		fmt.Println("failed to create smartBCH RPC client:", err.Error())
		return
	}

	for {
		handleAllRedeemingUTXOs(sbchClient)
		time.Sleep(1 * time.Minute)
	}
}

func handleAllRedeemingUTXOs(sbchClient *SbchClient) {
	operators, err := sbchClient.getOperators()
	if err != nil {
		fmt.Println("failed to get operators:", err.Error())
		return
	}

	monitors, err := sbchClient.getMonitors()
	if err != nil {
		fmt.Println("failed to get monitors:", err.Error())
		return
	}

	// TODO: check count of operators and monitors
	operatorPubkeys := getOperatorPubkeys(operators)
	monitorPubkeys := getMonitorPubkeys(monitors)
	ccCovenant, err := ccc.NewCcCovenantMainnet(operatorPubkeys, monitorPubkeys)
	if err != nil {
		fmt.Println("failed to create CcCovenant instance:", err.Error())
		return
	}

	utxos, err := sbchClient.getRedeemingUtxosForOperators()
	if err != nil {
		fmt.Println("failed to get redeeming UTXOs:", err.Error())
		return
	}

	for _, utxo := range utxos {
		handleRedeemingUTXO(ccCovenant, operators, utxo)
	}
}

func handleRedeemingUTXO(
	ccCovenant *ccc.CcCovenant,
	operators []sbchrpc.OperatorInfo,
	utxo *sbchrpc.UtxoInfo,
) {
	txid := utxo.Txid[:]
	vout := utxo.Index
	amt := int64(utxo.Amount)
	toAddr := sbchAddrToBchAddr(utxo.RedeemTarget)
	tx, sigHash, err := ccCovenant.GetRedeemByUserTxSigHash(txid, vout, amt, toAddr)
	if err != nil {
		fmt.Println("failed to call GetRedeemByUserTxSigHash:", err.Error())
		return
	}

	var sigs [][]byte
	for _, operator := range operators {
		sig, err := getSigByHash(operator.RpcUrl, sigHash)
		if err != nil {
			fmt.Println("failed to query sig by hash:", err.Error())
			continue
		}

		sigs = append(sigs, sig)
	}

	if len(sigs) < minOperatorSigCount {
		fmt.Println("not enough operator sigs")
		return
	}

	rawTx, err := ccCovenant.FinishRedeemByUserTx(tx, sigs)
	if err != nil {
		fmt.Println("failed to sign tx:", err.Error())
		return
	}
	fmt.Println("rawTx:", rawTx)

	err = broadcastBchTx(rawTx)
	if err != nil {
		fmt.Println("failed to broadcast BCH tx:", err.Error())
	}

	// TODO
}

func sbchAddrToBchAddr(addr gethcmn.Address) string {
	// TODO
	return addr.String()
}

func broadcastBchTx(rawTx string) error {
	// TODO
	fmt.Println(rawTx)
	return nil
}
