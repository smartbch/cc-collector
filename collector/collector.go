package collector

import (
	"encoding/hex"
	"fmt"
	"time"

	gethcmn "github.com/ethereum/go-ethereum/common"
	"github.com/gcash/bchd/chaincfg"
	"github.com/gcash/bchutil"

	ccc "github.com/smartbch/smartbch/crosschain/covenant"
	sbchclient "github.com/smartbch/smartbch/rpc/client"
	sbchrpc "github.com/smartbch/smartbch/rpc/types"
)

const (
	minOperatorSigCount = 7
)

func Run(sbchRpcUrl, bchRpcUrl, bchRpcUsername, bchRpcPassword string) {
	bchClient := newBchClient(bchRpcUrl, bchRpcUsername, bchRpcPassword)
	sbchClient, err := sbchclient.Dial(sbchRpcUrl)
	if err != nil {
		fmt.Println("failed to create smartBCH RPC client:", err.Error())
		return
	}

	for {
		handleAllPendingUTXOs(sbchClient, bchClient)
		time.Sleep(1 * time.Minute)
	}
}

func handleAllPendingUTXOs(sbchClient *sbchclient.Client, bchClient *BchRpcClient) {
	ccInfo, err := getCcInfo(sbchClient)

	if err != nil {
		fmt.Println("failed to get CcCovenantInfo:", err.Error())
		return
	}

	redeemingUtxos, err := getRedeemingUtxosForOperators(sbchClient)
	if err != nil {
		fmt.Println("failed to get redeeming UTXOs:", err.Error())
		return
	}
	if len(redeemingUtxos) > 0 {
		operatorPubkeys := getOperatorPubkeys(ccInfo.Operators)
		monitorPubkeys := getMonitorPubkeys(ccInfo.Monitors)
		ccCovenant, err := ccc.NewDefaultCcCovenant(operatorPubkeys, monitorPubkeys)
		if err != nil {
			fmt.Println("failed to create CcCovenant instance:", err.Error())
			return
		}

		_ccAddr, _ := ccCovenant.GetP2SHAddress()
		fmt.Println("ccCovenantAddr:", _ccAddr)
		for _, utxo := range redeemingUtxos {
			handleRedeemingUTXO(bchClient, ccCovenant, ccInfo.Operators, utxo)
		}
	}

	toBeConvertedUtxos, err := getToBeConvertedUtxosForOperators(sbchClient)
	if err != nil {
		fmt.Println("failed to get toBeConverted UTXOs:", err.Error())
		return
	}
	if len(toBeConvertedUtxos) > 0 {
		oldOperatorPubkeys := getOperatorPubkeys(ccInfo.OldOperators)
		oldMonitorPubkeys := getMonitorPubkeys(ccInfo.OldMonitors)
		newOperatorPubkeys := getOperatorPubkeys(ccInfo.Operators)
		newMonitorPubkeys := getMonitorPubkeys(ccInfo.Monitors)

		if len(oldOperatorPubkeys) == 0 {
			oldOperatorPubkeys = newOperatorPubkeys
		}
		if len(oldMonitorPubkeys) == 0 {
			oldMonitorPubkeys = newMonitorPubkeys
		}

		ccCovenant, err := ccc.NewDefaultCcCovenant(oldOperatorPubkeys, oldMonitorPubkeys)
		if err != nil {
			fmt.Println("failed to create CcCovenant instance:", err.Error())
			return
		}

		for _, utxo := range toBeConvertedUtxos {
			handleToBeConvertedUTXO(bchClient, ccCovenant, ccInfo.OldOperators,
				newOperatorPubkeys, newMonitorPubkeys, utxo)
		}
	}
}

func handleRedeemingUTXO(
	bchClient *BchRpcClient,
	ccCovenant *ccc.CcCovenant,
	operators []*sbchrpc.OperatorInfo,
	utxo *sbchrpc.UtxoInfo,
) {
	fmt.Println("handleRedeemingUTXO ...")
	fmt.Println("covenant:", utxo.CovenantAddr.String())
	fmt.Println("txid:", hex.EncodeToString(utxo.Txid[:]))
	fmt.Println("index:", utxo.Index)
	fmt.Println("amount:", utxo.Amount)
	fmt.Println("target:", utxo.RedeemTarget.String())
	fmt.Println("txSigHash:", hex.EncodeToString(utxo.TxSigHash))

	txid := utxo.Txid[:]
	vout := utxo.Index
	amt := int64(utxo.Amount)
	toAddr, err := sbchAddrToBchAddr(utxo.RedeemTarget)
	if err != nil {
		fmt.Println("failed to convert smartBCH address to BCH address:", err.Error())
		return
	}

	fmt.Println("toAddr:", toAddr)
	tx, sigHash, err := ccCovenant.GetRedeemByUserTxSigHash(txid, vout, amt, toAddr)
	fmt.Println("sigHash:", hex.EncodeToString(sigHash))
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

		fmt.Println(operator.RpcUrl, "sig:", hex.EncodeToString(sig))
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
	fmt.Println("rawTx:", hex.EncodeToString(rawTx))

	err = bchClient.sendRawTx(rawTx)
	if err != nil {
		fmt.Println("failed to broadcast BCH tx:", err.Error())
	}

	// TODO
}

func handleToBeConvertedUTXO(
	bchClient *BchRpcClient,
	oldCcCovenant *ccc.CcCovenant,
	oldOperators []*sbchrpc.OperatorInfo,
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
	fmt.Println("rawTx:", hex.EncodeToString(rawTx))

	err = bchClient.sendRawTx(rawTx)
	if err != nil {
		fmt.Println("failed to broadcast BCH tx:", err.Error())
	}

	// TODO
}

func sbchAddrToBchAddr(sbchAddr gethcmn.Address) (string, error) {
	bchAddr, err := bchutil.NewAddressPubKeyHash(sbchAddr[:], &chaincfg.TestNet3Params)
	return bchAddr.EncodeAddress(), err
}
