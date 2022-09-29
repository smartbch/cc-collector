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
		handleAllPendingUTXOs(sbchClient)
		time.Sleep(1 * time.Minute)
	}
}

func handleAllPendingUTXOs(sbchClient *SbchClient) {
	cccInfo, err := sbchClient.getCcCovenantInfo()
	if err != nil {
		fmt.Println("failed to get CcCovenantInfo:", err.Error())
		return
	}

	redeemingUtxos, err := sbchClient.getRedeemingUtxosForOperators()
	if err != nil {
		fmt.Println("failed to get redeeming UTXOs:", err.Error())
		return
	}
	if len(redeemingUtxos) > 0 {
		operatorPubkeys := getOperatorPubkeys(cccInfo.Operators)
		monitorPubkeys := getMonitorPubkeys(cccInfo.Monitors)
		ccCovenant, err := ccc.NewDefaultCcCovenant(operatorPubkeys, monitorPubkeys)
		if err != nil {
			fmt.Println("failed to create CcCovenant instance:", err.Error())
			return
		}

		for _, utxo := range redeemingUtxos {
			handleRedeemingUTXO(ccCovenant, cccInfo.Operators, utxo)
		}
	}

	toBeConvertedUtxos, err := sbchClient.getToBeConvertedUtxosForOperators()
	if err != nil {
		fmt.Println("failed to get redeeming UTXOs:", err.Error())
		return
	}
	if len(toBeConvertedUtxos) > 0 {
		oldOperatorPubkeys := getOperatorPubkeys(cccInfo.OldOperators)
		oldMonitorPubkeys := getMonitorPubkeys(cccInfo.OldMonitors)
		newOperatorPubkeys := getOperatorPubkeys(cccInfo.Operators)
		newMonitorPubkeys := getMonitorPubkeys(cccInfo.Monitors)
		ccCovenant, err := ccc.NewDefaultCcCovenant(oldOperatorPubkeys, oldMonitorPubkeys)
		if err != nil {
			fmt.Println("failed to create CcCovenant instance:", err.Error())
			return
		}

		for _, utxo := range toBeConvertedUtxos {
			handleToBeConvertedUTXO(ccCovenant, cccInfo.OldOperators,
				newOperatorPubkeys, newMonitorPubkeys, utxo)
		}
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

	_, rawTx, err := ccCovenant.FinishRedeemByUserTx(tx, sigs)
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

func handleToBeConvertedUTXO(
	oldCcCovenant *ccc.CcCovenant,
	oldOperators []sbchrpc.OperatorInfo,
	newOperatorPubkeys [][]byte,
	newMonitorPubkeys [][]byte,
	utxo *sbchrpc.UtxoInfo,
) {
	txid := utxo.Txid[:]
	vout := utxo.Index
	amt := int64(utxo.Amount)
	tx, sigHash, err := oldCcCovenant.GetConvertByOperatorsTxSigHash(txid, vout, amt,
		newOperatorPubkeys, newMonitorPubkeys)
	if err != nil {
		fmt.Println("failed to call GetConvertByOperatorsTxSigHash:", err.Error())
		return
	}

	var sigs [][]byte
	for _, operator := range oldOperators {
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

	_, rawTx, err := oldCcCovenant.FinishConvertByOperatorsTx(tx,
		newOperatorPubkeys, newMonitorPubkeys, sigs)
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

func broadcastBchTx(rawTx []byte) error {
	// TODO
	fmt.Println(string(rawTx))
	return nil
}