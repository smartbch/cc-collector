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
		time.Sleep(5 * time.Minute)
	}
}

func handleAllPendingUTXOs(sbchClient *sbchclient.Client, bchClient *BchRpcClient) {
	fmt.Println("handleAllPendingUTXOs ...")
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
	fmt.Println("redeemingUtxos:", len(redeemingUtxos))
	if len(redeemingUtxos) > 0 {
		operatorPubkeys := getOperatorPubkeys(ccInfo.Operators)
		monitorPubkeys := getMonitorPubkeys(ccInfo.Monitors)
		currCovenant, err := ccc.NewDefaultCcCovenant(operatorPubkeys, monitorPubkeys)
		if err != nil {
			fmt.Println("failed to create CcCovenant instance:", err.Error())
			return
		}
		currCovenantAddr, _ := currCovenant.GetP2SHAddress20()
		fmt.Println("ccCovenantAddr:", hex.EncodeToString(currCovenantAddr[:]))

		oldOperatorPubkeys := getOperatorPubkeys(ccInfo.OldOperators)
		oldMonitorPubkeys := getMonitorPubkeys(ccInfo.OldMonitors)
		if len(oldOperatorPubkeys) == 0 {
			oldOperatorPubkeys = operatorPubkeys
		}
		if len(oldMonitorPubkeys) == 0 {
			oldMonitorPubkeys = monitorPubkeys
		}
		oldCovenant, err := ccc.NewDefaultCcCovenant(oldOperatorPubkeys, oldMonitorPubkeys)
		if err != nil {
			fmt.Println("failed to create old CcCovenant instance:", err.Error())
			return
		}
		oldCovenantAddr, _ := oldCovenant.GetP2SHAddress20()
		fmt.Println("old ccCovenantAddr:", hex.EncodeToString(oldCovenantAddr[:]))

		for _, utxo := range redeemingUtxos {
			if utxo.CovenantAddr == currCovenantAddr {
				handleRedeemingUTXO(bchClient, currCovenant, ccInfo.Operators, utxo)
			} else if utxo.CovenantAddr == oldCovenantAddr {
				ops := ccInfo.OldOperators
				if len(ops) == 0 {
					ops = ccInfo.Operators
				}

				handleRedeemingUTXO(bchClient, oldCovenant, ops, utxo)
			} else {
				fmt.Println("unknown covenant address:", hex.EncodeToString(utxo.CovenantAddr[:]))
			}
		}
	}

	toBeConvertedUtxos, err := getToBeConvertedUtxosForOperators(sbchClient)
	if err != nil {
		fmt.Println("failed to get toBeConverted UTXOs:", err.Error())
		return
	}
	fmt.Println("toBeConvertedUtxos:", len(toBeConvertedUtxos))
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

		oldOps := ccInfo.OldOperators
		if len(oldOps) == 0 {
			oldOps = ccInfo.Operators
		}

		ccCovenant, err := ccc.NewDefaultCcCovenant(oldOperatorPubkeys, oldMonitorPubkeys)
		if err != nil {
			fmt.Println("failed to create CcCovenant instance:", err.Error())
			return
		}

		_ccAddr, _ := ccCovenant.GetP2SHAddress()
		fmt.Println("oldCcCovenantAddr:", _ccAddr)
		for _, utxo := range toBeConvertedUtxos {
			handleToBeConvertedUTXO(bchClient, ccCovenant, oldOps,
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

	nSigs := len(sigs)
	if nSigs < minOperatorSigCount {
		fmt.Println("not enough operator sigs:", nSigs)
		return
	}

	if nSigs > minOperatorSigCount {
		sigs = sigs[:minOperatorSigCount]
	}

	_, rawTx, err := ccCovenant.FinishRedeemByUserTx(tx, sigs)
	if err != nil {
		fmt.Println("failed to sign tx:", err.Error())
		return
	}
	//fmt.Println("rawTx:", hex.EncodeToString(rawTx))

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

	nSigs := len(sigs)
	if nSigs < minOperatorSigCount {
		fmt.Println("not enough operator sigs:", nSigs)
		return
	}

	if nSigs > minOperatorSigCount {
		sigs = sigs[:minOperatorSigCount]
	}

	_, rawTx, err := oldCcCovenant.FinishConvertByOperatorsTx(tx,
		newOperatorPubkeys, newMonitorPubkeys, sigs)
	if err != nil {
		fmt.Println("failed to sign tx:", err.Error())
		return
	}
	//fmt.Println("rawTx:", hex.EncodeToString(rawTx))

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
